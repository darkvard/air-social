package usecase

import (
	"context"
	"errors"
	"fmt"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/internal/domain/follow"
	"air-social/pkg"
)

type FollowChecker interface {
	GetRelationship(ctx context.Context, userID, targetID int64) (follow.Relationship, error)
	GetRelationships(ctx context.Context, userID int64, targetIDs []int64) (map[int64]follow.Relationship, error)
}

type UserChecker interface {
	FilterNonExistent(ctx context.Context, ids []int64) (missing []int64, err error)
}

type MediaVerifier interface {
	VerifyMedia(ctx context.Context, keys []string) error
}

type RealtimePublisher interface {
	PublishNewMessage(ctx context.Context, msg *chat.Message, conv *chat.Conversation) error
	PublishMessageEdited(ctx context.Context, msg *chat.Message) error
	PublishMessageDeleted(ctx context.Context, msgID string, convID string) error
	PublishReactionAdded(ctx context.Context, msgID string, convID string, reaction chat.Reaction) error
	PublishReactionRemoved(ctx context.Context, msgID string, convID string, userID int64) error
	PublishReadEvent(ctx context.Context, convID string, userID int64, lastReadMsgID string) error
}

type WriteDeps struct {
	ConvRepo      chat.ConversationRepository
	MsgRepo       chat.MessageRepository
	Unread        chat.UnreadStore
	FollowChecker FollowChecker
	UserChecker   UserChecker
	MediaVerifier MediaVerifier
	Event         common.EventPublisher
}

type ConversationWriteUseCase struct {
	deps WriteDeps
}

func NewWriteUseCase(d WriteDeps) *ConversationWriteUseCase {
	return &ConversationWriteUseCase{deps: d}
}

func (u *ConversationWriteUseCase) CreateOrGetDirect(ctx context.Context, senderID, recipientID int64) (*chat.Conversation, bool, error) {
	existingConv, err := u.deps.ConvRepo.GetDirect(ctx, senderID, recipientID)
	if err != nil {
		return nil, false, pkg.OrInternalError(err)
	}
	if existingConv != nil {
		u.enrichConversation(ctx, existingConv, senderID)
		return existingConv, false, nil
	}

	relationship, err := u.deps.FollowChecker.GetRelationship(ctx, senderID, recipientID)
	if err != nil {
		return nil, false, pkg.OrInternalError(err)
	}

	newConv := chat.NewDirectConversation(senderID, recipientID, relationship.IsFollower)
	if err := u.deps.ConvRepo.Create(ctx, newConv); err != nil {
		return nil, false, pkg.OrInternalError(err)
	}
	return newConv, true, nil
}

func (u *ConversationWriteUseCase) CreateGroup(ctx context.Context, params chat.CreateGroupParams) (*chat.Conversation, bool, error) {
	// clean input: deduplicate members, remove creator, enforce minimum count
	memberIDs, err := u.sanitizeGroupMembers(params)
	if err != nil {
		return nil, false, err
	}

	// validate media: ensure avatar key exists in storage
	if err := u.verifyAvatar(ctx, params.AvatarKey); err != nil {
		return nil, false, err
	}

	// idempotency: return existing conversation if client already created it
	if existing, err := u.checkClientConvID(ctx, params.ClientConvID); err != nil {
		return nil, false, err
	} else if existing != nil {
		u.enrichConversation(ctx, existing, params.CreatorID)
		return existing, false, nil
	}

	// check members exist and decide who lands in active vs pending inbox
	memberStates, err := u.resolveMembers(ctx, params.CreatorID, memberIDs)
	if err != nil {
		return nil, false, err
	}

	// write conversation, recover gracefully if a concurrent request won the race
	return u.persistGroup(ctx, params, memberIDs, memberStates)
}

func (u *ConversationWriteUseCase) UpdateGroup(ctx context.Context, params chat.UpdateGroupParams) (*chat.Conversation, error) {
	// load conv: ensure it exists and is a group
	conv, err := u.deps.ConvRepo.GetByID(ctx, params.ConvID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}
	if conv == nil {
		return nil, pkg.ErrNotFound
	}
	if conv.Type != chat.ConversationGroup {
		return nil, pkg.NewError(pkg.ErrBadRequest, "only group conversations can be updated")
	}

	// authorize: only admins can update group info
	actor := conv.FindParticipant(params.ActorID)
	if actor == nil || !actor.IsAdmin() {
		return nil, pkg.ErrForbidden
	}

	// guard: at least one field must change, otherwise reject early
	if params.Name == nil && params.AvatarKey == nil {
		return nil, pkg.NewError(pkg.ErrBadRequest, "nothing to update")
	}

	// validate media: ensure new avatar key exists in storage
	if params.AvatarKey != nil {
		if err := u.verifyAvatar(ctx, *params.AvatarKey); err != nil {
			return nil, err
		}
	}

	// persist: only non-nil fields are written, existing values are preserved
	if err := u.deps.ConvRepo.UpdateGroupInfo(ctx, params.ConvID, params.Name, params.AvatarKey); err != nil {
		return nil, pkg.OrInternalError(err)
	}

	// patch in-memory to avoid a redundant re-fetch from DB
	if params.Name != nil {
		conv.Name = *params.Name
	}
	if params.AvatarKey != nil {
		conv.AvatarKey = *params.AvatarKey
	}
	return conv, nil
}

// sanitizeGroupMembers deduplicates memberIDs, removes creatorID, and enforces minimum count.
func (u *ConversationWriteUseCase) sanitizeGroupMembers(params chat.CreateGroupParams) ([]int64, error) {
	seen := make(map[int64]struct{}, len(params.MemberIDs))
	memberIDs := make([]int64, 0, len(params.MemberIDs))
	for _, id := range params.MemberIDs {
		if id == params.CreatorID {
			continue
		}
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			memberIDs = append(memberIDs, id)
		}
	}
	if len(memberIDs) < 2 {
		return nil, pkg.NewError(pkg.ErrBadRequest, "group requires at least 2 unique members excluding creator")
	}
	return memberIDs, nil
}

// resolveMembers validates that all member IDs exist, then bulk-checks their follow relationship
// with the creator to determine each member's initial inbox state.
// Members who follow the creator land in Active (main inbox); others in Pending (message requests).
func (u *ConversationWriteUseCase) resolveMembers(ctx context.Context, creatorID int64, memberIDs []int64) (map[int64]chat.ParticipantState, error) {
	missing, err := u.deps.UserChecker.FilterNonExistent(ctx, memberIDs)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}
	if len(missing) > 0 {
		return nil, pkg.NewError(pkg.ErrNotFound, fmt.Sprintf("participants not found: %v", missing))
	}

	relationships, err := u.deps.FollowChecker.GetRelationships(ctx, creatorID, memberIDs)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}
	states := make(map[int64]chat.ParticipantState, len(memberIDs))
	for _, id := range memberIDs {
		if rel, ok := relationships[id]; ok && rel.IsFollower {
			states[id] = chat.StateActive
		} else {
			states[id] = chat.StatePending
		}
	}
	return states, nil
}

// persistGroup creates the conversation document and handles the idempotency conflict race.
// Returns (conv, isNew, err): isNew=false when a concurrent request already created the same ClientConvID.
func (u *ConversationWriteUseCase) persistGroup(
	ctx context.Context,
	params chat.CreateGroupParams,
	memberIDs []int64,
	memberStates map[int64]chat.ParticipantState,
) (*chat.Conversation, bool, error) {
	conv := chat.NewGroupConversation(chat.CreateGroupParams{
		CreatorID:    params.CreatorID,
		MemberIDs:    memberIDs,
		MemberStates: memberStates,
		Name:         params.Name,
		AvatarKey:    params.AvatarKey,
		ClientConvID: params.ClientConvID,
	})

	if err := u.deps.ConvRepo.Create(ctx, conv); err != nil {
		if errors.Is(err, pkg.ErrConflict) && params.ClientConvID != "" {
			// Race condition: a concurrent request created the same ClientConvID first.
			existing, findErr := u.deps.ConvRepo.GetByClientConvID(ctx, params.ClientConvID)
			if findErr != nil {
				return nil, false, pkg.OrInternalError(findErr)
			}
			if existing != nil {
				u.enrichConversation(ctx, existing, params.CreatorID)
				return existing, false, nil
			}
		}
		return nil, false, pkg.OrInternalError(err)
	}

	return conv, true, nil
}

// enrichConversation populates LastMessage and UnreadCount on an existing conversation
// returned by write operations. Errors are swallowed — fields stay at zero values if fetches fail.
func (u *ConversationWriteUseCase) enrichConversation(ctx context.Context, conv *chat.Conversation, userID int64) {
	if conv.LastMsgID != "" {
		msgs, err := u.deps.MsgRepo.GetByIDs(ctx, []string{conv.LastMsgID})
		if err == nil && len(msgs) > 0 {
			conv.LastMessage = &msgs[0]
		}
	}
	count, err := u.deps.Unread.Get(ctx, userID, conv.ID)
	if err == nil {
		conv.UnreadCount = int(count)
	}
}

// checkClientConvID provides idempotency for group creation.
// Mobile clients generate a UUID before sending the request and attach it as ClientConvID.
// If the network drops and the client retries, this check returns the already-created
// conversation instead of creating a duplicate — even if the first request succeeded.
// Empty ClientConvID means the caller opted out of idempotency (no guarantee against duplicates).
func (u *ConversationWriteUseCase) checkClientConvID(ctx context.Context, clientConvID string) (*chat.Conversation, error) {
	if clientConvID == "" {
		return nil, nil
	}
	existing, err := u.deps.ConvRepo.GetByClientConvID(ctx, clientConvID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}
	return existing, nil
}

func (u *ConversationWriteUseCase) verifyAvatar(ctx context.Context, avatarKey string) error {
	if avatarKey == "" {
		return nil
	}
	if err := u.deps.MediaVerifier.VerifyMedia(ctx, []string{avatarKey}); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.NewError(pkg.ErrBadRequest, "avatar media not found or invalid")
		}
		return pkg.OrInternalError(err)
	}
	return nil
}

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
	PublishReadEvent(ctx context.Context, convID string, userID int64, lastReadMsgID string) error
	PublishMessageDeleted(ctx context.Context, msgID string, convID string) error
	PublishReactionAdded(ctx context.Context, msgID string, convID string, reaction chat.Reaction) error
	PublishReactionRemoved(ctx context.Context, msgID string, convID string, userID int64) error
}

type WriteDeps struct {
	ConvRepo      chat.ConversationRepository
	FollowChecker FollowChecker
	UserChecker   UserChecker
	MediaVerifier MediaVerifier
	Event         common.EventPublisher
	RTPublisher   RealtimePublisher
}

type ConversationWriteUseCase struct {
	deps WriteDeps
}

func NewWriteUseCase(d WriteDeps) *ConversationWriteUseCase {
	return &ConversationWriteUseCase{deps: d}
}

func (u *ConversationWriteUseCase) CreateOrGetDirect(ctx context.Context, senderID, recipientID int64) (*chat.Conversation, error) {
	existingConv, err := u.deps.ConvRepo.GetDirect(ctx, senderID, recipientID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}
	if existingConv != nil {
		// TODO: populate existingConv.LastMessage from MessageRepository once message layer is implemented.
		return existingConv, nil
	}

	relationship, err := u.deps.FollowChecker.GetRelationship(ctx, senderID, recipientID)
	if err != nil {
		return nil, pkg.OrInternalError(err)
	}

	newConv := chat.NewDirectConversation(senderID, recipientID, relationship.IsFollower)
	if err := u.deps.ConvRepo.Create(ctx, newConv); err != nil {
		return nil, pkg.OrInternalError(err)
	}
	return newConv, nil
}

func (u *ConversationWriteUseCase) CreateGroup(ctx context.Context, params chat.CreateGroupParams) (*chat.Conversation, error) {
	memberIDs, err := u.sanitizeGroupMembers(params)
	if err != nil {
		return nil, err
	}

	if params.AvatarKey != "" {
		if err := u.deps.MediaVerifier.VerifyMedia(ctx, []string{params.AvatarKey}); err != nil {
			if errors.Is(err, pkg.ErrNotFound) {
				return nil, pkg.NewError(pkg.ErrBadRequest, "avatar media not found or invalid")
			}
			return nil, pkg.OrInternalError(err)
		}
	}

	if params.ClientConvID != "" {
		existing, err := u.deps.ConvRepo.GetByClientConvID(ctx, params.ClientConvID)
		if err != nil {
			return nil, pkg.OrInternalError(err)
		}
		if existing != nil {
			// TODO: populate existing.LastMessage from MessageRepository once message layer is implemented.
			return existing, nil
		}
	}

	memberStates, err := u.resolveMembers(ctx, params.CreatorID, memberIDs)
	if err != nil {
		return nil, err
	}

	return u.persistGroup(ctx, params, memberIDs, memberStates)
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
func (u *ConversationWriteUseCase) persistGroup(
	ctx context.Context,
	params chat.CreateGroupParams,
	memberIDs []int64,
	memberStates map[int64]chat.ParticipantState,
) (*chat.Conversation, error) {
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
				return nil, pkg.OrInternalError(findErr)
			}
			if existing != nil {
				return existing, nil
			}
		}
		return nil, pkg.OrInternalError(err)
	}

	return conv, nil
}

func (u *ConversationWriteUseCase) UpdateGroup(ctx context.Context, params chat.UpdateGroupParams) error {
	return nil
}

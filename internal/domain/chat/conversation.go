package chat

import (
	"time"

	"air-social/pkg"
)

type ConversationType string

const (
	ConversationDirect ConversationType = "direct"
	ConversationGroup  ConversationType = "group"
)

var TimeFormat = time.RFC3339Nano

type Conversation struct {
	ID           string
	Type         ConversationType
	Participants []Participant
	Name         string // group only
	AvatarKey    string // group only
	ClientConvID string // group only; optional idempotency key set by client on creation
	CreatedBy    int64
	LastMessage  *Message // populated by repo (separate query); nil if no messages yet
	UnreadCount  int      // populated from Redis UnreadStore; not stored in MongoDB
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (c Conversation) GetCursor() string {
	return c.UpdatedAt.Format(TimeFormat)
}

func (c Conversation) FindParticipant(userID int64) *Participant {
	for i := range c.Participants {
		if c.Participants[i].UserID == userID {
			return &c.Participants[i]
		}
	}
	return nil
}

func NewDirectConversation(senderID, targetID int64, isTargetFollowingSender bool) *Conversation {
	targetState := StatePending
	if isTargetFollowingSender {
		targetState = StateActive
	}

	now := pkg.TimeNowUTC()
	return &Conversation{
		ID:   pkg.NewULID(),
		Type: ConversationDirect,
		Participants: []Participant{
			{
				UserID:   senderID,
				Role:     RoleMember,
				State:    StateActive,
				JoinedAt: now,
			},
			{
				UserID:   targetID,
				Role:     RoleMember,
				State:    targetState,
				JoinedAt: now,
			},
		},
		CreatedBy: senderID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewGroupConversation(params CreateGroupParams) *Conversation {
	now := pkg.TimeNowUTC()
	participants := make([]Participant, 0, 1+len(params.MemberIDs))
	participants = append(participants, Participant{
		UserID:   params.CreatorID,
		Role:     RoleAdmin,
		State:    StateActive,
		JoinedAt: now,
	})
	for _, id := range params.MemberIDs {
		state := StatePending // default: no follow relationship with creator → pending
		if s, ok := params.MemberStates[id]; ok {
			state = s
		}
		participants = append(participants, Participant{
			UserID:   id,
			Role:     RoleMember,
			State:    state,
			JoinedAt: now,
		})
	}
	return &Conversation{
		ID:           pkg.NewULID(),
		Type:         ConversationGroup,
		Participants: participants,
		Name:         params.Name,
		AvatarKey:    params.AvatarKey,
		ClientConvID: params.ClientConvID,
		CreatedBy:    params.CreatorID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

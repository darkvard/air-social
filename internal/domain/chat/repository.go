package chat

import "context"

type ConversationRepository interface {
	Create(ctx context.Context, conv *Conversation) error
	GetByID(ctx context.Context, id string) (*Conversation, error)
	FindDirect(ctx context.Context, userAID, userBID int64) (*Conversation, error)
	GetParticipantConversations(ctx context.Context, params GetConversationsParams) ([]Conversation, error)

	AddParticipant(ctx context.Context, convID string, p Participant) error
	RemoveParticipant(ctx context.Context, convID string, userID int64) error
	UpdateParticipantState(ctx context.Context, convID string, userID int64, state ParticipantState) error
	UpdateParticipantRole(ctx context.Context, convID string, userID int64, role ParticipantRole) error

	// UpdateGroupInfo only updates non-nil fields — nil means keep existing value.
	UpdateGroupInfo(ctx context.Context, convID string, name, avatarKey *string) error

	UpdateLastRead(ctx context.Context, convID string, userID int64, messageID string) error
	TouchConversation(ctx context.Context, convID, lastMsgID string) error
}

type MessageRepository interface {
	Create(ctx context.Context, msg *Message) error
	GetByID(ctx context.Context, id string) (*Message, error)
	GetConversationMessages(ctx context.Context, params GetMessagesParams) ([]Message, error)

	// GetByClientMsgID is used for idempotency: when Create returns ErrConflict,
	// the use case calls this to return the already-saved message instead of erroring.
	GetByClientMsgID(ctx context.Context, clientMsgID string) (*Message, error)
	Update(ctx context.Context, msg *Message) error
	SoftDelete(ctx context.Context, id string, userID int64) error

	AddReaction(ctx context.Context, msgID string, r Reaction) error
	RemoveReaction(ctx context.Context, msgID string, userID int64) error
}

type PresenceStore interface {
	SetOnline(ctx context.Context, userID int64) error
	SetOffline(ctx context.Context, userID int64) error
	RefreshPresence(ctx context.Context, userID int64) error

	IsOnline(ctx context.Context, userID int64) (bool, error)
}

type UnreadStore interface {
	Increment(ctx context.Context, userID int64, convID string) error
	Reset(ctx context.Context, userID int64, convID string) error

	Get(ctx context.Context, userID int64, convID string) (int64, error)
	GetAll(ctx context.Context, userID int64) (map[string]int64, error)
}

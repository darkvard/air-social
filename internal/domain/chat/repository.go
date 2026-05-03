package chat

import "context"

type ConversationRepository interface {
	Create(ctx context.Context, conv *Conversation) error
	UpdateGroupInfo(ctx context.Context, convID string, name, avatarKey *string) error
	AddParticipant(ctx context.Context, convID string, p Participant) error
	RemoveParticipant(ctx context.Context, convID string, userID int64) error
	UpdateParticipantState(ctx context.Context, convID string, userID int64, state ParticipantState) error
	UpdateParticipantRole(ctx context.Context, convID string, userID int64, role ParticipantRole) error
	UpdateLastRead(ctx context.Context, convID string, userID int64, messageID string) error
	Touch(ctx context.Context, convID, lastMsgID string) error

	GetList(ctx context.Context, params GetConversationsParams) ([]Conversation, error)
	GetDirect(ctx context.Context, userAID, userBID int64) (*Conversation, error)
	GetByID(ctx context.Context, id string) (*Conversation, error)
	GetByClientConvID(ctx context.Context, clientConvID string) (*Conversation, error)
}

type MessageRepository interface {
	Create(ctx context.Context, msg *Message) error
	Update(ctx context.Context, msg *Message) error
	SoftDelete(ctx context.Context, id string) error
	AddReaction(ctx context.Context, msgID string, r *Reaction) error
	RemoveReaction(ctx context.Context, msgID string, userID int64) error

	GetByID(ctx context.Context, id string) (*Message, error)
	GetByIDs(ctx context.Context, ids []string) ([]Message, error)
	GetList(ctx context.Context, params GetMessagesParams) ([]Message, error)
	GetByClientMsgID(ctx context.Context, clientMsgID string, senderID int64) (*Message, error)
}

type PresenceStore interface {
	SetOnline(ctx context.Context, userID int64) error
	IsOnlineBatch(ctx context.Context, userIDs []int64) (map[int64]bool, error)
}

type UnreadStore interface {
	Increment(ctx context.Context, userID int64, convID string) error
	Reset(ctx context.Context, userID int64, convID string) error
	Get(ctx context.Context, userID int64, convID string) (int64, error)
	GetAll(ctx context.Context, userID int64) (map[string]int64, error)
}

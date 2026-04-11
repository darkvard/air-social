package chat

import (
	"time"

	"air-social/internal/domain/common"
)

type GetConversationsParams struct {
	UserID int64
	State  ParticipantState
	Query  common.CursorQueryParams[string]
}

func (p GetConversationsParams) GetTimeFromCursor() (time.Time, error) {
	return time.Parse(TimeFormat, p.Query.Cursor)
}

type CreateGroupParams struct {
	CreatorID    int64
	MemberIDs    []int64
	MemberStates map[int64]ParticipantState // resolved by usecase; missing key defaults to StatePending
	Name         string
	AvatarKey    string
	ClientConvID string // optional; empty = no idempotency guarantee
}

type UpdateGroupParams struct {
	ConvID    string
	ActorID   int64
	Name      *string
	AvatarKey *string
}

type GetMessagesParams struct {
	ConversationID string
	UserID         int64
	Query          common.CursorQueryParams[string]
}

type SendMessageParams struct {
	ConversationID string
	SenderID       int64
	Type           MessageType
	Content        string
	ReplyToID      *string
	ClientMsgID    string // client-generated (UUID/nanoid); empty = no dedup
}

type EditMessageParams struct {
	MessageID string
	UserID    int64
	Content   string
}

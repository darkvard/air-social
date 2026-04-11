package chat

import "time"

type MessageType string

const (
	MessageText   MessageType = "text"
	MessageImage  MessageType = "image"
	MessageVideo  MessageType = "video"
	MessageAudio  MessageType = "audio"
	MessageFile   MessageType = "file"
	MessageSystem MessageType = "system"
)

type Message struct {
	ID             string
	ConversationID string
	SenderID       int64
	Type           MessageType
	Content        string
	ClientMsgID    string  // client-generated idempotency key; dedup on retry
	ReplyToID      *string // nil = not a reply
	Reactions      []Reaction
	IsDeleted      bool // soft delete; content cleared but row kept for thread context
	IsEdited       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (m Message) GetCursor() string { return m.ID }

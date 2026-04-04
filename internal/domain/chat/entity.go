package chat

import "time"

type ConversationType string

const (
	ConversationDirect ConversationType = "direct"
	ConversationGroup  ConversationType = "group"
)

type ParticipantState string

const (
	StateActive  ParticipantState = "active"
	StatePending ParticipantState = "pending"
	StateIgnored ParticipantState = "ignored"
	StateMuted   ParticipantState = "muted"
)

type ParticipantRole string

const (
	RoleAdmin  ParticipantRole = "admin"
	RoleMember ParticipantRole = "member"
)

type MessageType string

const (
	MessageText   MessageType = "text"
	MessageImage  MessageType = "image"
	MessageVideo  MessageType = "video"
	MessageAudio  MessageType = "audio"
	MessageFile   MessageType = "file"
	MessageSystem MessageType = "system"
)

type Participant struct {
	UserID     int64
	Role       ParticipantRole  // admin | member; direct conv: both member
	State      ParticipantState // active | pending | ignored | muted
	JoinAt     time.Time
	LastReadID string // ULID of last read message; "" = never read
}

func (p Participant) HasReadAnyMessage() bool {
	return p.LastReadID != ""
}

type Conversation struct {
	ID           string // ULID — time-sortable, used as cursor
	Type         ConversationType
	Participants []Participant
	Name         string // group only
	AvatarKey    string // group only
	CreatedBy    int64
	LastMessage  *Message // populated by repo (separate query); nil if no messages yet
	UnreadCount  int      // populated from Redis UnreadStore; not stored in MongoDB
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (c Conversation) GetCursor() string { return c.ID }

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

type ReactionType string

const (
	ReactionLike  ReactionType = "👍"
	ReactionLove  ReactionType = "❤️"
	ReactionHaha  ReactionType = "😂"
	ReactionWow   ReactionType = "😮"
	ReactionSad   ReactionType = "😢"
	ReactionAngry ReactionType = "😠"
)

var allowedReactions = map[ReactionType]bool{
	ReactionLike: true, ReactionLove: true, ReactionHaha: true,
	ReactionWow: true, ReactionSad: true, ReactionAngry: true,
}

func (r ReactionType) IsValid() bool {
	return allowedReactions[r]
}

type Reaction struct {
	UserID    int64
	Type      ReactionType
	CreatedAt time.Time
}

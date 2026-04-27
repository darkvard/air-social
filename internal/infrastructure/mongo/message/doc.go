package message

import (
	"time"

	"air-social/internal/domain/chat"
)

const (
	fieldID             = "_id"
	fieldConversationID = "conversation_id"
	fieldSenderID       = "sender_id"
	fieldType           = "type"
	fieldContent        = "content"
	fieldClientMsgID    = "client_msg_id"
	fieldReplyToID      = "reply_to_id"
	fieldReactions      = "reactions"
	fieldUserID         = "user_id"
	fieldIsDeleted      = "is_deleted"
	fieldIsEdited       = "is_edited"
	fieldCreatedAt      = "created_at"
	fieldUpdatedAt      = "updated_at"
)

const (
	fieldReactionUserID = fieldReactions + "." + fieldUserID
)

type reactionDoc struct {
	UserID    int64     `bson:"user_id"`
	Type      string    `bson:"type"`
	CreatedAt time.Time `bson:"created_at"`
}

type messageDoc struct {
	ID             string        `bson:"_id"`
	ConversationID string        `bson:"conversation_id"`
	SenderID       int64         `bson:"sender_id"`
	Type           string        `bson:"type"`
	Content        string        `bson:"content"`
	ClientMsgID    string        `bson:"client_msg_id,omitempty"`
	ReplyToID      *string       `bson:"reply_to_id,omitempty"`
	Reactions      []reactionDoc `bson:"reactions"`
	IsDeleted      bool          `bson:"is_deleted"`
	IsEdited       bool          `bson:"is_edited"`
	CreatedAt      time.Time     `bson:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at"`
}

func (d *messageDoc) toDomain() *chat.Message {
	reactions := make([]chat.Reaction, len(d.Reactions))
	for i, r := range d.Reactions {
		reactions[i] = chat.Reaction{
			UserID:    r.UserID,
			Type:      chat.ReactionType(r.Type),
			CreatedAt: r.CreatedAt,
		}
	}
	return &chat.Message{
		ID:             d.ID,
		ConversationID: d.ConversationID,
		SenderID:       d.SenderID,
		Type:           chat.MessageType(d.Type),
		Content:        d.Content,
		ClientMsgID:    d.ClientMsgID,
		ReplyToID:      d.ReplyToID,
		Reactions:      reactions,
		IsDeleted:      d.IsDeleted,
		IsEdited:       d.IsEdited,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

func messageFromDomain(m *chat.Message) *messageDoc {
	reactions := make([]reactionDoc, len(m.Reactions))
	for i, r := range m.Reactions {
		reactions[i] = reactionDoc{
			UserID:    r.UserID,
			Type:      string(r.Type),
			CreatedAt: r.CreatedAt,
		}
	}
	return &messageDoc{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Type:           string(m.Type),
		Content:        m.Content,
		ClientMsgID:    m.ClientMsgID,
		ReplyToID:      m.ReplyToID,
		Reactions:      reactions,
		IsDeleted:      m.IsDeleted,
		IsEdited:       m.IsEdited,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func reactionFromDomain(r *chat.Reaction) reactionDoc {
	return reactionDoc{
		UserID:    r.UserID,
		Type:      string(r.Type),
		CreatedAt: r.CreatedAt,
	}
}

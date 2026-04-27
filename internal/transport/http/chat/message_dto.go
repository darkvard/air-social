package chat

import (
	"time"

	chatdomain "air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/internal/transport/http/shared"
)

type MessagePathParam struct {
	ConvID string `uri:"id"`
	MsgID  string `uri:"msgID"`
}

type SendMessageReq struct {
	Type        string  `json:"type" binding:"required,oneof=text image video audio file"`
	Content     string  `json:"content" binding:"required"`
	ReplyToID   *string `json:"reply_to_id,omitempty"`
	ClientMsgID string  `json:"client_msg_id,omitempty"`
}

func (r SendMessageReq) ToDomain(convID string, senderID int64) chatdomain.SendMessageParams {
	return chatdomain.SendMessageParams{
		ConversationID: convID,
		SenderID:       senderID,
		Type:           chatdomain.MessageType(r.Type),
		Content:        r.Content,
		ReplyToID:      r.ReplyToID,
		ClientMsgID:    r.ClientMsgID,
	}
}

type GetMessagesReq struct {
	Cursor string `form:"cursor" binding:"omitempty"`
	Limit  int    `form:"limit,default=20" binding:"omitempty,min=1,max=50"`
}

func (r GetMessagesReq) ToDomain(convID string, userID int64) chatdomain.GetMessagesParams {
	return chatdomain.GetMessagesParams{
		ConversationID: convID,
		UserID:         userID,
		Query: common.CursorQueryParams[string]{
			Cursor: r.Cursor,
			Limit:  r.Limit,
		},
	}
}

type EditMessageReq struct {
	Content string `json:"content" binding:"required"`
}

type ReactionReq struct {
	Type string `json:"type" binding:"required,oneof=like love haha wow sad angry"`
}

type MessagesRes = shared.CursorPaginatedResponse[MessageRes, string]

type ReactionRes struct {
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type MessageRes struct {
	ID             string        `json:"id"`
	ConversationID string        `json:"conversation_id"`
	SenderID       int64         `json:"sender_id"`
	Type           string        `json:"type"`
	Content        string        `json:"content"`
	ReplyToID      *string       `json:"reply_to_id,omitempty"`
	Reactions      []ReactionRes `json:"reactions"`
	IsDeleted      bool          `json:"is_deleted"`
	IsEdited       bool          `json:"is_edited"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

func toMessageRes(m *chatdomain.Message) MessageRes {
	reactions := make([]ReactionRes, len(m.Reactions))
	for i, r := range m.Reactions {
		reactions[i] = ReactionRes{
			UserID:    r.UserID,
			Type:      string(r.Type),
			CreatedAt: r.CreatedAt,
		}
	}
	return MessageRes{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Type:           string(m.Type),
		Content:        m.Content,
		ReplyToID:      m.ReplyToID,
		Reactions:      reactions,
		IsDeleted:      m.IsDeleted,
		IsEdited:       m.IsEdited,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

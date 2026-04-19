package chat

import (
	"time"

	chatdomain "air-social/internal/domain/chat"
	"air-social/internal/domain/common"
	"air-social/internal/transport/http/shared"
)

type PathIDParam = shared.PathStringIDParam

type GetConversationsReq struct {
	State  string `form:"state,default=active" binding:"omitempty,oneof=active pending ignored muted"`
	Cursor string `form:"cursor" binding:"omitempty"`
	Limit  int    `form:"limit,default=10" binding:"omitempty,min=1,max=50"`
}

func (q GetConversationsReq) ToDomain(userID int64) chatdomain.GetConversationsParams {
	return chatdomain.GetConversationsParams{
		UserID: userID,
		State:  chatdomain.ParticipantState(q.State),
		Query: common.CursorQueryParams[string]{
			Cursor: q.Cursor,
			Limit:  q.Limit,
		},
	}
}

type CreateDirectReq struct {
	TargetUserID int64 `json:"target_user_id" binding:"required,gt=0"`
}

type CreateGroupReq struct {
	Name           string  `json:"name" binding:"required,max=100"`
	ParticipantIDs []int64 `json:"participant_ids" binding:"required,min=2"`
	AvatarKey      string  `json:"avatar_key,omitempty"`
	ClientConvID   string  `json:"client_conv_id,omitempty"`
}

type UpdateGroupReq struct {
	Name      *string `json:"name,omitempty" binding:"omitempty,max=100"`
	AvatarKey *string `json:"avatar_key,omitempty"`
}

type ConversationsRes = shared.CursorPaginatedResponse[ConversationRes, string]

type ParticipantRes struct {
	UserID     int64     `json:"user_id"`
	Role       string    `json:"role"`
	State      string    `json:"state"`
	JoinedAt   time.Time `json:"joined_at"`
	LastReadID string    `json:"last_read_id,omitempty"`
}

type ConversationRes struct {
	ID           string           `json:"id"`
	Type         string           `json:"type"`
	Name         string           `json:"name,omitempty"`
	AvatarKey    string           `json:"avatar_key,omitempty"`
	CreatedBy    int64            `json:"created_by"`
	Participants []ParticipantRes `json:"participants"`
	UnreadCount  int              `json:"unread_count,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
}

func toConversationResponse(conv *chatdomain.Conversation) ConversationRes {
	participants := make([]ParticipantRes, len(conv.Participants))
	for i, p := range conv.Participants {
		participants[i] = ParticipantRes{
			UserID:     p.UserID,
			Role:       string(p.Role),
			State:      string(p.State),
			JoinedAt:   p.JoinedAt,
			LastReadID: p.LastReadID,
		}
	}

	return ConversationRes{
		ID:           conv.ID,
		Type:         string(conv.Type),
		Name:         conv.Name,
		AvatarKey:    conv.AvatarKey,
		CreatedBy:    conv.CreatedBy,
		Participants: participants,
		UnreadCount:  conv.UnreadCount,
		CreatedAt:    conv.CreatedAt,
		UpdatedAt:    conv.UpdatedAt,
	}
}

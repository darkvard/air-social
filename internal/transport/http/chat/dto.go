package chat

import (
	"time"

	chatdomain "air-social/internal/domain/chat"
)

type CreateDirectReq struct {
	TargetUserID int64 `json:"target_user_id" binding:"required,gt=0"`
}

type CreateGroupReq struct {
	Name           string  `json:"name" binding:"required,max=100"`
	ParticipantIDs []int64 `json:"participant_ids" binding:"required,min=2"`
	AvatarKey      string  `json:"avatar_key,omitempty"`
}

type ParticipantResponse struct {
	UserID     int64     `json:"user_id"`
	Role       string    `json:"role"`
	State      string    `json:"state"`
	JoinedAt   time.Time `json:"joined_at"`
	LastReadID string    `json:"last_read_id,omitempty"`
}

type ConversationResponse struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`
	Name         string                `json:"name,omitempty"`
	AvatarKey    string                `json:"avatar_key,omitempty"`
	CreatedBy    int64                 `json:"created_by"`
	Participants []ParticipantResponse `json:"participants"`
	UnreadCount  int                   `json:"unread_count,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

func toConversationResponse(conv *chatdomain.Conversation) ConversationResponse {
	participants := make([]ParticipantResponse, len(conv.Participants))
	for i, p := range conv.Participants {
		participants[i] = ParticipantResponse{
			UserID:     p.UserID,
			Role:       string(p.Role),
			State:      string(p.State),
			JoinedAt:   p.JoinedAt,
			LastReadID: p.LastReadID,
		}
	}

	return ConversationResponse{
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

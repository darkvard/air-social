package notification

import (
	"time"

	"air-social/internal/domain/common"
	notif "air-social/internal/domain/notification"
	"air-social/internal/transport/http/shared"
)

type CursorQueryParams struct {
	Cursor int64 `form:"cursor" binding:"omitempty,min=0"`
	Limit  int   `form:"limit,default=20" binding:"omitempty,min=1,max=50"`
}

func (q CursorQueryParams) toDomain(userID int64) notif.GetParams {
	return notif.GetParams{
		UserID: userID,
		Query: common.CursorQueryParams[int64]{
			Cursor: q.Cursor,
			Limit:  q.Limit,
		},
	}
}

type PathIDParam = shared.PathIDParam
type MetaCursor = shared.MetaCursor[int64]
type CursorPaginatedResponse[T any] = shared.CursorPaginatedResponse[T, int64]

type NotificationResponse struct {
	ID         int64          `json:"id"`
	Type       string         `json:"type"`
	ActorID    int64          `json:"actor_id"`
	TargetID   *int64         `json:"target_id,omitempty"`
	TargetType string         `json:"target_type,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
	Read       bool           `json:"read"`
	ReadAt     *time.Time     `json:"read_at,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

func toNotificationResponse(n *notif.Notification) NotificationResponse {
	return NotificationResponse{
		ID:         n.ID,
		Type:       string(n.Type),
		ActorID:    n.ActorID,
		TargetID:   n.TargetID,
		TargetType: n.TargetType,
		Data:       n.Data,
		Read:       n.Read,
		ReadAt:     n.ReadAt,
		CreatedAt:  n.CreatedAt,
	}
}

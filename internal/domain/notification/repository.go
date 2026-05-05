package notification

import "context"

type Repository interface {
	Upsert(ctx context.Context, n *Notification) error
	Delete(ctx context.Context, userID, actorID int64, typ NotificationType, targetID *int64) error
	GetList(ctx context.Context, params GetParams) ([]Notification, error)
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	MarkRead(ctx context.Context, userID, id int64) error
	MarkAllRead(ctx context.Context, userID int64) error
}

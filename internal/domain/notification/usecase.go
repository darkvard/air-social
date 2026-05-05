package notification

import (
	"context"

	"air-social/internal/domain/chat"
	"air-social/internal/domain/common"
)
 
type NotificationPusher interface {
	PushNotification(ctx context.Context, userID int64, data any) error
}

type UseCase interface {
	Create(ctx context.Context, n *Notification) error
	Delete(ctx context.Context, userID, actorID int64, typ NotificationType, targetID *int64) error
	GetNotifications(ctx context.Context, params GetParams) (common.CursorPaginatedResult[Notification, int64], error)
	GetUnreadCount(ctx context.Context, userID int64) (int64, error)
	MarkRead(ctx context.Context, userID int64, id int64) error
	MarkAllRead(ctx context.Context, userID int64) error
}

type Deps struct {
	Repo     Repository
	Presence chat.PresenceStore
	Pusher   NotificationPusher
}

type usecase struct {
	deps Deps
}

func NewUseCase(deps Deps) UseCase {
	return &usecase{deps: deps}
}

func (u *usecase) Create(ctx context.Context, n *Notification) error {
	if err := u.deps.Repo.Upsert(ctx, n); err != nil {
		return err
	}
	if statusMap, err := u.deps.Presence.IsOnlineBatch(ctx, []int64{n.UserID}); err == nil && statusMap[n.UserID] {
		_ = u.deps.Pusher.PushNotification(ctx, n.UserID, n)
	}
	return nil
}

func (u *usecase) Delete(ctx context.Context, userID, actorID int64, typ NotificationType, targetID *int64) error {
	return u.deps.Repo.Delete(ctx, userID, actorID, typ, targetID)
}

func (u *usecase) GetNotifications(ctx context.Context, params GetParams) (common.CursorPaginatedResult[Notification, int64], error) {
	params.Query.NormalizePagination()
	items, err := u.deps.Repo.GetList(ctx, params)
	if err != nil {
		return common.CursorPaginatedResult[Notification, int64]{}, err
	}
	return common.NewCursorPaginatedResult(items, params.Query.Limit), nil
}

func (u *usecase) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	return u.deps.Repo.GetUnreadCount(ctx, userID)
}

func (u *usecase) MarkRead(ctx context.Context, userID int64, id int64) error {
	return u.deps.Repo.MarkRead(ctx, userID, id)
}

func (u *usecase) MarkAllRead(ctx context.Context, userID int64) error {
	return u.deps.Repo.MarkAllRead(ctx, userID)
}

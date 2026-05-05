package notification

import (
	"github.com/gin-gonic/gin"

	notif "air-social/internal/domain/notification"
	"air-social/internal/transport/http/middleware"
	"air-social/internal/transport/http/shared"
	"air-social/pkg"
)

type Handler struct {
	usecase notif.UseCase
}

func NewHandler(usecase notif.UseCase) Handler {
	return Handler{usecase: usecase}
}

// GetNotifications godoc
//
//	@Summary	List notifications
//	@Tags		Notifications
//	@Security	BearerAuth
//	@Param		cursor	query		int	false	"Cursor (last notification ID)"
//	@Param		limit	query		int	false	"Limit (default 20, max 50)"
//	@Success	200		{object}	CursorPaginatedResponse[NotificationResponse]
//	@Router		/notifications [get]
func (h Handler) GetNotifications(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var q CursorQueryParams
	if err := c.ShouldBindQuery(&q); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.usecase.GetNotifications(c.Request.Context(), q.toDomain(claims.UserID))
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	items := make([]NotificationResponse, len(result.Data))
	for i := range result.Data {
		items[i] = toNotificationResponse(&result.Data[i])
	}

	pkg.Success(c, CursorPaginatedResponse[NotificationResponse]{
		Data: items,
		Meta: shared.MetaCursor[int64]{
			NextCursor:  result.NextCursor,
			HasNextPage: result.HasNextPage,
		},
	})
}

// GetUnreadCount godoc
//
//	@Summary	Get unread notification count
//	@Tags		Notifications
//	@Security	BearerAuth
//	@Success	200	{object}	UnreadCountResponse
//	@Router		/notifications/unread-count [get]
func (h Handler) GetUnreadCount(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	count, err := h.usecase.GetUnreadCount(c.Request.Context(), claims.UserID)
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, UnreadCountResponse{Count: count})
}

// MarkAllRead godoc
//
//	@Summary	Mark all notifications as read
//	@Tags		Notifications
//	@Security	BearerAuth
//	@Success	204
//	@Router		/notifications/read [put]
func (h Handler) MarkAllRead(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	if err := h.usecase.MarkAllRead(c.Request.Context(), claims.UserID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

// MarkOneRead godoc
//
//	@Summary	Mark a notification as read
//	@Tags		Notifications
//	@Security	BearerAuth
//	@Param		id	path	int	true	"Notification ID"
//	@Success	204
//	@Router		/notifications/{id}/read [put]
func (h Handler) MarkOneRead(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var path PathIDParam
	if err := c.ShouldBindUri(&path); err != nil {
		pkg.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.usecase.MarkRead(c.Request.Context(), claims.UserID, path.ID); err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.NoContent(c)
}

package feed

import (
	"github.com/gin-gonic/gin"

	"air-social/internal/domain/common"
	"air-social/internal/domain/feed"
	"air-social/internal/transport/http/middleware"
	"air-social/pkg"
)

// Handler handles HTTP requests for the newsfeed.
type Handler struct {
	provider common.LinkProvider
	usecase  feed.UseCase
}

// NewHandler creates a new feed Handler.
func NewHandler(provider common.LinkProvider, usecase feed.UseCase) Handler {
	return Handler{
		provider: provider,
		usecase:  usecase,
	}
}

// GetNewsfeed godoc
//
//	@Summary		Get newsfeed
//	@Description	Returns a cursor-paginated list of posts from users the authenticated user follows.
//	@Description	Posts are ordered from newest to oldest based on their creation timestamp.
//	@Description	On the first request, omit the cursor (or set it to 0) to get the latest posts.
//	@Description	To paginate, pass the `next_cursor` from the previous response as the `cursor` param.
//	@Tags			Feed
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			cursor	query		int	false	"Cursor for pagination (Unix millisecond timestamp of the last post in previous page). Omit or set 0 for first page."
//	@Param			limit	query		int	false	"Number of posts to return (default: 20, max: 50)"
//	@Success		200		{object}	NewsfeedResponse
//	@Failure		401		{object}	pkg.Response
//	@Failure		500		{object}	pkg.Response
//	@Router			/feed [get]
func (h Handler) GetNewsfeed(c *gin.Context) {
	claims, err := middleware.GetTokenClaims(c)
	if err != nil {
		pkg.Unauthorized(c, "unauthorized")
		return
	}

	var req FeedQueryParams
	if err := c.ShouldBindQuery(&req); err != nil {
		pkg.HandleValidateError(c, err)
		return
	}

	result, err := h.usecase.Query.GetNewsfeed(c.Request.Context(), claims.UserID, req.toDomain())
	if err != nil {
		pkg.HandleServiceError(c, err)
		return
	}

	pkg.Success(c, toFeedResponse(result, h.provider.PublicFile))
}

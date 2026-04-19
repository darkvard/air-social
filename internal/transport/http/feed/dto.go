package feed

import (
	"air-social/internal/domain/common"
	"air-social/internal/domain/post"
	posthttp "air-social/internal/transport/http/post"
	"air-social/internal/transport/http/shared"
)

// NewsfeedResponse is the paginated response envelope for GET /feed.
// Defined as a concrete struct (not a generic alias) so swaggo can generate docs correctly.
type NewsfeedResponse struct {
	Data []posthttp.PostResponse  `json:"data"`
	Meta shared.MetaCursor[int64] `json:"meta"`
}

// FeedQueryParams holds query parameters for the newsfeed endpoint.
// Cursor is a Unix millisecond timestamp: 0 means "fetch from the newest post".
// No sort field — feed is always newest-first.
type FeedQueryParams struct {
	Cursor int64 `form:"cursor" binding:"omitempty,min=0"`
	Limit  int   `form:"limit,default=20" binding:"omitempty,min=1,max=50"`
}

func (q FeedQueryParams) toDomain() common.CursorQueryParams[int64] {
	return common.CursorQueryParams[int64]{
		Cursor: q.Cursor,
		Limit:  q.Limit,
	}
}

func toFeedResponse(
	result common.CursorPaginatedResult[*post.Post, int64],
	resolveURL func(string) string,
) posthttp.CursorPaginatedResponse[posthttp.PostResponse] {
	data := make([]posthttp.PostResponse, len(result.Data))
	for i, p := range result.Data {
		data[i] = posthttp.ToPostResponse(p, resolveURL)
	}
	return posthttp.CursorPaginatedResponse[posthttp.PostResponse]{
		Data: data,
		Meta: posthttp.MetaCursor{
			NextCursor:  result.NextCursor,
			HasNextPage: result.HasNextPage,
		},
	}
}

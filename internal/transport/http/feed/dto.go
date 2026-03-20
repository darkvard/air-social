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
	Data []posthttp.PostResponse `json:"data"`
	Meta shared.MetaCursor       `json:"meta"`
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

// toFeedResponse maps the domain result to the shared post HTTP response types.
// Struct types are reused from the post transport package; mapping logic lives here
// because it requires a URL resolver.
func toFeedResponse(
	result common.CursorPaginatedResult[*post.Post, int64],
	resolveURL func(string) string,
) posthttp.CursorPaginatedResponse[posthttp.PostResponse] {
	data := make([]posthttp.PostResponse, len(result.Data))
	for i, p := range result.Data {
		data[i] = toPostResponse(p, resolveURL)
	}
	return posthttp.CursorPaginatedResponse[posthttp.PostResponse]{
		Data: data,
		Meta: posthttp.MetaCursor{
			NextCursor:  result.NextCursor,
			HasNextPage: result.HasNextPage,
		},
	}
}

func toPostResponse(p *post.Post, resolveURL func(string) string) posthttp.PostResponse {
	if p == nil {
		return posthttp.PostResponse{}
	}
	return posthttp.PostResponse{
		ID:             p.ID,
		OriginalPostID: p.OriginalPostID,
		Content:        p.Content,
		Visibility:     string(p.Visibility),
		LikesCount:     p.Stat.LikesCount,
		CommentsCount:  p.Stat.CommentsCount,
		SharesCount:    p.Stat.SharesCount,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
		Media:          toMediaItemResponses(p.Media, resolveURL),
		IsLiked:        p.IsLiked,
		User:           toAuthorResponse(p.Author, resolveURL),
	}
}

func toAuthorResponse(a *post.Author, resolveURL func(string) string) *posthttp.UserResponse {
	if a == nil {
		return nil
	}
	return &posthttp.UserResponse{
		ID:         a.ID,
		Fullname:   a.FullName,
		Avatar:     resolveURL(a.Avatar),
		IsVerified: a.IsVerified,
	}
}

func toMediaItemResponses(media []post.Media, resolveURL func(string) string) []posthttp.MediaItemResponse {
	result := make([]posthttp.MediaItemResponse, len(media))
	for i, m := range media {
		result[i] = posthttp.MediaItemResponse{
			ID:        m.ID,
			URL:       resolveURL(m.MediaKey),
			MediaType: m.MediaType,
			Width:     m.Metadata.Width,
			Height:    m.Metadata.Height,
			Duration:  m.Metadata.Duration,
			FileName:  m.Metadata.FileName,
		}
	}
	return result
}

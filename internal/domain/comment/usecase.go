package comment

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"

	"air-social/internal/domain/common"
	"air-social/internal/domain/follow"
	"air-social/internal/domain/post"
	"air-social/internal/domain/stats"
	"air-social/pkg"
)

type UseCase interface {
	GetComments(ctx context.Context, postID int64, params GetCursorParams) (common.CursorPaginatedResult[Comment, int64], error)

	GetReplies(ctx context.Context, parentID int64, params GetCursorParams) (common.CursorPaginatedResult[Comment, int64], error)

	CreateComment(ctx context.Context, params CreateParams) (*Comment, error)

	UpdateComment(ctx context.Context, params UpdateParams) (*Comment, error)

	DeleteComment(ctx context.Context, commentID, userID int64) error
}

type LikeChecker interface {
	IsCommentLiked(ctx context.Context, commentIDs []int64, userID int64) (map[int64]bool, error)
}
type StatsFetcher interface {
	GetCommentsStats(ctx context.Context, commentIDs []int64) (map[int64]stats.CommentStats, error)
}

type PostFetcher interface {
	GetPostDetail(ctx context.Context, postID, viewerID int64) (*post.Post, error)
}

type FollowChecker interface {
	GetRelationship(ctx context.Context, userID, targetID int64) (follow.Relationship, error)
}

type MediaVerifier interface {
	VerifyMedia(ctx context.Context, keys []string) error
}

type Deps struct {
	CommentRepo   Repository
	PostFetcher   PostFetcher
	FollowChecker FollowChecker
	MediaVerifier MediaVerifier
	LikeChecker   LikeChecker
	StatsFetcher  StatsFetcher
	Event         common.EventPublisher
}

type usecase struct {
	commentRepo   Repository
	postFetcher   PostFetcher
	followChecker FollowChecker
	mediaVerifier MediaVerifier
	likeChecker   LikeChecker
	statsFetcher  StatsFetcher
	event         common.EventPublisher
}

func NewUseCase(deps Deps) *usecase {
	return &usecase{
		commentRepo:   deps.CommentRepo,
		postFetcher:   deps.PostFetcher,
		followChecker: deps.FollowChecker,
		mediaVerifier: deps.MediaVerifier,
		likeChecker:   deps.LikeChecker,
		statsFetcher:  deps.StatsFetcher,
		event:         deps.Event,
	}
}

func (u *usecase) GetComments(ctx context.Context, postID int64, params GetCursorParams) (common.CursorPaginatedResult[Comment, int64], error) {
	var empty common.CursorPaginatedResult[Comment, int64]
	params.Query.NormalizePagination()

	// fetch comments from db
	comments, err := u.commentRepo.GetComments(ctx, postID, params)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return empty, nil
		}
		return empty, pkg.OrInternalError(err)
	}

	// fetch stats & isLiked from cache and db
	commentsPtr := make([]*Comment, len(comments))
	for i := range comments {
		commentsPtr[i] = &comments[i]
	}
	u.mapCommentMetadata(ctx, commentsPtr, params.UserID)

	return common.NewCursorPaginatedResult(comments, params.Query.Limit), nil
}

func (u *usecase) GetReplies(ctx context.Context, parentID int64, params GetCursorParams) (common.CursorPaginatedResult[Comment, int64], error) {
	var empty common.CursorPaginatedResult[Comment, int64]
	params.Query.NormalizePagination()

	// fetch comments from db
	comments, err := u.commentRepo.GetReplies(ctx, parentID, params)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return empty, nil
		}
		return empty, pkg.OrInternalError(err)
	}

	// fetch stats & isLiked from cache and db
	commentsPtr := make([]*Comment, len(comments))
	for i := range comments {
		commentsPtr[i] = &comments[i]
	}
	u.mapCommentMetadata(ctx, commentsPtr, params.UserID)

	return common.NewCursorPaginatedResult(comments, params.Query.Limit), nil
}

func (u *usecase) DeleteComment(ctx context.Context, commentID int64, userID int64) error {
	comment, err := u.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return pkg.NewError(pkg.ErrBadRequest, "comment not found")
		}
		return pkg.OrInternalError(err)
	}

	if comment.UserID != userID {
		return pkg.NewError(pkg.ErrForbidden, "only the author has that right")
	}

	if err := u.commentRepo.Delete(ctx, comment.ID); err != nil {
		return pkg.OrInternalError(err)
	}

	_ = u.addCommentEvent(ctx, userID, *comment, common.EventCommentDeleted)
	return nil
}

func (u *usecase) UpdateComment(ctx context.Context, params UpdateParams) (*Comment, error) {
	comment, err := u.commentRepo.GetByID(ctx, params.CommentID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(pkg.ErrBadRequest, "comment not found")
		}
		return nil, pkg.OrInternalError(err)
	}

	if comment.UserID != params.UserID {
		return nil, pkg.NewError(pkg.ErrForbidden, "only the author has that right")
	}

	if params.Content != nil {
		comment.Content = *params.Content
	}
	if params.Media != nil {
		res, err := u.validateMedia(ctx, params.Media)
		if err != nil {
			return nil, err
		}
		comment.Media = res
	}

	if err := u.commentRepo.Update(ctx, comment); err != nil {
		return nil, pkg.OrInternalError(err)
	}
	return comment, nil
}

func (u *usecase) CreateComment(ctx context.Context, params CreateParams) (*Comment, error) {
	_, err := u.validatePostVisibility(ctx, params.PostID, params.UserID)
	if err != nil {
		return nil, err
	}

	finalParentID, err := u.resolveParentID(ctx, params.ParentID, params.PostID)
	if err != nil {
		return nil, err
	}

	if err := u.mediaVerifier.VerifyMedia(ctx, params.GetMediaKeys()); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(pkg.ErrBadRequest, "media validation failed")
		}
		return nil, pkg.OrInternalError(err)
	}

	comment := &Comment{
		PostID:   params.PostID,
		UserID:   params.UserID,
		ParentID: finalParentID,
		Content:  params.Content,
		Media:    params.Media,
	}

	if err := u.commentRepo.Create(ctx, comment); err != nil {
		return nil, pkg.OrInternalError(err)
	}

	_ = u.addCommentEvent(ctx, params.UserID, *comment, common.EventCommentCreated)

	return comment, nil
}

func (u *usecase) validatePostVisibility(ctx context.Context, postID, userID int64) (*post.Post, error) {
	p, err := u.postFetcher.GetPostDetail(ctx, postID, userID)
	if err != nil {
		return nil, pkg.NewError(pkg.ErrBadRequest, "post not found")
	}
	if p == nil {
		return nil, pkg.NewError(pkg.ErrBadRequest, "post not found")
	}

	if p.UserID == userID {
		return p, nil
	}

	switch p.Visibility {
	case post.VisibilityPrivate:
		return nil, pkg.NewError(pkg.ErrForbidden, "post is private")
	case post.VisibilityFollowers:
		rel, err := u.followChecker.GetRelationship(ctx, userID, p.UserID)
		if err != nil || !rel.IsFollowing {
			return nil, pkg.NewError(pkg.ErrForbidden, "only followers can comment")
		}
	}
	return p, nil
}

func (u *usecase) resolveParentID(ctx context.Context, parentID *int64, postID int64) (*int64, error) {
	if parentID == nil || *parentID <= 0 {
		return nil, nil
	}

	parent, err := u.commentRepo.GetByID(ctx, *parentID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(pkg.ErrBadRequest, "parent comment not found or has been deleted")
		}
		return nil, pkg.OrInternalError(err)
	}

	if parent.PostID != postID {
		return nil, pkg.NewError(pkg.ErrBadRequest, "parent comment belongs to another post")
	}

	// One-level depth
	if parent.ParentID != nil {
		return nil, pkg.NewError(pkg.ErrBadRequest, "cannot reply to a sub-comment, please reply to the root comment")
	}

	return parentID, nil
}

func (u *usecase) validateMedia(ctx context.Context, params []Media) ([]Media, error) {
	size := len(params)
	keys := make([]string, size)
	for i, m := range params {
		keys[i] = m.MediaKey
	}

	if err := u.mediaVerifier.VerifyMedia(ctx, keys); err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(pkg.ErrBadRequest, "media validation failed")
		}
		return nil, pkg.OrInternalError(err)
	}

	return params, nil
}

func (u *usecase) addCommentEvent(ctx context.Context, actorID int64, comment Comment, typ common.EventType) error {
	data := common.CommentEventPayload{
		PostID:    comment.PostID,
		CommentID: comment.ID,
		ParentID:  comment.ParentID,
		ActorID:   actorID,
		OwnerID:   comment.UserID,
		Typ:       typ,
	}
	return u.event.Publish(ctx, common.NewEvent(typ, data))
}

func (u *usecase) mapCommentMetadata(ctx context.Context, comments []*Comment, viewerID int64) {
	size := len(comments)
	if size == 0 {
		return
	}

	ids := make([]int64, size)
	for i, c := range comments {
		ids[i] = c.ID
	}

	statsMap, likedMap := u.fetchStatsAndLikes(ctx, ids, viewerID)

	for _, c := range comments {
		if statsMap != nil {
			if st, ok := statsMap[c.ID]; ok {
				c.Stat.LikesCount = st.LikesCount
				c.Stat.RepliesCount = st.RepliesCount
			}
		}
		if likedMap != nil {
			b := likedMap[c.ID]
			c.IsLiked = &b
		}
	}
}

func (u *usecase) fetchStatsAndLikes(ctx context.Context, ids []int64, viewerID int64) (map[int64]stats.CommentStats, map[int64]bool) {
	var statsMap map[int64]stats.CommentStats
	var likedMap map[int64]bool

	var g errgroup.Group

	g.Go(func() error {
		res, err := u.statsFetcher.GetCommentsStats(ctx, ids)
		if err != nil {
			pkg.Log().Error("failed to fetch comment stats, skipping", err)
			return nil
		}
		statsMap = res
		return nil
	})

	g.Go(func() error {
		res, err := u.likeChecker.IsCommentLiked(ctx, ids, viewerID)
		if err != nil {
			pkg.Log().Error("failed to fetch comment isLiked status, skipping", err)
			return nil
		}
		likedMap = res
		return nil
	})

	_ = g.Wait()

	return statsMap, likedMap
}

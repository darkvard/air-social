package post

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/sync/errgroup"

	"air-social/internal/domain/common"
	"air-social/internal/domain/stats"
	"air-social/pkg"
)

type UseCase interface {
	GetPostDetail(ctx context.Context, postID, viewerID int64) (*Post, error)
	CreatePost(ctx context.Context, params CreateParams) (*Post, error)
	UpdatePost(ctx context.Context, params UpdateParams) (*Post, error)
	DeletePost(ctx context.Context, postID int64, userID int64) error
	GetUserPosts(ctx context.Context, params GetCursorParams) (common.CursorPaginatedResult[Post, int64], error)
	GetPostSharers(ctx context.Context, params GetSharersParams) (common.CursorPaginatedResult[Sharer, int64], error)
}

type MediaVerifier interface {
	VerifyMedia(ctx context.Context, keys []string) error
}

type StatsFetcher interface {
	GetPostsStats(ctx context.Context, postIDs []int64) (map[int64]stats.PostStats, error)
}

type LikeChecker interface {
	IsPostLiked(ctx context.Context, postIDs []int64, userID int64) (map[int64]bool, error)
}

type Deps struct {
	PostRepo      Repository
	MediaVerifier MediaVerifier
	StatsFetcher  StatsFetcher
	LikeChecker   LikeChecker
	Event         common.EventPublisher
}

type usecase struct {
	postRepo      Repository
	mediaVerifier MediaVerifier
	likeChecker   LikeChecker
	statsFetcher  StatsFetcher
	event         common.EventPublisher
}

func NewUseCase(deps Deps) *usecase {
	return &usecase{
		postRepo:      deps.PostRepo,
		mediaVerifier: deps.MediaVerifier,
		statsFetcher:  deps.StatsFetcher,
		likeChecker:   deps.LikeChecker,
		event:         deps.Event,
	}
}

func (u *usecase) GetPostDetail(ctx context.Context, postID, viewerID int64) (*Post, error) {
	post, err := u.postRepo.GetDetail(ctx, postID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(err, "post not found")
		}
		return nil, pkg.OrInternalError(err)
	}

	u.mapPostMetadata(ctx, []*Post{post}, viewerID)

	return post, nil
}

func (u *usecase) CreatePost(ctx context.Context, params CreateParams) (*Post, error) {
	if strings.TrimSpace(params.Content) == "" && len(params.Media) == 0 {
		return nil, pkg.NewError(pkg.ErrInvalidData, "content or media is required")
	}

	post := &Post{
		OriginalPostID: params.OriginalPostID,
		UserID:         params.UserID,
		Content:        params.Content,
		Visibility:     params.Visibility,
	}
	if len(params.Media) > 0 {
		media, err := u.validateMedia(ctx, params.Media)
		if err != nil {
			if errors.Is(err, pkg.ErrNotFound) {
				return nil, pkg.NewError(pkg.ErrBadRequest, "media validation failed")
			}
			return nil, pkg.OrInternalError(err)
		}
		post.Media = media
	}

	if err := u.postRepo.Create(ctx, post); err != nil {
		if errors.Is(err, pkg.ErrInvalidData) {
			return nil, pkg.NewError(pkg.ErrInvalidData, "original post not found")
		}
		return nil, err
	}

	if post.OriginalPostID != nil {
		_ = u.addShareEvent(ctx, *post, true)
	}

	return post, nil
}

func (u *usecase) UpdatePost(ctx context.Context, params UpdateParams) (*Post, error) {
	post, err := u.getPostOwner(ctx, params.PostID, params.UserID)
	if err != nil {
		return nil, err
	}

	if params.Content != nil {
		post.Content = *params.Content
	}
	if params.Visibility != nil {
		post.Visibility = Visibility(*params.Visibility)
	}
	if params.Media != nil {
		media, err := u.validateMedia(ctx, params.Media)
		if err != nil {
			if errors.Is(err, pkg.ErrNotFound) {
				return nil, pkg.NewError(pkg.ErrBadRequest, "media items not found or invalid")
			}
			return nil, pkg.OrInternalError(err)
		}
		post.Media = media // update
	} else {
		post.Media = nil // skip
	}

	if err := u.postRepo.Update(ctx, post); err != nil {
		return nil, pkg.ErrInternal
	}
	return post, nil
}

func (u *usecase) DeletePost(ctx context.Context, postID, userID int64) error {
	post, err := u.getPostOwner(ctx, postID, userID)
	if err != nil {
		return err
	}
	if post.OriginalPostID != nil {
		_ = u.addShareEvent(ctx, *post, false)
	}
	return pkg.OrInternalError(u.postRepo.Delete(ctx, postID))
}

func (u *usecase) GetUserPosts(ctx context.Context, params GetCursorParams) (common.CursorPaginatedResult[Post, int64], error) {
	var empty common.CursorPaginatedResult[Post, int64]
	params.Query.NormalizePagination()

	posts, err := u.postRepo.GetUserPosts(ctx, params)
	if err != nil {
		return empty, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	postsPtr := make([]*Post, len(posts))
	for i := range posts {
		postsPtr[i] = &posts[i]
	}
	u.mapPostMetadata(ctx, postsPtr, params.UserID)

	result := common.NewCursorPaginatedResult(posts, params.Query.Limit)
	return result, nil
}

func (u *usecase) GetPostSharers(ctx context.Context, params GetSharersParams) (common.CursorPaginatedResult[Sharer, int64], error) {
	var empty common.CursorPaginatedResult[Sharer, int64]
	params.Query.NormalizePagination()

	sharers, err := u.postRepo.GetPostSharers(ctx, params)
	if err != nil {
		return empty, pkg.OrInternalError(err, pkg.ErrNotFound)
	}

	result := common.NewCursorPaginatedResult(sharers, params.Query.Limit)
	return result, nil
}

func (u *usecase) validateMedia(ctx context.Context, params []MediaParams) ([]Media, error) {
	size := len(params)

	keys := make([]string, size)
	for i, m := range params {
		keys[i] = m.MediaKey
	}
	if err := u.mediaVerifier.VerifyMedia(ctx, keys); err != nil {
		return nil, err
	}

	media := make([]Media, size)
	for i, v := range params {
		media[i] = v.ToDomain()
	}
	return media, nil
}

func (u *usecase) getPostOwner(ctx context.Context, postID, viewerID int64) (*Post, error) {
	post, err := u.postRepo.GetByID(ctx, postID)
	if err != nil {
		if errors.Is(err, pkg.ErrNotFound) {
			return nil, pkg.NewError(err, "post not found")
		}
		return nil, pkg.OrInternalError(err)
	}

	if post.UserID != viewerID {
		return nil, pkg.ErrForbidden
	}
	return post, nil
}

func (u *usecase) addShareEvent(ctx context.Context, post Post, isShare bool) error {
	data := common.ShareEventPayload{
		OriginalPostID: *post.OriginalPostID,
		NewPostID:      post.ID,
		ActorID:        post.UserID,
		IsShared:       isShare,
	}
	return u.event.Publish(ctx, common.NewEvent(common.EventPostShare, data))
}

func (u *usecase) mapPostMetadata(ctx context.Context, posts []*Post, viewerID int64) {
	size := len(posts)
	if size == 0 {
		return
	}

	ids := make([]int64, size)
	for i, p := range posts {
		ids[i] = p.ID
	}

	statsMap, likedMap := u.fetchStatsAndLikes(ctx, ids, viewerID)

	for _, p := range posts {
		if statsMap != nil {
			if st, ok := statsMap[p.ID]; ok {
				p.Stat.LikesCount = st.LikesCount
				p.Stat.CommentsCount = st.CommentsCount
				p.Stat.SharesCount = st.SharesCount
			}
		}
		if likedMap != nil {
			b := likedMap[p.ID]
			p.IsLiked = &b
		}
	}
}

func (u *usecase) fetchStatsAndLikes(ctx context.Context, ids []int64, viewerID int64) (map[int64]stats.PostStats, map[int64]bool) {
	var statsMap map[int64]stats.PostStats
	var likedMap map[int64]bool

	var g errgroup.Group

	g.Go(func() error {
		res, err := u.statsFetcher.GetPostsStats(ctx, ids)
		if err != nil {
			pkg.Log().Error("failed to fetch post stats, skipping", err)
			return nil
		}
		statsMap = res
		return nil
	})

	g.Go(func() error {
		res, err := u.likeChecker.IsPostLiked(ctx, ids, viewerID)
		if err != nil {
			pkg.Log().Error("failed to fetch isLiked status, skipping", err)
			return nil
		}
		likedMap = res
		return nil
	})

	_ = g.Wait()

	return statsMap, likedMap
}

package provider

import (
	"time"

	"air-social/internal/cache"
	"air-social/internal/config"
	"air-social/internal/domain/auth"
	"air-social/internal/domain/comment"
	"air-social/internal/domain/feed"
	feedusecase "air-social/internal/domain/feed/usecase"
	"air-social/internal/domain/follow"
	"air-social/internal/domain/health"
	"air-social/internal/domain/like"
	"air-social/internal/domain/media"
	"air-social/internal/domain/post"
	"air-social/internal/domain/search"
	"air-social/internal/domain/stats"
	"air-social/internal/domain/user"
	userusecase "air-social/internal/domain/user/usecase"
	"air-social/internal/infrastructure/minio"
	"air-social/internal/infrastructure/mongo"
	"air-social/internal/infrastructure/postgres"
	"air-social/internal/infrastructure/rabbitmq"
	"air-social/internal/infrastructure/redis"
)

const (
	followerCacheL1TTL = 2 * time.Minute
	followerCacheL2TTL = 5 * time.Minute

	userSummaryCacheL1TTL = 5 * time.Minute
	userSummaryCacheL2TTL = 12 * time.Hour

	postCacheL1TTL = 2 * time.Minute
	postCacheL2TTL = 6 * time.Hour
)

type UseCase struct {
	Health  health.UseCase
	User    user.UseCase
	Auth    auth.UseCase
	Media   media.UseCase
	Follow  follow.UseCase
	Post    post.UseCase
	Like    like.UseCase
	Comment comment.UseCase
	Stats   stats.UseCase
	Feed    feed.UseCase
	Search  search.UseCase
}

type UseCaseDeps struct {
	Cfg     config.Config
	Infra   *Infrastructure
	Prov    Provider
	Repo    Repository
	Adapter Adapter
}

func NewUseCase(deps UseCaseDeps) UseCase {
	healthUC := getHealthUseCase(deps)
	mediaUC := getMediaUseCase(deps)
	userUC := getUserUseCase(deps, mediaUC)
	authUC := getAuthUseCase(deps, userUC.Fetch, userUC.Account)

	l1 := cache.NewMemCache[[]int64](1000, followerCacheL1TTL)
	l2 := redis.NewRedisStore[[]int64](deps.Infra.Redis)
	followerCache := cache.NewTieredCache(l1, l2, followerCacheL1TTL, followerCacheL2TTL)

	followUC := getFollowUseCase(deps, userUC.Fetch, followerCache)
	likeUC := getLikeUseCase(deps)
	statsUC := getStatsUseCase(deps)

	postL1 := cache.NewMemCache[post.Post](1000, postCacheL1TTL)
	postL2 := redis.NewRedisStore[post.Post](deps.Infra.Redis)
	postTiered := cache.NewTieredCache(postL1, postL2, postCacheL1TTL, postCacheL2TTL)
	postUC := getPostUseCase(deps, mediaUC, statsUC, likeUC, post.NewCache(postTiered))
	commentUC := getCommentUseCase(deps, postUC, followUC, mediaUC, likeUC, statsUC)
	feedUC := getFeedUseCase(deps, followUC, postUC, followerCache)
	searchUC := search.NewUseCase(deps.Repo.Search, postUC)

	return UseCase{
		Health:  healthUC,
		User:    userUC,
		Auth:    authUC,
		Media:   mediaUC,
		Follow:  followUC,
		Post:    postUC,
		Like:    likeUC,
		Comment: commentUC,
		Stats:   statsUC,
		Feed:    feedUC,
		Search:  searchUC,
	}
}

func getHealthUseCase(deps UseCaseDeps) health.UseCase {
	healths := make(map[string]health.Checker)
	healths["postgres"] = postgres.NewHealth(deps.Infra.Postgres)
	healths["mongo"] = mongo.NewHealth(deps.Infra.Mongo)
	healths["redis"] = redis.NewHealth(deps.Infra.Redis)
	healths["rabbitmq"] = rabbitmq.NewHealth(deps.Infra.Rabbit, deps.Cfg.RabbitMQ)
	healths["minio"] = minio.NewHealth(deps.Infra.Minio, deps.Cfg.MinIO.BucketPublic)

	return health.NewUseCase(
		healths, deps.Prov.Link.SystemProvider,
	)
}

func getMediaUseCase(deps UseCaseDeps) media.UseCase {
	return media.NewUseCase(
		media.Deps{
			Bucket: media.Bucket{
				Public:  deps.Cfg.MinIO.BucketPublic,
				Private: deps.Cfg.MinIO.BucketPrivate,
			},
			Storage: deps.Adapter.Media,
			Link:    deps.Prov.Link,
		},
	)
}

func getUserUseCase(deps UseCaseDeps, confirmer userusecase.MediaConfirmer) user.UseCase {
	l1 := cache.NewMemCache[user.UserSummary](500, 5*time.Minute)
	l2 := redis.NewRedisStore[user.UserSummary](deps.Infra.Redis)
	userTiered := cache.NewTieredCache(l1, l2, userSummaryCacheL1TTL, userSummaryCacheL2TTL)
	userCache := user.NewCache(userTiered)

	return user.UseCase{
		Account: userusecase.NewAccountUseCase(userusecase.AccountDeps{Repo: deps.Repo.User, Cache: userCache}),
		Profile: userusecase.NewProfileUseCase(userusecase.ProfileDeps{Repo: deps.Repo.User, Cache: userCache, Media: confirmer}),
		Fetch:   userusecase.NewFetchUseCase(userusecase.FetchDeps{Repo: deps.Repo.User, Cache: userCache, Link: deps.Prov.Link.LinkProvider}),
	}
}

func getAuthUseCase(deps UseCaseDeps, userFetch user.FetchUseCase, userAccount user.AccountUseCase) auth.UseCase {
	return auth.NewUseCase(
		auth.Deps{
			TokenRepo:      deps.Repo.Token,
			TokenProvider:  deps.Prov.Token,
			VerifyProvider: deps.Prov.Verify,
			UserFetch:      userFetch,
			UserAccount:    userAccount,
		},
	)
}

func getFollowUseCase(deps UseCaseDeps, fetcher follow.UserFetcher, cacheInvalidator follow.CacheInvalidator) follow.UseCase {
	return follow.NewUseCase(
		follow.Deps{
			FollowRepo:       deps.Repo.Follow,
			UserFetcher:      fetcher,
			CacheInvalidator: cacheInvalidator,
		},
	)
}

func getPostUseCase(deps UseCaseDeps, mediaVerifier post.MediaVerifier, statsFetcher post.StatsFetcher, likeChecker post.LikeChecker, postCache post.Cache) post.UseCase {
	return post.NewUseCase(
		post.Deps{
			PostRepo:      deps.Repo.Post,
			MediaVerifier: mediaVerifier,
			StatsFetcher:  statsFetcher,
			LikeChecker:   likeChecker,
			Event:         deps.Adapter.EventPub,
			PostCache:     postCache,
		},
	)
}

func getCommentUseCase(
	deps UseCaseDeps,
	postFetcher comment.PostFetcher,
	followFetcher comment.FollowChecker,
	mediaVerifier comment.MediaVerifier,
	likeChecker comment.LikeChecker,
	statsFetcher comment.StatsFetcher,
) comment.UseCase {
	return comment.NewUseCase(comment.Deps{
		CommentRepo:   deps.Repo.Comment,
		PostFetcher:   postFetcher,
		FollowChecker: followFetcher,
		MediaVerifier: mediaVerifier,
		LikeChecker:   likeChecker,
		StatsFetcher:  statsFetcher,
		Event:         deps.Adapter.EventPub,
	})
}

func getLikeUseCase(deps UseCaseDeps) like.UseCase {
	return like.NewUsecase(
		like.Deps{
			Repo:  deps.Repo.Like,
			Event: deps.Adapter.EventPub,
		},
	)
}

func getStatsUseCase(deps UseCaseDeps) stats.UseCase {
	return stats.NewUseCase(
		stats.Deps{
			Repo:  deps.Repo.Stats,
			Cache: deps.Prov.Stats,
		},
	)
}

func getFeedUseCase(deps UseCaseDeps, followFetcher feedusecase.FollowFetcher, postFetcher feedusecase.PostFetcher, followerCache cache.TieredStore[[]int64]) feed.UseCase {
	return feed.UseCase{
		Command: feedusecase.NewCommandUseCase(feedusecase.CommandDeps{
			CacheProvider: deps.Prov.Feed,
			FollowFetcher: followFetcher,
			FollowerCache: followerCache,
		}),
		Query: feedusecase.NewQueryUseCase(feedusecase.QueryDeps{
			CacheProvider: deps.Prov.Feed,
			PostFetcher:   postFetcher,
		}),
	}
}

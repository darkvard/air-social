package provider

import (
	"air-social/internal/config"
	"air-social/internal/domain/auth"
	"air-social/internal/domain/comment"
	"air-social/internal/domain/follow"
	"air-social/internal/domain/health"
	"air-social/internal/domain/like"
	"air-social/internal/domain/media"
	"air-social/internal/domain/post"
	"air-social/internal/domain/user"
	userusecase "air-social/internal/domain/user/usecase"
	"air-social/internal/infrastructure/minio"
	"air-social/internal/infrastructure/postgres"
	"air-social/internal/infrastructure/rabbitmq"
	"air-social/internal/infrastructure/redis"
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
	followUC := getFollowUseCase(deps, userUC.Fetch)
	postUC := getPostUseCase(deps, mediaUC)
	likeUC := getLikeUseCase(deps)
	commentUC := getCommentUseCase(deps, postUC, followUC, mediaUC, likeUC)

	return UseCase{
		Health:  healthUC,
		User:    userUC,
		Auth:    authUC,
		Media:   mediaUC,
		Follow:  followUC,
		Post:    postUC,
		Like:    likeUC,
		Comment: commentUC,
	}
}

func getHealthUseCase(deps UseCaseDeps) health.UseCase {
	healths := make(map[string]health.Checker)
	healths["postgres"] = postgres.NewHealth(deps.Infra.DB)
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
	d := userusecase.Deps{
		Repo:  deps.Repo.User,
		Cache: deps.Adapter.Cache,
		Link:  deps.Prov.Link.LinkProvider,
		Media: confirmer,
	}
	return user.UseCase{
		Account: userusecase.NewAccountUseCase(d),
		Profile: userusecase.NewProfileUseCase(d),
		Fetch:   userusecase.NewFetchUseCase(d),
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
			Cache:          deps.Adapter.Cache,
		},
	)
}

func getFollowUseCase(deps UseCaseDeps, fetcher follow.UserFetcher) follow.UseCase {
	return follow.NewUseCase(
		follow.Deps{
			FollowRepo:  deps.Repo.Follow,
			UserFetcher: fetcher,
		},
	)
}

func getPostUseCase(deps UseCaseDeps, mediaVerifier post.MediaVerifier) post.UseCase {
	return post.NewUseCase(
		post.Deps{
			PostRepo:      deps.Repo.Post,
			MediaVerifier: mediaVerifier,
		},
	)
}

func getCommentUseCase(
	deps UseCaseDeps,
	postFetcher comment.PostFetcher,
	followFetcher comment.FollowChecker,
	mediaVerifier comment.MediaVerifier,
	likeChecker comment.LikeChecker,
) comment.UseCase {
	return comment.NewUseCase(comment.Deps{
		CommentRepo:   deps.Repo.Comment,
		PostFetcher:   postFetcher,
		FollowChecker: followFetcher,
		MediaVerifier: mediaVerifier,
		LikeChecker:   likeChecker,
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

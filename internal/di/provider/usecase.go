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
	uuc "air-social/internal/domain/user/usecase"
	"air-social/internal/infrastructure/mailer"
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
	Comment comment.UseCase
	Like    like.UseCase
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

	return UseCase{
		Health:  healthUC,
		User:    userUC,
		Auth:    authUC,
		Media:   mediaUC,
		Follow:  followUC,
		Post:    postUC,
		Comment: nil,
		Like:    nil,
	}
}

func getHealthUseCase(deps UseCaseDeps) health.UseCase {
	healths := make(map[string]health.Checker)
	healths["postgres"] = postgres.NewHealth(deps.Infra.DB)
	healths["redis"] = redis.NewHealth(deps.Infra.Redis)
	healths["rabbitmq"] = rabbitmq.NewHealth(deps.Infra.Rabbit, deps.Cfg.RabbitMQ)
	healths["minio"] = minio.NewHealth(deps.Infra.Minio)
	healths["mailtrap"] = mailer.NewHealth(deps.Infra.Mailtrap)

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

func getUserUseCase(deps UseCaseDeps, confirmer uuc.MediaConfirmer) user.UseCase {
	d := uuc.Deps{
		Repo:  deps.Repo.User,
		Cache: deps.Adapter.Cache,
		Link:  deps.Prov.Link.LinkProvider,
		Media: confirmer,
	}
	return user.UseCase{
		Account: uuc.NewAccountUseCase(d),
		Profile: uuc.NewProfileUseCase(d),
		Fetch:   uuc.NewFetchUseCase(d),
	}
}

func getAuthUseCase(deps UseCaseDeps, userFetch user.FetchUseCase, userAccount user.AccountUseCase) auth.UseCase {
	return auth.NewUseCase(
		auth.Deps{
			TokenRepo:     deps.Repo.Token,
			TokenProvider: deps.Prov.Token,
			UserFetch:     userFetch,
			UserAccount:   userAccount,
			Cache:         deps.Adapter.Cache,
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

func getCommentUseCase() comment.UseCase {
	return nil
}

func getLikeUseCase() like.UseCase {
	return nil
}

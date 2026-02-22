package provider

// import (
//
//
// 

// 	"air-social/internal/config"
// 	"air-social/internal/domain"
// 	"air-social/internal/service"
// )

// type Services struct {
// 	Health service.HealthService
// 	Auth   service.AuthService
// 	User   service.UserService
// 	Token  service.TokenService
// 	Email  service.EmailService
// 	Verify service.VerifyService
// 	Media  service.MediaService
// 	Follow service.FollowService
// 	Post   service.PostService
// }

// func NewServices(
// 	cfg config.Config,
// 	urlFactory domain.URLFactory,
// 	infra *Infrastructure,
// 	repo *Repository,
// 	adapter *Adapter,
// ) *Services {

// 	media := service.NewMediaService(adapter.Media, cfg.MinIO, urlFactory)
// 	health := service.NewHealthService(infra.DB, infra.Redis, infra.GetRabbit(cfg.RabbitMQ), infra.Minio, urlFactory)
// 	token := service.NewTokenService(repo.Token, cfg.Token)
// 	verify := service.NewVerifyService(adapter.Cache, adapter.EventPub, urlFactory)
// 	user := service.NewUserService(repo.User, media, adapter.Cache, urlFactory)
// 	auth := service.NewAuthService(user, token, verify, adapter.Cache)
// 	email := service.NewEmailService(adapter.Mailer)
// 	follow := service.NewFollowService(repo.Follow, adapter.Cache, user)
// 	post := service.NewPostService(repo.Post, media, user)

// 	return &Services{
// 		Health: health,
// 		Auth:   auth,
// 		User:   user,
// 		Token:  token,
// 		Email:  email,
// 		Verify: verify,
// 		Media:  media,
// 		Follow: follow,
// 		Post:   post,
// 	}
// }

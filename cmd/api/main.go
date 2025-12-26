package main

import (
	"air-social/internal/app"
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/infrastructure/mailer"
	"air-social/pkg"
	"air-social/templates"
)

func main() {
	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}
	defer app.Cleanup()

	pkg.NewLogger(app.Config.Server.Env)
	welcome(app.Config.Server)

	testEmail(app.Config.Mailer)

	app.Run()
}

func welcome(server config.ServerConfig) {
	pkg.Log().Infow("server started",
		"container_port", server.Port,
		"host_port", server.HostPort,
	)
}

func testEmail(cfg config.MailConfig) {
	err := mailer.NewMailtrap(cfg).Send(
		&domain.EmailEnvelope{
			To:           "User@gmail.com",
			LayoutFile:   templates.LayoutPath,
			TemplateFile: templates.VerifyEmailPath,
			Data: domain.VerifyEmailData{
				Name:   "Test User",
				Link:   "https://air-social.com/verify?token=abc123",
				Expiry: "24 hours",
			},
		},
	)
	if err != nil {
		pkg.Log().Errorw("test email send failed", "error", err)
	} else {
		pkg.Log().Infow("test email sent successfully")
	}
}

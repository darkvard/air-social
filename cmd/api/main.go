package main

import (
	"air-social/internal/app"
	"air-social/internal/config"
	"air-social/internal/domain"
	"air-social/internal/infrastructure/mailer"
	"air-social/pkg"
)

func main() {
	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}
	defer app.Cleanup()

	pkg.NewLogger(app.Config.Server.Env )
	welcome(app.Config.Server)

	testMailTrap(&app.Config.Mailer)

	app.Run()
}

func welcome(server config.ServerConfig) {
	pkg.Log().Infow("server started",
		"container_port", server.Port,
		"host_port", server.HostPort,
	)
}

func testMailTrap(cfg *config.MailConfig) {
	go func() {
		envelope := &domain.EmailEnvelope{
			To:           "test@example.com",
			TemplateFile: "welcome.tmpl",
			Data: map[string]interface{}{
				"Name":        "Dev",
				"LuckyNumber": 6868,
			},
		}

		mail := mailer.NewMailtrapSender(cfg)
		if err := mail.Send(envelope); err != nil {
			pkg.Log().Errorw("send email", "error", err.Error())
		} else {
			pkg.Log().Infow("send email success")
		}
	}()

}

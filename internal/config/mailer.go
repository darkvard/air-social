package config

type MailConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
}

func MailCfg() MailConfig {
	return MailConfig{
		Host:        getString("MAILTRAP_HOST", "smtp.mailtrap.io"),
		Port:        getInt("MAILTRAP_PORT", 587),
		Username:    getString("MAILTRAP_USERNAME", "fd2dfd6851a016"),
		Password:    getString("MAILTRAP_PASSWORD", ""),
		FromAddress: getString("MAILTRAP_FROM_ADDRESS", "no-reply@airsocial.com"),
		FromName:    getString("MAILTRAP_FROM_NAME", "Air Social"),
	}
}

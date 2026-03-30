package config

import "time"

type LimiterConfig struct {
	Limit  int
	Window time.Duration // sliding window counter rate limiter
}

// Some rate limit configs (e.g. feed, like...) are intentionally omitted.
// These are handled by the algorithm layer to avoid duplicated logic.
type RateLimiterConfig struct {
	Login          LimiterConfig
	Register       LimiterConfig
	ForgotPassword LimiterConfig
	Post           LimiterConfig
	Comment        LimiterConfig
	Follow         LimiterConfig
	MediaPresigned LimiterConfig
}

func RateLimiterCfg() RateLimiterConfig {
	return RateLimiterConfig{
		Login: LimiterConfig{
			Limit:  10,
			Window: time.Minute * 15,
		},
		Register: LimiterConfig{
			Limit:  10,
			Window: time.Minute * 15,
		},
		ForgotPassword: LimiterConfig{
			Limit:  3,
			Window: time.Hour,
		},
		Post: LimiterConfig{
			Limit:  20,
			Window: time.Hour,
		},
		Comment: LimiterConfig{
			Limit:  30,
			Window: time.Minute * 15,
		},
		Follow: LimiterConfig{
			Limit:  50,
			Window: time.Hour,
		},
		MediaPresigned: LimiterConfig{		 
			Limit:  20,
			Window: time.Hour,
		},
	}
}

package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

type SeedConfig struct {
	Users struct {
		Total int `yaml:"total"`
	}
	Follows struct {
		PerUser int `yaml:"per_user"`
	}
	Posts struct {
		PerUser      int `yaml:"per_user"`
		MediaPerPost int `yaml:"media_per_post"`
	}
	Comments struct {
		PerPost         int `yaml:"per_post"`
		MediaPerComment int `yaml:"media_per_comment"`
	}
	Likes struct {
		PerPost    int `yaml:"per_post"`
		PerComment int `yaml:"per_comment"`
	}
	Conversations struct {
		DirectActive  int `yaml:"direct_active"`
		DirectPending int `yaml:"direct_pending"`
		Groups        int `yaml:"groups"`
	}
}

func Load() SeedConfig {
	var cfg SeedConfig

	data, err := os.ReadFile("seed.yml")
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return cfg
}

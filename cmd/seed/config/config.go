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
		PerUser int `yaml:"per_user"`
	}
	Comments struct {
		PerPost int `yaml:"per_post"`
	}
	Likes struct {
		PerPost int `yaml:"per_post"`
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

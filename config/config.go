package config

import (
	"strconv"

	"github.com/ZakirAvrora/exchange-rate/pkg/env"
)

const _defaultMaxPoolSize = 1

type Config struct {
	HTTP           HTTP
	PostgresConfig PostgresConfig
}

type HTTP struct {
	Port string
}

type PostgresConfig struct {
	Host     string
	Port     string
	DbName   string
	User     string
	Password string
	PoolMax  int
}

func NewConfig(path string) *Config {
	env.CheckDotEnv(path)
	maxPool, err := strconv.Atoi(env.MustGet("PG_POOL_MAX"))
	if err != nil {
		// NoReturnErr: use defaultMaxPoolSize
		maxPool = _defaultMaxPoolSize
	}

	return &Config{
		HTTP: HTTP{
			Port: "8080",
		},
		PostgresConfig: PostgresConfig{
			Host:     env.MustGet("PG_DATABASE_HOST"),
			Port:     env.MustGet("PG_DATABASE_PORT"),
			User:     env.MustGet("PG_DATABASE_USER"),
			DbName:   env.MustGet("PG_DATABASE_DB"),
			Password: env.MustGet("PG_DATABASE_PASSWORD"),
			PoolMax:  maxPool,
		},
	}
}

package config

import "time"

type Config struct {
	MySQL struct {
		DSN                string `mapstructure:"dsn"`
		Automigrate        bool
		MaxOpenConnections int `mapstructure:"max_open_connections"`
		MaxIdleConnections int `mapstructure:"max_idle_connections"`
	}

	Redis struct {
		URL         string        `mapstructure:"url"`
		MaxIdle     int           `mapstructure:"max_idle"`
		MaxActive   int           `mapstructure:"max_active"`
		IdleTimeout time.Duration `mapstructure:"idle_timeout"`
	}
}

var C Config

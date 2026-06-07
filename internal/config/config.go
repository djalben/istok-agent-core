package config

import (
	"gitlab.com/libs-artifex/envparse"
	"gitlab.com/libs-artifex/wrapper/v2"
)

// Config — конфигурация сервиса из окружения и .env (envparse).
type Config struct {
	Port string `env:"PORT" default:"8080"`

	GoEnv        string `env:"GO_ENV"`
	RailwayEnv   string `env:"RAILWAY_ENVIRONMENT"`
	AnthropicKey string `env:"ANTHROPIC_API_KEY"`
	ReplicateKey string `env:"REPLICATE_API_TOKEN"`

	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS"`
	JWTSecret          string `env:"JWT_SECRET"`
	DatabaseURL        string `env:"DATABASE_URL"`

	LogLevel string `env:"LOG_LEVEL" default:"log"`
	LogPlain bool   `env:"LOG_PLAIN" default:"false"`

	AutoFixMaxRetries int `env:"AUTO_FIX_MAX_RETRIES" default:"2"`
}

// IsProduction — production, если GO_ENV или RAILWAY_ENVIRONMENT равны "production".
func (c Config) IsProduction() bool {
	env := c.RailwayEnv
	if env == "" {
		env = c.GoEnv
	}

	return env == "production"
}

// Parse загружает конфиг из переменных окружения и .env.
func Parse() (Config, error) {
	var cfg Config

	err := envparse.Process("", &cfg)
	if err != nil {
		return Config{}, wrapper.Wrap(err)
	}

	return cfg, nil
}

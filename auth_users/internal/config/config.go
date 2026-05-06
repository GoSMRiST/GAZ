package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	AuthHostAddress string        `env:"REST_AUTH_HOST_ADDRESS"`
	UserHostAddress string        `env:"REST_USER_HOST_ADDRESS"`
	AuthGrpcPort    string        `env:"GRPC_AUTH_PORT"`
	UserGrpcPort    string        `env:"GRPC_USER_PORT"`
	ServTimeout     time.Duration `env:"SERV_TIMEOUT"`
	DBHost          string        `env:"DB_HOST"`
	DBPort          string        `env:"DB_PORT"`
	DBUser          string        `env:"DB_USER"`
	DBPassword      string        `env:"DB_PASSWORD"`
	DBName          string        `env:"DB_NAME"`
	RedisAddress    string        `env:"REDIS_ADDRESS"`
	RedisPassword   string        `env:"REDIS_PASSWORD"`
	RedisDb         int           `env:"REDIS_DB"`
	SmtpHost        string        `env:"SMTP_HOST"`
	SmtpPort        string        `env:"SMTP_PORT"`
	SmtpUser        string        `env:"SMTP_USER"`
	SmtpPass        string        `env:"SMTP_PASS"`
	AppURL          string        `env:"APP_URL"`
	LogLevel        string        `env:"LOG_LEVEL"`
	JwtSecret       string        `env:"JWT_SECRET"`
	TokenTTL        time.Duration `env:"TOKEN_TTL"`
}

func InitConfig() (*Config, error) {
	_ = godotenv.Load("internal/config/config.env") // в проде переменные придут из окружения

	conf := Config{}

	if err := env.Parse(&conf); err != nil {
		return nil, err
	}

	if err := conf.validate(); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (c *Config) validate() error {
	required := map[string]string{
		"REST_AUTH_HOST_ADDRESS": c.AuthHostAddress,
		"REST_USER_HOST_ADDRESS": c.UserHostAddress,
		"GRPC_AUTH_PORT":         c.AuthGrpcPort,
		"GRPC_USER_PORT":         c.UserGrpcPort,
		"DB_HOST":                c.DBHost,
		"DB_PORT":                c.DBPort,
		"DB_USER":                c.DBUser,
		"DB_PASSWORD":            c.DBPassword,
		"DB_NAME":                c.DBName,
		"REDIS_ADDRESS":          c.RedisAddress,
		"SMTP_HOST":              c.SmtpHost,
		"SMTP_PORT":              c.SmtpPort,
		"SMTP_USER":              c.SmtpUser,
		"SMTP_PASS":              c.SmtpPass,
		"APP_URL":                c.AppURL,
		"JWT_SECRET":             c.JwtSecret,
	}

	for key, val := range required {
		if val == "" {
			return fmt.Errorf("config: %s is required", key)
		}
	}

	if c.TokenTTL == 0 {
		return fmt.Errorf("config: TOKEN_TTL is required")
	}

	if c.ServTimeout == 0 {
		c.ServTimeout = 5 * time.Second
	}

	return nil
}

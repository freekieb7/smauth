package internal

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	OAuth    OAuthConfig
	Init     InitConfig
}

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
)

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Environment  Environment
}

type ContextKey string

const (
	SessionContextKey     ContextKey = "session"
	AccessTokenContextKey ContextKey = "access_token"
)

type DatabaseConfig struct {
	URL      string `env:"DATABASE_URL"`
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
}

func (d *DatabaseConfig) GetURL() string {
	if d.URL != "" {
		return d.URL
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}

type LoggerConfig struct {
	Level slog.Level
}

type OAuthConfig struct {
	TokenExpiration time.Duration
}

type InitConfig struct {
	AdminEmail    string
	AdminPassword string
}

func NewConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:        3001,
			Environment: getEnvEnvironment("SERVER_ENVIRONMENT", EnvironmentDevelopment),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", ""),
		},
		Logger: LoggerConfig{
			Level: getEnvLogLevel("LOGGER_LEVEL", slog.LevelWarn),
		},
		OAuth: OAuthConfig{
			TokenExpiration: time.Duration(getEnvInt("OAUTH_TOKEN_EXPIRATION_MINUTES", 60)) * time.Minute,
		},
		Init: InitConfig{
			AdminEmail:    getEnv("ADMIN_EMAIL", "admin@example.com"),
			AdminPassword: getEnv("ADMIN_PASSWORD", "admin"),
		},
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	if defaultValue == "" {
		panic("Missing required environment variable: " + key)
	}

	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvEnvironment(key string, defaultValue Environment) Environment {
	if value, exists := os.LookupEnv(key); exists {
		return Environment(value)
	}
	return defaultValue
}

func getEnvLogLevel(key string, defaultValue slog.Level) slog.Level {
	if value, exists := os.LookupEnv(key); exists {
		switch value {
		case "debug":
			return slog.LevelDebug
		case "info":
			return slog.LevelInfo
		case "warn":
			return slog.LevelWarn
		case "error":
			return slog.LevelError
		}
	}
	return defaultValue
}

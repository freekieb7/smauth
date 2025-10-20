package internal

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Setup    SetupConfig
	Server   ServerConfig
	Database DatabaseConfig
	OAuth    OAuthConfig
	Logger   LoggerConfig
}

type SetupConfig struct {
	AdminEmail    string
	AdminPassword string
}

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
)

type ServerConfig struct {
	URL         string
	Name        string
	Port        int
	Environment Environment
	Version     string
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

type OAuthConfig struct {
	TokenExpiration time.Duration
	Microsoft       MicrosoftOAuthConfig
}

type MicrosoftOAuthConfig struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	RedirectURI  string
}

func (m *MicrosoftOAuthConfig) Enabled() bool {
	return m.ClientID != "" && m.ClientSecret != "" && m.RedirectURI != ""
}

type LoggerConfig struct {
	Level slog.Level
}

func NewConfig() Config {
	return Config{
		Setup: SetupConfig{
			AdminEmail:    getEnv("ADMIN_EMAIL", "admin@smauth.com", false),
			AdminPassword: getEnv("ADMIN_PASSWORD", "admin", false),
		},
		Server: ServerConfig{
			URL:         getEnv("SERVER_URL", "http://localhost:3001", false),
			Name:        getEnv("SERVER_NAME", "smauth", false),
			Port:        getEnvInt("SERVER_PORT", 3001, false),
			Environment: getEnvEnvironment("SERVER_ENVIRONMENT", EnvironmentProduction, false),
			Version:     getEnv("SERVER_VERSION", "development", false),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "", true),
		},
		OAuth: OAuthConfig{
			TokenExpiration: time.Duration(getEnvInt("OAUTH_TOKEN_EXPIRATION_MINUTES", 60, false)) * time.Minute,
			Microsoft: MicrosoftOAuthConfig{
				ClientID:     getEnv("MICROSOFT_CLIENT_ID", "", false),
				ClientSecret: getEnv("MICROSOFT_CLIENT_SECRET", "", false),
				TenantID:     getEnv("MICROSOFT_TENANT_ID", "common", false),
				RedirectURI:  getEnv("MICROSOFT_REDIRECT_URI", "", false),
			},
		},
		Logger: LoggerConfig{
			Level: getEnvLogLevel("LOGGER_LEVEL", slog.LevelInfo, false),
		},
	}
}

func getEnv(key string, defaultValue string, required bool) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	if required {
		panic("Missing required environment variable: " + key)
	}

	return defaultValue
}

func getEnvInt(key string, defaultValue int, required bool) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}

	if required {
		panic("Missing required environment variable: " + key)
	}

	return defaultValue
}

func getEnvEnvironment(key string, defaultValue Environment, required bool) Environment {
	if value, exists := os.LookupEnv(key); exists {
		return Environment(value)
	}
	if required {
		panic("Missing required environment variable: " + key)
	}
	return defaultValue
}

func getEnvLogLevel(key string, defaultValue slog.Level, required bool) slog.Level {
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

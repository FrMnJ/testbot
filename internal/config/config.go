package config

import (
	"fmt"
	"os"
)

var (
	ErrMissingModel        error = fmt.Errorf("MODEL is required (or set GENKIT_MODEL)")
	ErrMissingWebsocketURL error = fmt.Errorf("WEBSOCKET_URL is required")
	ErrMissingDatabaseURL  error = fmt.Errorf("DATABASE_URL is required")
	ErrMissingJWTSecret    error = fmt.Errorf("JWT_SECRET is required")
)

type Config struct {
	JWTSecret    string
	WebsocketURL string
	Model        string
	DatabaseURL  string
}

var cfg *Config

func LoadConfig() (*Config, error) {
	model := getEnv("MODEL", "")
	if model == "" {
		model = getEnv("GENKIT_MODEL", "")
	}

	cfg := &Config{
		JWTSecret:    getEnv("JWT_SECRET", ""),
		WebsocketURL: getEnv("WEBSOCKET_URL", ""),
		Model:        model,
		DatabaseURL:  getEnv("DATABASE_URL", ""),
	}

	err := validateConfig(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Model == "" {
		return ErrMissingModel
	}

	if cfg.WebsocketURL == "" {
		return ErrMissingWebsocketURL
	}

	if cfg.DatabaseURL == "" {
		return ErrMissingDatabaseURL
	}

	if cfg.JWTSecret == "" {
		return ErrMissingJWTSecret
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

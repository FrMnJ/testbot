package main

import (
	"fmt"
	"os"
)

var (
	ErrMissingModel        error = fmt.Errorf("MODEL is required (or set GENKIT_MODEL)")
	ErrMissingWebsocketURL error = fmt.Errorf("WEBSOCKET_URL is required")
	ErrMissingUserJWT      error = fmt.Errorf("USER_JWT is required")
)

type Config struct {
	UserJWT      string
	WebsocketURL string
	Model        string
}

var cfg *Config

func LoadConfig() (*Config, error) {
	model := getEnv("MODEL", "")
	if model == "" {
		model = getEnv("GENKIT_MODEL", "")
	}

	cfg := &Config{
		UserJWT:      getEnv("USER_JWT", ""),
		WebsocketURL: getEnv("WEBSOCKET_URL", ""),
		Model:        model,
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

	if cfg.UserJWT == "" {
		return ErrMissingUserJWT
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

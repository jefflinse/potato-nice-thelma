package config

import "os"

// Config holds the application configuration.
type Config struct {
	Port string
}

// Load reads configuration from environment variables and returns a populated
// Config. It returns an error if any required variables are missing.
func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port: port,
	}, nil
}

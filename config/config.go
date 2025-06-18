package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port        string
	Mode        string
	AppName     string
	AppVersion  string
	ReadTimeout time.Duration
	Database    struct {
		URL string
	}
	GoogleCloud struct {
		ProjectID       string
		Location        string
		Catalog         string
		CredentialsFile string
	}
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	// Create config with default values
	config := &Config{
		Port:        getEnv("PORT", "8080"),
		Mode:        getEnv("GIN_MODE", "debug"),
		AppName:     getEnv("APP_NAME", "geosearch-poc"),
		AppVersion:  getEnv("APP_VERSION", "1.0.0"),
		ReadTimeout: 10 * time.Second,
	}

	// Load database configuration
	config.Database.URL = getEnv("DATABASE_URL", "")

	// Load Google Cloud configuration
	config.GoogleCloud.ProjectID = getEnv("GOOGLE_CLOUD_PROJECT_ID", "")
	config.GoogleCloud.Location = getEnv("GOOGLE_CLOUD_LOCATION", "")
	config.GoogleCloud.Catalog = getEnv("GOOGLE_CLOUD_CATALOG", "")
	config.GoogleCloud.CredentialsFile = getEnv("GOOGLE_CLOUD_CREDENTIALS_FILE", "")

	// Validate required configuration
	if config.Database.URL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if config.GoogleCloud.ProjectID == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_PROJECT_ID is required")
	}

	if config.GoogleCloud.Location == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_LOCATION is required")
	}

	if config.GoogleCloud.Catalog == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_CATALOG is required")
	}

	if config.GoogleCloud.CredentialsFile == "" {
		return nil, fmt.Errorf("GOOGLE_CLOUD_CREDENTIALS_FILE is required")
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

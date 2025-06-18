package config

import (
	"fmt"
	"os"
	"strconv"
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
	H3 struct {
		Resolution int
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

	// Load H3 configuration
	config.H3.Resolution = getEnvInt("H3_RESOLUTION", 9)

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

// getEnvInt gets an environment variable as an integer or returns a default value
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

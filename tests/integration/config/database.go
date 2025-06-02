package config

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDatabaseConfig holds the test database configuration
type TestDatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	PoolSize int
}

// NewTestDatabaseConfig creates a new test database configuration
func NewTestDatabaseConfig() *TestDatabaseConfig {
	port, _ := strconv.Atoi(getEnv("TEST_DB_PORT", "5432"))
	poolSize, _ := strconv.Atoi(getEnv("TEST_DB_POOL_SIZE", "5"))

	return &TestDatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     port,
		User:     getEnv("TEST_DB_USER", "postgres"),
		Password: getEnv("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnv("TEST_DB_NAME", "geosearch_test"),
		PoolSize: poolSize,
	}
}

// GetDSN returns the database connection string
func (c *TestDatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
	)
}

// NewPool creates a new database connection pool
func (c *TestDatabaseConfig) NewPool() (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(c.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	config.MaxConns = int32(c.PoolSize)
	config.MinConns = 1
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	return pool, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

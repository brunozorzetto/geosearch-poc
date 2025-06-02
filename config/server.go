package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxHeaderBytes  int
	ShutdownTimeout time.Duration
}

// NewServerConfig creates a new server configuration from environment variables
func NewServerConfig() *ServerConfig {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	readTimeout, _ := strconv.Atoi(getEnv("SERVER_READ_TIMEOUT", "5"))
	writeTimeout, _ := strconv.Atoi(getEnv("SERVER_WRITE_TIMEOUT", "10"))
	maxHeaderBytes, _ := strconv.Atoi(getEnv("SERVER_MAX_HEADER_BYTES", "1048576"))
	shutdownTimeout, _ := strconv.Atoi(getEnv("SERVER_SHUTDOWN_TIMEOUT", "10"))

	return &ServerConfig{
		Port:            port,
		ReadTimeout:     time.Duration(readTimeout) * time.Second,
		WriteTimeout:    time.Duration(writeTimeout) * time.Second,
		MaxHeaderBytes:  maxHeaderBytes,
		ShutdownTimeout: time.Duration(shutdownTimeout) * time.Second,
	}
}

// GetAddress returns the server address
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}

// SetupGinMode sets up the Gin mode based on environment
func SetupGinMode() {
	mode := getEnv("GIN_MODE", "release")
	gin.SetMode(mode)
}

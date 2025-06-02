package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"geosearch-poc/handlers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port        string
	Mode        string
	AppName     string
	AppVersion  string
	ReadTimeout time.Duration
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	// Carrega o arquivo .env se ele existir
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	// Define valores padrão
	config := &Config{
		Port:        getEnv("PORT", "8080"),
		Mode:        getEnv("GIN_MODE", "debug"),
		AppName:     getEnv("APP_NAME", "geosearch-poc"),
		AppVersion:  getEnv("APP_VERSION", "1.0.0"),
		ReadTimeout: 10 * time.Second,
	}

	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	// Carrega configurações
	config := loadConfig()

	// Configura o modo do Gin
	gin.SetMode(config.Mode)

	// Cria uma nova instância do Gin com configurações personalizadas
	r := gin.New()

	// Middleware global
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Adiciona informações da aplicação ao contexto
	r.Use(func(c *gin.Context) {
		c.Set("app_name", config.AppName)
		c.Set("app_version", config.AppVersion)
		c.Next()
	})

	// Registra rotas
	r.GET("/health", handlers.HealthCheck)

	// Configura o servidor
	serverAddr := fmt.Sprintf(":%s", config.Port)
	log.Printf("Starting %s v%s on %s", config.AppName, config.AppVersion, serverAddr)

	// Inicia o servidor
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

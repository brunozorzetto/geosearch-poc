package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"geosearch-poc/config"
	"geosearch-poc/handlers"
	"geosearch-poc/repository/postgres"
	"geosearch-poc/service"

	h3 "github.com/uber/h3-go/v3"

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
	appConfig := loadConfig()

	// Configura o modo do Gin
	gin.SetMode(appConfig.Mode)

	// Cria uma nova instância do Gin com configurações personalizadas
	r := gin.New()

	// Middleware global
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Adiciona informações da aplicação ao contexto
	r.Use(func(c *gin.Context) {
		c.Set("app_name", appConfig.AppName)
		c.Set("app_version", appConfig.AppVersion)
		c.Next()
	})

	// Load database configuration
	dbConfig := config.NewDatabaseConfig()

	// Create database connection
	pool, err := dbConfig.NewPool()
	if err != nil {
		log.Fatalf("Failed to create database pool: %v", err)
	}
	defer pool.Close()

	// Initialize repositories
	storeRepo := postgres.NewStoreRepository(pool)
	productRepo := postgres.NewProductRepository(pool)

	// Initialize H3 indexer
	h3Indexer := h3.NewIndexer(9) // Resolution 9 for ~1km cells

	// Initialize services
	searchService := service.NewSearchService(storeRepo, productRepo, h3Indexer, 10)

	// Initialize handlers
	searchHandler := handlers.NewSearchHandler(searchService)

	// Registra rotas
	r.GET("/health", handlers.HealthCheck)

	// API routes
	api := r.Group("/api/v1")
	{
		// Search routes
		search := api.Group("/search")
		{
			search.GET("/stores", searchHandler.Search)
		}
	}

	// Configura o servidor
	serverAddr := fmt.Sprintf(":%s", appConfig.Port)
	log.Printf("Starting %s v%s on %s", appConfig.AppName, appConfig.AppVersion, serverAddr)

	// Inicia o servidor
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

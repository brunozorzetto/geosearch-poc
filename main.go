package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"geosearch-poc/config"
	"geosearch-poc/handlers"
	"geosearch-poc/pkg/h3"
	"geosearch-poc/repository/postgres"
	"geosearch-poc/service"
	"geosearch-poc/service/vertexai"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Load configuration
	appConfig, err := config.Load()
	if err != nil {
		log.Fatal("Error loading configuration:", err)
	}

	// Initialize database connection
	db, err := pgxpool.New(context.Background(), appConfig.Database.URL)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Initialize repositories
	storeRepo := postgres.NewStoreRepository(db)
	productRepo := postgres.NewProductRepository(db)

	// Initialize H3 indexer
	h3Indexer := h3.NewIndexer(9) // Resolution 9 for ~1km cells

	// Initialize Vertex AI client
	vertexAIClient, err := vertexai.NewClient(
		appConfig.GoogleCloud.ProjectID,
		appConfig.GoogleCloud.Location,
		appConfig.GoogleCloud.Catalog,
		appConfig.GoogleCloud.CredentialsFile,
	)
	if err != nil {
		log.Fatal("Error initializing Vertex AI client:", err)
	}
	defer vertexAIClient.Close()

	// Initialize services
	storeSearchService := service.NewStoreSearchService(storeRepo, h3Indexer)
	productSearchService := service.NewProductSearchService(productRepo, storeRepo, h3Indexer, vertexAIClient)

	// Initialize handlers
	searchHandler := handlers.NewSearchHandler(storeSearchService)
	storeHandler := handlers.NewStoreHandler(storeRepo, h3Indexer)
	productHandler := handlers.NewProductHandler(productRepo, productSearchService)

	// Initialize router
	router := gin.Default()

	// Register routes
	api := router.Group("/api/v1")
	{
		// Search routes
		api.GET("/search", searchHandler.Search)

		// Store routes
		stores := api.Group("/stores")
		{
			stores.POST("", storeHandler.Create)
			stores.GET("/:id", storeHandler.GetByID)
			stores.PUT("/:id", storeHandler.Update)
			stores.DELETE("/:id", storeHandler.Delete)
		}

		// Product routes
		products := api.Group("/products")
		{
			products.POST("", productHandler.Create)
			products.GET("/:id", productHandler.GetByID)
			products.PUT("/:id", productHandler.Update)
			products.DELETE("/:id", productHandler.Delete)
			products.GET("/search", productHandler.Search)
		}
	}

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

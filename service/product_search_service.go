package service

import (
	"context"
	"fmt"
	"strings"

	"geosearch-poc/domain"
	"geosearch-poc/repository"
	"geosearch-poc/service/vertexai"

	"github.com/google/uuid"
)

type ProductSearchService struct {
	productRepo repository.ProductRepository
	storeRepo   repository.StoreRepository
	h3Indexer   H3Indexer
	vertexAI    vertexai.VertexAIClientInterface
}

func NewProductSearchService(
	productRepo repository.ProductRepository,
	storeRepo repository.StoreRepository,
	h3Indexer H3Indexer,
	vertexAI vertexai.VertexAIClientInterface,
) *ProductSearchService {
	return &ProductSearchService{
		productRepo: productRepo,
		storeRepo:   storeRepo,
		h3Indexer:   h3Indexer,
		vertexAI:    vertexAI,
	}
}

// SearchProducts performs a product search using H3 for geographic filtering and Vertex AI for semantic search
func (s *ProductSearchService) SearchProducts(ctx context.Context, params *domain.ProductSearchParams) (*domain.ProductSearchResponse, error) {
	// 1. Get H3 cells within radius
	cells, err := s.h3Indexer.GetCellsInRadius(params.Latitude, params.Longitude, params.Radius)
	if err != nil {
		return nil, fmt.Errorf("failed to get H3 cells: %w", err)
	}

	// 2. Get stores in those cells
	stores, err := s.storeRepo.GetByH3Cells(ctx, cells)
	if err != nil {
		return nil, fmt.Errorf("failed to get stores: %w", err)
	}

	if len(stores) == 0 {
		return &domain.ProductSearchResponse{
			Products:      []domain.Product{},
			TotalCount:    0,
			NextPageToken: "",
		}, nil
	}

	// 3. Create filter for Vertex AI search
	filters := []string{
		fmt.Sprintf("attributes.store_id IN (%s)", strings.Join(getStoreIDs(stores), ",")),
	}

	if params.Category != "" {
		filters = append(filters, fmt.Sprintf("categories = '%s'", params.Category))
	}

	if params.Brand != "" {
		filters = append(filters, fmt.Sprintf("brands = '%s'", params.Brand))
	}

	if params.MinPrice > 0 {
		filters = append(filters, fmt.Sprintf("price_info.price >= %f", params.MinPrice))
	}

	if params.MaxPrice > 0 {
		filters = append(filters, fmt.Sprintf("price_info.price <= %f", params.MaxPrice))
	}

	// 4. Execute search with Vertex AI
	searchParams := vertexai.SearchParams{
		Query:     params.Query,
		Filter:    strings.Join(filters, " AND "),
		PageSize:  params.PageSize,
		PageToken: params.PageToken,
	}

	results, nextPageToken, err := s.vertexAI.Search(ctx, searchParams)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	// 5. Convert results to domain products
	products := make([]domain.Product, len(results))
	for i, result := range results {
		products[i] = convertToProduct(result)
	}

	return &domain.ProductSearchResponse{
		Products:      products,
		TotalCount:    int64(len(products)),
		NextPageToken: nextPageToken,
	}, nil
}

// getStoreIDs extracts store IDs from store list
func getStoreIDs(stores []domain.Store) []string {
	ids := make([]string, len(stores))
	for i, store := range stores {
		ids[i] = fmt.Sprintf("'%s'", store.ID.String())
	}
	return ids
}

// convertToProduct converts a Vertex AI search result to a domain product
func convertToProduct(result vertexai.SearchResult) domain.Product {
	// Parse UUIDs
	productID, _ := uuid.Parse(result.ID)
	storeID, _ := uuid.Parse(result.Attributes["store_id"])

	return domain.Product{
		ID:          productID,
		StoreID:     storeID,
		H3Index:     result.Attributes["h3_index"],
		Name:        result.Title,
		Description: result.Description,
		Price:       result.PriceInfo.Price,
		Category:    result.Categories[0], // Assuming first category is primary
		Brand:       result.Brands[0],     // Assuming first brand is primary
		Stock:       int(result.Availability.Stock),
		Images:      result.Images,
		CreatedAt:   result.CreatedAt,
		UpdatedAt:   result.UpdatedAt,
	}
}

// GetVertexAIClient returns the Vertex AI client for external operations
func (s *ProductSearchService) GetVertexAIClient() vertexai.VertexAIClientInterface {
	return s.vertexAI
}

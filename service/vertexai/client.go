package vertexai

import (
	"context"
	"fmt"
	"time"

	"geosearch-poc/domain"

	retail "cloud.google.com/go/retail/apiv2"
	"cloud.google.com/go/retail/apiv2/retailpb"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// Client represents a Vertex AI Retail Search client
type Client struct {
	projectID     string
	location      string
	catalog       string
	searchClient  *retail.SearchClient
	productClient *retail.ProductClient
}

// SearchParams represents the parameters for a product search
type SearchParams struct {
	Query     string
	Filter    string
	PageSize  int
	PageToken string
}

// SearchResult represents a product search result
type SearchResult struct {
	ID           string
	Title        string
	Description  string
	PriceInfo    PriceInfo
	Categories   []string
	Brands       []string
	Availability Availability
	Attributes   map[string]string
	Images       []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// PriceInfo represents product price information
type PriceInfo struct {
	Price         float64
	Currency      string
	OriginalPrice float64
}

// Availability represents product availability information
type Availability struct {
	Stock     int64
	Available bool
}

// NewClient creates a new Vertex AI Retail Search client
func NewClient(projectID, location, catalog string, credentialsFile string) (*Client, error) {
	ctx := context.Background()

	// Create search client
	searchClient, err := retail.NewSearchClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create search client: %w", err)
	}

	// Create product client
	productClient, err := retail.NewProductClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create product client: %w", err)
	}

	return &Client{
		projectID:     projectID,
		location:      location,
		catalog:       catalog,
		searchClient:  searchClient,
		productClient: productClient,
	}, nil
}

// Search performs a product search
func (c *Client) Search(ctx context.Context, params SearchParams) ([]SearchResult, string, error) {
	// Create search request
	req := &retailpb.SearchRequest{
		Placement: fmt.Sprintf("projects/%s/locations/%s/catalogs/%s/placements/default_search", c.projectID, c.location, c.catalog),
		Query:     params.Query,
		Filter:    params.Filter,
		PageSize:  int32(params.PageSize),
	}

	if params.PageToken != "" {
		req.PageToken = params.PageToken
	}

	// Execute search
	iter := c.searchClient.Search(ctx, req)
	results := make([]SearchResult, 0)

	for {
		result, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to get next result: %w", err)
		}
		results = append(results, convertSearchResult(result))
	}

	return results, iter.PageInfo().Token, nil
}

// convertSearchResult converts a Retail Search result to our domain model
func convertSearchResult(result *retailpb.SearchResponse_SearchResult) SearchResult {
	// Extract attributes
	attributes := make(map[string]string)
	for k, v := range result.Product.Attributes {
		if text := v.GetText(); len(text) > 0 {
			attributes[k] = text[0] // Get first text value
		}
	}

	// Extract price info
	priceInfo := PriceInfo{}
	if result.Product.PriceInfo != nil {
		priceInfo.Price = float64(result.Product.PriceInfo.Price)
		priceInfo.Currency = result.Product.PriceInfo.CurrencyCode
		priceInfo.OriginalPrice = float64(result.Product.PriceInfo.OriginalPrice)
	}

	// Extract availability - handle as enum
	availability := Availability{}
	switch result.Product.Availability {
	case retailpb.Product_IN_STOCK:
		availability.Stock = 1
		availability.Available = true
	case retailpb.Product_OUT_OF_STOCK:
		availability.Stock = 0
		availability.Available = false
	default:
		availability.Stock = 0
		availability.Available = false
	}

	// Convert images
	var images []string
	for _, img := range result.Product.Images {
		images = append(images, img.Uri)
	}

	// Get timestamps
	var createdAt, updatedAt time.Time
	if result.Product.PublishTime != nil {
		createdAt = result.Product.PublishTime.AsTime()
		updatedAt = result.Product.PublishTime.AsTime()
	}

	return SearchResult{
		ID:           result.Product.Id,
		Title:        result.Product.Title,
		Description:  result.Product.Description,
		PriceInfo:    priceInfo,
		Categories:   result.Product.Categories,
		Brands:       result.Product.Brands,
		Availability: availability,
		Attributes:   attributes,
		Images:       images,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}

// Close closes the client connections
func (c *Client) Close() error {
	if err := c.searchClient.Close(); err != nil {
		return fmt.Errorf("failed to close search client: %w", err)
	}
	if err := c.productClient.Close(); err != nil {
		return fmt.Errorf("failed to close product client: %w", err)
	}
	return nil
}

func (c *Client) CreateProduct(ctx context.Context, product *retailpb.Product) error {
	req := &retailpb.CreateProductRequest{
		Parent:    fmt.Sprintf("projects/%s/locations/%s/catalogs/%s/branches/default_branch", c.projectID, c.location, c.catalog),
		Product:   product,
		ProductId: product.Id,
	}

	_, err := c.productClient.CreateProduct(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (c *Client) UpdateProduct(ctx context.Context, product *retailpb.Product) error {
	req := &retailpb.UpdateProductRequest{
		Product: product,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"title", "description", "price_info", "categories", "brands",
				"gtin", "availability", "attributes", "images",
			},
		},
	}

	_, err := c.productClient.UpdateProduct(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

func (c *Client) DeleteProduct(ctx context.Context, productID string) error {
	req := &retailpb.DeleteProductRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/catalogs/%s/branches/default_branch/products/%s",
			c.projectID, c.location, c.catalog, productID),
	}

	err := c.productClient.DeleteProduct(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// Helper function to convert a search result to a domain product
func convertToDomainProduct(result *retailpb.SearchResponse_SearchResult) *domain.Product {
	if result == nil || result.Product == nil {
		return nil
	}

	p := result.Product

	var price float64
	if p.PriceInfo != nil {
		price = float64(p.PriceInfo.Price)
	}

	var stock int
	if p.AvailableQuantity != nil {
		stock = int(p.AvailableQuantity.GetValue())
	}

	var images []string
	if len(p.Images) > 0 {
		images = make([]string, len(p.Images))
		for i, img := range p.Images {
			images[i] = img.Uri
		}
	}

	var category, brand string
	if len(p.Categories) > 0 {
		category = p.Categories[0]
	}
	if len(p.Brands) > 0 {
		brand = p.Brands[0]
	}

	var createdAt, updatedAt time.Time
	if p.PublishTime != nil {
		createdAt = p.PublishTime.AsTime()
		updatedAt = p.PublishTime.AsTime()
	}

	product := &domain.Product{
		ID:          uuid.MustParse(result.Id),
		Name:        p.Title,
		Description: p.Description,
		Price:       price,
		Category:    category,
		Brand:       brand,
		SKU:         p.Gtin,
		Stock:       stock,
		H3Index:     "", // This should be set by the caller based on the store's location
		Images:      images,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	return product
}

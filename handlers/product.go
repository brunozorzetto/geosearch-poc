package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"geosearch-poc/domain"
	"geosearch-poc/repository"
	"geosearch-poc/service"

	retail "cloud.google.com/go/retail/apiv2/retailpb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProductHandler handles product-related requests
type ProductHandler struct {
	productRepo      repository.ProductRepository
	productSearchSvc *service.ProductSearchService
}

// NewProductHandler creates a new product handler instance
func NewProductHandler(
	productRepo repository.ProductRepository,
	productSearchSvc *service.ProductSearchService,
) *ProductHandler {
	return &ProductHandler{
		productRepo:      productRepo,
		productSearchSvc: productSearchSvc,
	}
}

// Create handles product creation
func (h *ProductHandler) Create(c *gin.Context) {
	var req struct {
		StoreID     string   `json:"store_id" binding:"required"`
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Price       float64  `json:"price" binding:"required"`
		Category    string   `json:"category"`
		Brand       string   `json:"brand"`
		SKU         string   `json:"sku" binding:"required"`
		Stock       int      `json:"stock" binding:"required"`
		H3Index     string   `json:"h3_index"`
		Images      []string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	storeID, err := uuid.Parse(req.StoreID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store_id"})
		return
	}

	product := domain.NewProduct(
		storeID,
		req.Name,
		req.Description,
		req.Price,
		req.Category,
		req.Brand,
		req.SKU,
		req.Stock,
		req.H3Index,
		req.Images,
	)

	if err := h.productRepo.Create(c.Request.Context(), product); err != nil {
		// Verificar se é um erro de constraint de foreign key
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "store not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}

	// Index in Retail Search (async to avoid blocking the response)
	go func() {
		ctx := context.Background()
		if err := h.indexProductInRetailSearch(ctx, product); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to index product %s in Retail Search: %v", product.ID, err)
		} else {
			log.Printf("Successfully indexed product %s in Retail Search", product.ID)
		}
	}()

	c.JSON(http.StatusCreated, product)
}

// GetByID handles retrieving a product by ID
func (h *ProductHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	product, err := h.productRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetByStoreID handles retrieving all products for a store
func (h *ProductHandler) GetByStoreID(c *gin.Context) {
	storeID, err := uuid.Parse(c.Param("store_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	products, err := h.productRepo.GetByStoreID(c.Request.Context(), storeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// Update handles product updates
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var req struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Price       float64  `json:"price"`
		Category    string   `json:"category"`
		Brand       string   `json:"brand"`
		SKU         string   `json:"sku"`
		Stock       int      `json:"stock"`
		H3Index     string   `json:"h3_index"`
		Images      []string `json:"images"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	product, err := h.productRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	product.Update(
		req.Name,
		req.Description,
		req.Price,
		req.Category,
		req.Brand,
		req.SKU,
		req.Stock,
		req.H3Index,
		req.Images,
	)

	if err := h.productRepo.Update(c.Request.Context(), product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}

	// Update in Retail Search (async to avoid blocking the response)
	go func() {
		ctx := context.Background()
		if err := h.updateProductInRetailSearch(ctx, product); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to update product %s in Retail Search: %v", product.ID, err)
		} else {
			log.Printf("Successfully updated product %s in Retail Search", product.ID)
		}
	}()

	c.JSON(http.StatusOK, product)
}

// Delete handles product deletion
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	if err := h.productRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		return
	}

	// Remove from Retail Search (async to avoid blocking the response)
	go func() {
		ctx := context.Background()
		if err := h.removeProductFromRetailSearch(ctx, id.String()); err != nil {
			// Log error but don't fail the request
			log.Printf("Failed to remove product %s from Retail Search: %v", id, err)
		} else {
			log.Printf("Successfully removed product %s from Retail Search", id)
		}
	}()

	c.Status(http.StatusNoContent)
}

// Search handles product search using Vertex AI Search
func (h *ProductHandler) Search(c *gin.Context) {
	// Parse query parameters
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
		return
	}

	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
		return
	}

	radius, err := strconv.ParseFloat(c.Query("radius"), 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid radius"})
		return
	}

	// Optional parameters
	pageSize := 10 // default page size
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if parsedPageSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedPageSize > 0 {
			pageSize = parsedPageSize
		}
	}

	params := &domain.ProductSearchParams{
		Query:     query,
		Latitude:  lat,
		Longitude: lng,
		Radius:    radius,
		PageSize:  pageSize,
		PageToken: c.Query("page_token"),
		Category:  c.Query("category"),
		Brand:     c.Query("brand"),
	}

	// Parse price range if provided
	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			params.MinPrice = minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			params.MaxPrice = maxPrice
		}
	}

	// Perform search
	result, err := h.productSearchSvc.SearchProducts(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search products"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// indexProductInRetailSearch indexes a product in Vertex AI Retail Search
func (h *ProductHandler) indexProductInRetailSearch(ctx context.Context, product *domain.Product) error {
	// Convert domain product to Retail Search product
	retailProduct := h.convertToRetailProduct(product)

	// Create product in Retail Search
	return h.productSearchSvc.GetVertexAIClient().CreateProduct(ctx, retailProduct)
}

// updateProductInRetailSearch updates a product in Vertex AI Retail Search
func (h *ProductHandler) updateProductInRetailSearch(ctx context.Context, product *domain.Product) error {
	// Convert domain product to Retail Search product
	retailProduct := h.convertToRetailProduct(product)

	// Update product in Retail Search
	return h.productSearchSvc.GetVertexAIClient().UpdateProduct(ctx, retailProduct)
}

// removeProductFromRetailSearch removes a product from Vertex AI Retail Search
func (h *ProductHandler) removeProductFromRetailSearch(ctx context.Context, productID string) error {
	return h.productSearchSvc.GetVertexAIClient().DeleteProduct(ctx, productID)
}

// convertToRetailProduct converts a domain product to Retail Search format
func (h *ProductHandler) convertToRetailProduct(product *domain.Product) *retail.Product {
	retailProduct := &retail.Product{
		Id:          product.ID.String(),
		Title:       product.Name,
		Description: product.Description,
		PriceInfo: &retail.PriceInfo{
			Price:        float32(product.Price), // Convert to float32
			CurrencyCode: "BRL",
		},
		Categories: []string{product.Category},
		Brands:     []string{product.Brand},
		Gtin:       product.SKU,
		Availability: func() retail.Product_Availability {
			if product.Stock > 0 {
				return retail.Product_IN_STOCK
			}
			return retail.Product_OUT_OF_STOCK
		}(),
		Attributes: map[string]*retail.CustomAttribute{
			"store_id": {
				Text: []string{product.StoreID.String()},
			},
			"h3_index": {
				Text: []string{product.H3Index},
			},
		},
	}

	// Add images if available
	if len(product.Images) > 0 {
		retailProduct.Images = make([]*retail.Image, len(product.Images))
		for i, imgURL := range product.Images {
			retailProduct.Images[i] = &retail.Image{
				Uri: imgURL,
			}
		}
	}

	return retailProduct
}

package handlers

import (
	"net/http"
	"strconv"

	"geosearch-poc/domain"
	"geosearch-poc/service"

	"github.com/gin-gonic/gin"
)

// SearchHandler handles store search requests
type SearchHandler struct {
	searchService *service.SearchService
}

// NewSearchHandler creates a new search handler instance
func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{
		searchService: searchService,
	}
}

// Search handles the store search endpoint
func (h *SearchHandler) Search(c *gin.Context) {
	// Parse query parameters
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
	limit := 10 // default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // default offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Create search parameters
	params := domain.NewSearchParams(lat, lng, radius).
		WithPagination(limit, offset)

	// Perform search
	result, err := h.searchService.SearchStores(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search stores"})
		return
	}

	// Add delivery radius information to the response
	type StoreResponse struct {
		ID             int     `json:"id"`
		Name           string  `json:"name"`
		Latitude       float64 `json:"latitude"`
		Longitude      float64 `json:"longitude"`
		H3Index        string  `json:"h3_index"`
		DeliveryRadius int     `json:"delivery_radius"` // in meters
		Distance       float64 `json:"distance"`        // in kilometers
	}

	response := struct {
		Stores []StoreResponse `json:"stores"`
		Total  int             `json:"total"`
	}{
		Stores: make([]StoreResponse, len(result.Stores)),
		Total:  result.Total,
	}

	for i, store := range result.Stores {
		response.Stores[i] = StoreResponse{
			ID:             store.Store.ID,
			Name:           store.Store.Name,
			Latitude:       store.Store.Latitude,
			Longitude:      store.Store.Longitude,
			H3Index:        store.Store.H3Index,
			DeliveryRadius: store.Store.DeliveryRadius,
			Distance:       store.Distance,
		}
	}

	c.JSON(http.StatusOK, response)
}

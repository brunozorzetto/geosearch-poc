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
	storeSearchService *service.StoreSearchService
}

// NewSearchHandler creates a new search handler instance
func NewSearchHandler(storeSearchService *service.StoreSearchService) *SearchHandler {
	return &SearchHandler{
		storeSearchService: storeSearchService,
	}
}

// Search handles store search requests
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

	// Create search parameters
	params := domain.StoreSearchParams{
		Latitude:  lat,
		Longitude: lng,
		Radius:    radius,
	}

	// Execute search
	stores, err := h.storeSearchService.Search(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return results
	c.JSON(http.StatusOK, gin.H{
		"stores": stores,
		"total":  len(stores),
	})
}

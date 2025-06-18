package handlers

import (
	"geosearch-poc/domain"
	"geosearch-poc/pkg/h3"
	"geosearch-poc/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StoreHandler struct {
	storeRepo repository.StoreRepository
	indexer   *h3.Indexer
}

func NewStoreHandler(storeRepo repository.StoreRepository, indexer *h3.Indexer) *StoreHandler {
	return &StoreHandler{
		storeRepo: storeRepo,
		indexer:   indexer,
	}
}

func (h *StoreHandler) Create(c *gin.Context) {
	var req struct {
		Name           string  `json:"name"`
		CategoryID     string  `json:"category_id"` // optional for now
		Latitude       float64 `json:"latitude"`
		Longitude      float64 `json:"longitude"`
		Address        string  `json:"address"`
		DeliveryRadius float64 `json:"delivery_radius"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Validate required fields
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if req.Latitude == 0 && req.Longitude == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "latitude and longitude are required"})
		return
	}

	// For now, use a dummy category if not provided
	catID := uuid.Nil
	if req.CategoryID != "" {
		var err error
		catID, err = uuid.Parse(req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
	}

	// Calculate H3 index from latitude and longitude
	h3Index, err := h.indexer.GetCellFromLatLng(req.Latitude, req.Longitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate H3 index"})
		return
	}

	store := &domain.Store{
		ID:             uuid.New(),
		Name:           req.Name,
		CategoryID:     catID,
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		H3Index:        h3Index,
		Address:        req.Address,
		DeliveryRadius: req.DeliveryRadius,
	}

	if err := h.storeRepo.Create(c.Request.Context(), store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create store"})
		return
	}

	c.JSON(http.StatusCreated, store)
}

// GetByID handles retrieving a store by ID
func (h *StoreHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	store, err := h.storeRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	c.JSON(http.StatusOK, store)
}

// Update handles store updates
func (h *StoreHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	var req struct {
		Name           string  `json:"name"`
		CategoryID     string  `json:"category_id"`
		Latitude       float64 `json:"latitude"`
		Longitude      float64 `json:"longitude"`
		Address        string  `json:"address"`
		DeliveryRadius float64 `json:"delivery_radius"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	store, err := h.storeRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	// Calculate new H3 index if coordinates changed
	h3Index := store.H3Index
	if req.Latitude != 0 && req.Longitude != 0 {
		h3Index, err = h.indexer.GetCellFromLatLng(req.Latitude, req.Longitude)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate new H3 index"})
			return
		}
	}

	// Parse category ID if provided
	catID := store.CategoryID
	if req.CategoryID != "" {
		catID, err = uuid.Parse(req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
	}

	store.Update(
		req.Name,
		catID,
		req.Latitude,
		req.Longitude,
		h3Index,
		req.Address,
		req.DeliveryRadius,
	)

	if err := h.storeRepo.Update(c.Request.Context(), store); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update store"})
		return
	}

	c.JSON(http.StatusOK, store)
}

// Delete handles store deletion
func (h *StoreHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid store id"})
		return
	}

	// Check if store exists
	_, err = h.storeRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "store not found"})
		return
	}

	if err := h.storeRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete store"})
		return
	}

	c.Status(http.StatusNoContent)
}

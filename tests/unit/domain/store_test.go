package domain_test

import (
	"testing"

	"geosearch-poc/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	tests := []struct {
		name           string
		storeName      string
		categoryID     uuid.UUID
		latitude       float64
		longitude      float64
		address        string
		deliveryRadius float64
		checkResult    func(t *testing.T, store *domain.Store)
	}{
		{
			name:           "should create store with all fields",
			storeName:      "Test Store",
			categoryID:     uuid.New(),
			latitude:       -23.550520,
			longitude:      -46.633308,
			address:        "Rua Teste, 123",
			deliveryRadius: 5.0,
			checkResult: func(t *testing.T, store *domain.Store) {
				assert.NotEqual(t, uuid.Nil, store.ID)
				assert.Equal(t, "Test Store", store.Name)
				assert.NotEqual(t, uuid.Nil, store.CategoryID)
				assert.Equal(t, -23.550520, store.Latitude)
				assert.Equal(t, -46.633308, store.Longitude)
				assert.Equal(t, "Rua Teste, 123", store.Address)
				assert.Equal(t, 5.0, store.DeliveryRadius)
				assert.NotEmpty(t, store.H3Index)
				assert.False(t, store.CreatedAt.IsZero())
				assert.False(t, store.UpdatedAt.IsZero())
			},
		},
		{
			name:           "should create store with minimal fields",
			storeName:      "Minimal Store",
			categoryID:     uuid.Nil,
			latitude:       -23.550520,
			longitude:      -46.633308,
			address:        "Rua Teste, 456",
			deliveryRadius: 2.0,
			checkResult: func(t *testing.T, store *domain.Store) {
				assert.NotEqual(t, uuid.Nil, store.ID)
				assert.Equal(t, "Minimal Store", store.Name)
				assert.Equal(t, uuid.Nil, store.CategoryID)
				assert.Equal(t, -23.550520, store.Latitude)
				assert.Equal(t, -46.633308, store.Longitude)
				assert.Equal(t, "Rua Teste, 456", store.Address)
				assert.Equal(t, 2.0, store.DeliveryRadius)
				assert.NotEmpty(t, store.H3Index)
				assert.False(t, store.CreatedAt.IsZero())
				assert.False(t, store.UpdatedAt.IsZero())
			},
		},
		{
			name:           "should create store with zero delivery radius",
			storeName:      "Store with Zero Delivery",
			categoryID:     uuid.New(),
			latitude:       -23.550520,
			longitude:      -46.633308,
			address:        "Rua Teste, 789",
			deliveryRadius: 0.0,
			checkResult: func(t *testing.T, store *domain.Store) {
				assert.Equal(t, 0.0, store.DeliveryRadius)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create store
			store := domain.NewStore(
				tt.storeName,
				tt.categoryID,
				tt.latitude,
				tt.longitude,
				"8928308280fffff", // H3 index
				tt.address,
				tt.deliveryRadius,
			)

			// Run checks
			tt.checkResult(t, store)
		})
	}
}

func TestStore_Update(t *testing.T) {
	originalStore := domain.NewStore(
		"Original Store",
		uuid.New(),
		-23.550520,
		-46.633308,
		"8928308280fffff", // H3 index
		"Original Address",
		5.0,
	)

	originalCreatedAt := originalStore.CreatedAt
	originalUpdatedAt := originalStore.UpdatedAt

	tests := []struct {
		name        string
		updates     func(*domain.Store)
		checkResult func(t *testing.T, store *domain.Store)
	}{
		{
			name: "should update all fields",
			updates: func(s *domain.Store) {
				s.Update(
					"Updated Store",
					uuid.New(),
					-23.560520,
					-46.643308,
					"8928308281fffff", // Updated H3 index
					"Updated Address",
					7.0,
				)
			},
			checkResult: func(t *testing.T, store *domain.Store) {
				assert.Equal(t, "Updated Store", store.Name)
				assert.NotEqual(t, originalStore.CategoryID, store.CategoryID)
				assert.Equal(t, -23.560520, store.Latitude)
				assert.Equal(t, -46.643308, store.Longitude)
				assert.Equal(t, "Updated Address", store.Address)
				assert.Equal(t, 7.0, store.DeliveryRadius)
				assert.NotEmpty(t, store.H3Index)

				// Check timestamps
				assert.Equal(t, originalCreatedAt, store.CreatedAt)      // Should not change
				assert.True(t, store.UpdatedAt.After(originalUpdatedAt)) // Should be updated
			},
		},
		{
			name: "should update partial fields",
			updates: func(s *domain.Store) {
				s.Update(
					"Partially Updated Store",
					s.CategoryID, // Keep original
					s.Latitude,   // Keep original
					s.Longitude,  // Keep original
					s.H3Index,    // Keep original
					s.Address,    // Keep original
					10.0,         // Update delivery radius
				)
			},
			checkResult: func(t *testing.T, store *domain.Store) {
				assert.Equal(t, "Partially Updated Store", store.Name)
				assert.Equal(t, originalStore.CategoryID, store.CategoryID)
				assert.Equal(t, originalStore.Latitude, store.Latitude)
				assert.Equal(t, originalStore.Longitude, store.Longitude)
				assert.Equal(t, originalStore.Address, store.Address)
				assert.Equal(t, 10.0, store.DeliveryRadius)

				// Check timestamps
				assert.Equal(t, originalCreatedAt, store.CreatedAt)      // Should not change
				assert.True(t, store.UpdatedAt.After(originalUpdatedAt)) // Should be updated
			},
		},
		{
			name: "should update with zero values",
			updates: func(s *domain.Store) {
				s.Update(
					"Zero Updated Store",
					uuid.Nil,
					0.0,
					0.0,
					"", // Empty H3 index
					"",
					0.0,
				)
			},
			checkResult: func(t *testing.T, store *domain.Store) {
				assert.Equal(t, "Zero Updated Store", store.Name)
				assert.Equal(t, uuid.Nil, store.CategoryID)
				assert.Equal(t, 0.0, store.Latitude)
				assert.Equal(t, 0.0, store.Longitude)
				assert.Equal(t, "", store.Address)
				assert.Equal(t, 0.0, store.DeliveryRadius)

				// Check timestamps
				assert.Equal(t, originalCreatedAt, store.CreatedAt)      // Should not change
				assert.True(t, store.UpdatedAt.After(originalUpdatedAt)) // Should be updated
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the original store
			store := &domain.Store{
				ID:             originalStore.ID,
				Name:           originalStore.Name,
				CategoryID:     originalStore.CategoryID,
				Latitude:       originalStore.Latitude,
				Longitude:      originalStore.Longitude,
				H3Index:        originalStore.H3Index,
				Address:        originalStore.Address,
				DeliveryRadius: originalStore.DeliveryRadius,
				CreatedAt:      originalStore.CreatedAt,
				UpdatedAt:      originalStore.UpdatedAt,
			}

			// Apply updates
			tt.updates(store)

			// Run checks
			tt.checkResult(t, store)
		})
	}
}

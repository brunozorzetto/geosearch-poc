package domain

import (
	"time"

	"github.com/google/uuid"
)

// Store represents a store in the system
type Store struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	CategoryID     uuid.UUID `json:"category_id"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	H3Index        string    `json:"h3_index"`
	Address        string    `json:"address"`
	DeliveryRadius float64   `json:"delivery_radius"`
	Distance       float64   `json:"distance"` // Distance from search point in kilometers
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// StoreSearchParams represents the parameters for a store search
type StoreSearchParams struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius"`
}

// NewStore creates a new store
func NewStore(
	name string,
	categoryID uuid.UUID,
	latitude float64,
	longitude float64,
	h3Index string,
	address string,
	deliveryRadius float64,
) *Store {
	now := time.Now()
	return &Store{
		ID:             uuid.New(),
		Name:           name,
		CategoryID:     categoryID,
		Latitude:       latitude,
		Longitude:      longitude,
		H3Index:        h3Index,
		Address:        address,
		DeliveryRadius: deliveryRadius,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Update updates the store information
func (s *Store) Update(
	name string,
	categoryID uuid.UUID,
	latitude float64,
	longitude float64,
	h3Index string,
	address string,
	deliveryRadius float64,
) {
	s.Name = name
	s.CategoryID = categoryID
	s.Latitude = latitude
	s.Longitude = longitude
	s.H3Index = h3Index
	s.Address = address
	s.DeliveryRadius = deliveryRadius
	s.UpdatedAt = time.Now()
}

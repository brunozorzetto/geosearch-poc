package domain

import (
	"time"
)

// Store represents a store in the system
type Store struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	H3Index        string    `json:"h3_index"`
	DeliveryRadius int       `json:"delivery_radius"` // Maximum delivery radius in meters
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// NewStore creates a new store instance
func NewStore(name string, latitude, longitude float64, h3Index string, deliveryRadius int) *Store {
	now := time.Now()
	return &Store{
		Name:           name,
		Latitude:       latitude,
		Longitude:      longitude,
		H3Index:        h3Index,
		DeliveryRadius: deliveryRadius,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Update updates the store information
func (s *Store) Update(name string, latitude, longitude float64, h3Index string, deliveryRadius int) {
	s.Name = name
	s.Latitude = latitude
	s.Longitude = longitude
	s.H3Index = h3Index
	s.DeliveryRadius = deliveryRadius
	s.UpdatedAt = time.Now()
}

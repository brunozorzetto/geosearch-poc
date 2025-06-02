package domain

import (
	"time"
)

// Store represents a physical store location
type Store struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	H3Index   string    `json:"h3_index"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewStore creates a new store instance
func NewStore(name, category string, latitude, longitude float64, address string) *Store {
	now := time.Now()
	return &Store{
		Name:      name,
		Category:  category,
		Latitude:  latitude,
		Longitude: longitude,
		Address:   address,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the store information
func (s *Store) Update(name, category string, latitude, longitude float64, address string) {
	s.Name = name
	s.Category = category
	s.Latitude = latitude
	s.Longitude = longitude
	s.Address = address
	s.UpdatedAt = time.Now()
}

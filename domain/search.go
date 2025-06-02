package domain

// SearchParams represents the parameters for a store search
type SearchParams struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Radius    float64 `json:"radius"` // in meters
	Category  string  `json:"category,omitempty"`
	Limit     int     `json:"limit,omitempty"`
	Offset    int     `json:"offset,omitempty"`
}

// SearchResponse represents the response for a store search
type SearchResponse struct {
	Stores []StoreWithProducts `json:"stores"`
	Total  int                 `json:"total"`
}

// StoreWithProducts represents a store with its products
type StoreWithProducts struct {
	Store    Store     `json:"store"`
	Products []Product `json:"products"`
	Distance float64   `json:"distance"` // in meters
}

// NewSearchParams creates a new search parameters instance with default values
func NewSearchParams(latitude, longitude, radius float64) *SearchParams {
	return &SearchParams{
		Latitude:  latitude,
		Longitude: longitude,
		Radius:    radius,
		Limit:     10, // default limit
		Offset:    0,  // default offset
	}
}

// WithCategory adds a category filter to the search parameters
func (sp *SearchParams) WithCategory(category string) *SearchParams {
	sp.Category = category
	return sp
}

// WithPagination adds pagination parameters to the search
func (sp *SearchParams) WithPagination(limit, offset int) *SearchParams {
	sp.Limit = limit
	sp.Offset = offset
	return sp
}

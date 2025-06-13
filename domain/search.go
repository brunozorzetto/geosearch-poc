package domain

// SearchParams represents the parameters for a store search
type SearchParams struct {
	Latitude  float64
	Longitude float64
	Radius    float64
	Limit     int
	Offset    int
}

// NewSearchParams creates a new SearchParams instance
func NewSearchParams(lat, lng, radius float64) *SearchParams {
	return &SearchParams{
		Latitude:  lat,
		Longitude: lng,
		Radius:    radius,
		Limit:     10,
		Offset:    0,
	}
}

// WithPagination sets the pagination parameters
func (p *SearchParams) WithPagination(limit, offset int) *SearchParams {
	p.Limit = limit
	p.Offset = offset
	return p
}

// StoreWithDistance represents a store with its distance from the search point
type StoreWithDistance struct {
	Store    Store
	Distance float64
}

// SearchResponse represents the response of a store search
type SearchResponse struct {
	Stores []StoreWithDistance
	Total  int
}

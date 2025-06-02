package h3

import (
	h3 "github.com/uber/h3-go/v3"
)

// Indexer handles H3 geospatial indexing operations
type Indexer struct {
	resolution int
}

// NewIndexer creates a new H3 indexer with the specified resolution
func NewIndexer(resolution int) *Indexer {
	return &Indexer{
		resolution: resolution,
	}
}

// LatLngToH3 converts latitude and longitude to H3 index
func (i *Indexer) LatLngToH3(lat, lng float64) string {
	index := h3.FromGeo(h3.GeoCoord{Latitude: lat, Longitude: lng}, i.resolution)
	return h3.ToString(index)
}

// GetNeighbors returns the H3 indexes of neighboring cells
func (i *Indexer) GetNeighbors(h3Index string) ([]string, error) {
	index := h3.FromString(h3Index) // This will panic if the string is invalid
	neighbors := h3.KRing(index, 1)
	result := make([]string, len(neighbors))
	for j, n := range neighbors {
		result[j] = h3.ToString(n)
	}
	return result, nil
}

// GetCellsInRadius returns all H3 cells within a specified radius (in meters)
func (i *Indexer) GetCellsInRadius(lat, lng float64, radiusMeters float64) ([]string, error) {
	center := h3.FromGeo(h3.GeoCoord{Latitude: lat, Longitude: lng}, i.resolution)
	k := int(radiusMeters / 1000)
	if k < 1 {
		k = 1
	}
	cells := h3.KRing(center, k)
	result := make([]string, len(cells))
	for j, n := range cells {
		result[j] = h3.ToString(n)
	}
	return result, nil
}

// getResolutionForRadius determines the appropriate H3 resolution for a given radius
func (i *Indexer) getResolutionForRadius(radiusMeters float64) int {
	// This is a simplified version. In production, you might want to use a more sophisticated
	// algorithm to determine the appropriate resolution based on the radius
	if radiusMeters <= 100 {
		return 9
	} else if radiusMeters <= 1000 {
		return 8
	} else if radiusMeters <= 10000 {
		return 7
	} else {
		return 6
	}
}

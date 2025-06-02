package h3

import (
	"github.com/uber/h3-go/v3"
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
	geo := h3.LatLng{
		Lat: lat,
		Lng: lng,
	}
	index := h3.LatLngToCell(geo, i.resolution)
	return index.String()
}

// GetNeighbors returns the H3 indexes of neighboring cells
func (i *Indexer) GetNeighbors(h3Index string) ([]string, error) {
	index, err := h3.ParseCell(h3Index)
	if err != nil {
		return nil, err
	}

	neighbors := h3.GridDisk(index, 1)
	result := make([]string, len(neighbors))
	for j, n := range neighbors {
		result[j] = n.String()
	}
	return result, nil
}

// GetCellsInRadius returns all H3 cells within a specified radius (in meters)
func (i *Indexer) GetCellsInRadius(lat, lng float64, radiusMeters float64) ([]string, error) {
	geo := h3.LatLng{
		Lat: lat,
		Lng: lng,
	}

	// Convert radius to appropriate resolution
	resolution := i.getResolutionForRadius(radiusMeters)

	// Get the center cell
	center := h3.LatLngToCell(geo, resolution)

	// Get all cells within the radius
	cells := h3.GridDiskDistances(center, int(radiusMeters/1000)) // Convert meters to kilometers

	result := make([]string, 0)
	for _, cell := range cells {
		result = append(result, cell.String())
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

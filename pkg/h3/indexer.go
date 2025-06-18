package h3

import (
	"fmt"
	"math"

	"github.com/uber/h3-go/v4"
)

// Indexer handles H3 indexing operations
type Indexer struct {
	resolution int
}

// NewIndexer creates a new H3 indexer with the specified resolution
func NewIndexer(resolution int) *Indexer {
	return &Indexer{
		resolution: resolution,
	}
}

// GetCellsInRadius returns all H3 cells within the specified radius of a point
func (i *Indexer) GetCellsInRadius(lat, lng, radiusKm float64) ([]string, error) {
	// Convert lat/lng to H3 cell
	cell, err := h3.LatLngToCell(h3.NewLatLng(lat, lng), i.resolution)
	if err != nil {
		return nil, fmt.Errorf("failed to convert lat/lng to cell: %w", err)
	}

	// Calculate the number of rings needed based on radius
	// H3 resolution 9 has cells of approximately 1km
	rings := int(math.Ceil(radiusKm))

	// Get cells in rings
	cells, err := h3.GridDisk(cell, rings)
	if err != nil {
		return nil, fmt.Errorf("failed to get grid disk: %w", err)
	}

	// Convert cells to strings
	result := make([]string, len(cells))
	for j, c := range cells {
		result[j] = h3.IndexToString(uint64(c))
	}

	return result, nil
}

// GetDistance calculates the distance between two points in kilometers
func (i *Indexer) GetDistance(lat1, lng1, lat2, lng2 float64) float64 {
	return h3.GreatCircleDistanceKm(
		h3.NewLatLng(lat1, lng1),
		h3.NewLatLng(lat2, lng2),
	)
}

// GetCellFromLatLng converts a lat/lng pair to an H3 cell
func (i *Indexer) GetCellFromLatLng(lat, lng float64) (string, error) {
	cell, err := h3.LatLngToCell(h3.NewLatLng(lat, lng), i.resolution)
	if err != nil {
		return "", fmt.Errorf("failed to convert lat/lng to cell: %w", err)
	}
	return h3.IndexToString(uint64(cell)), nil
}

// GetLatLngFromCell converts an H3 cell to a lat/lng pair
func (i *Indexer) GetLatLngFromCell(cellStr string) (float64, float64, error) {
	index := h3.IndexFromString(cellStr)
	cell := h3.Cell(index)
	latLng, err := h3.CellToLatLng(cell)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to convert cell to lat/lng: %w", err)
	}
	return latLng.Lat, latLng.Lng, nil
}

// GetCellsInRadiusWithDistance returns all H3 cells within the specified radius of a point,
// along with their distances from the center point
func (i *Indexer) GetCellsInRadiusWithDistance(lat, lng, radiusKm float64) (map[string]float64, error) {
	// Convert lat/lng to H3 cell
	cell, err := h3.LatLngToCell(h3.NewLatLng(lat, lng), i.resolution)
	if err != nil {
		return nil, fmt.Errorf("failed to convert lat/lng to cell: %w", err)
	}

	// Calculate the number of rings needed based on radius
	rings := int(math.Ceil(radiusKm))

	// Get cells in rings
	cells, err := h3.GridDisk(cell, rings)
	if err != nil {
		return nil, fmt.Errorf("failed to get grid disk: %w", err)
	}

	// Convert cells to strings and calculate distances
	result := make(map[string]float64)
	for _, c := range cells {
		latLng, err := h3.CellToLatLng(c)
		if err != nil {
			return nil, fmt.Errorf("failed to convert cell to lat/lng: %w", err)
		}
		distance := h3.GreatCircleDistanceKm(h3.NewLatLng(lat, lng), latLng)
		if distance <= radiusKm {
			result[h3.IndexToString(uint64(c))] = distance
		}
	}

	return result, nil
}

// GetH3Index returns the H3 index for a given latitude and longitude
func (i *Indexer) GetH3Index(lat, lng float64) string {
	geo := h3.LatLng{
		Lat: lat,
		Lng: lng,
	}
	index, _ := h3.LatLngToCell(geo, i.resolution)
	return h3.IndexToString(uint64(index))
}

// GetH3CellsInRadius returns all H3 cells within a given radius (in kilometers) of a point
func (i *Indexer) GetH3CellsInRadius(lat, lng float64, radiusKm float64) ([]string, error) {
	geo := h3.LatLng{
		Lat: lat,
		Lng: lng,
	}
	index, err := h3.LatLngToCell(geo, i.resolution)
	if err != nil {
		return nil, err
	}

	// Convert radius from kilometers to meters
	radiusM := radiusKm * 1000

	// Get edge length in meters for the current resolution
	edgeLengthM, err := h3.EdgeLengthM(h3.DirectedEdge(i.resolution))
	if err != nil {
		return nil, err
	}

	// Get all cells within the radius
	cells, err := h3.GridDisk(index, int(math.Ceil(radiusM/edgeLengthM)))
	if err != nil {
		return nil, err
	}

	// Convert cells to strings
	result := make([]string, len(cells))
	for j, cell := range cells {
		result[j] = h3.IndexToString(uint64(cell))
	}

	return result, nil
}

// GetNeighbors returns the H3 indexes of neighboring cells
func (i *Indexer) GetNeighbors(h3Index string) ([]string, error) {
	index := h3.IndexFromString(h3Index)
	cell := h3.Cell(index)
	neighbors, err := h3.GridRing(cell, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get grid ring: %w", err)
	}
	result := make([]string, len(neighbors))
	for j, n := range neighbors {
		result[j] = h3.IndexToString(uint64(n))
	}
	return result, nil
}

// GetKRing returns all H3 indexes within k rings of the center index
func (i *Indexer) GetKRing(centerIndex string, k int) ([]string, error) {
	// Get the k-ring of indexes
	indexes := make([]string, 0)
	index := h3.IndexFromString(centerIndex)
	cell := h3.Cell(index)
	ring, err := h3.GridRing(cell, k)
	if err != nil {
		return nil, fmt.Errorf("failed to get grid ring: %w", err)
	}

	// Convert back to strings
	for _, c := range ring {
		indexes = append(indexes, h3.IndexToString(uint64(c)))
	}

	return indexes, nil
}

// GetH3IndexesInRadius returns all H3 indexes within a given radius (in meters) from a center index
func (i *Indexer) GetH3IndexesInRadius(centerIndex string, radiusMeters float64) ([]string, error) {
	// Convert radius from meters to kilometers
	radiusKm := radiusMeters / 1000.0

	// Calculate the number of rings needed based on radius
	// H3 resolution 9 has cells of approximately 1km
	rings := int(math.Ceil(radiusKm))

	// Get the k-ring of indexes
	index := h3.IndexFromString(centerIndex)
	cell := h3.Cell(index)
	ring, err := h3.GridRing(cell, rings)
	if err != nil {
		return nil, fmt.Errorf("failed to get grid ring: %w", err)
	}

	// Convert back to strings
	result := make([]string, len(ring))
	for j, c := range ring {
		result[j] = h3.IndexToString(uint64(c))
	}

	return result, nil
}

// GetCell returns the H3 cell for the given coordinates (required by service.H3Indexer)
func (i *Indexer) GetCell(lat, lng float64) (string, error) {
	cell, err := h3.LatLngToCell(h3.NewLatLng(lat, lng), i.resolution)
	if err != nil {
		return "", err
	}
	return h3.IndexToString(uint64(cell)), nil
}

package service

import (
	"github.com/uber/h3-go/v4"
)

// H3Indexer defines the interface for H3 indexing operations
type H3Indexer interface {
	// GetCellsInRadius returns all H3 cells within the specified radius (in kilometers)
	GetCellsInRadius(lat, lng, radius float64) ([]string, error)

	// GetCell returns the H3 cell for the given coordinates
	GetCell(lat, lng float64) (string, error)

	// GetDistance returns the distance between two points in kilometers
	GetDistance(lat1, lng1, lat2, lng2 float64) float64
}

type H3IndexerStruct struct {
	resolution int
}

func NewH3Indexer(resolution int) *H3IndexerStruct {
	return &H3IndexerStruct{
		resolution: resolution,
	}
}

func (i *H3IndexerStruct) GetCellsInRadius(lat, lng, radius float64) ([]string, error) {
	cell, err := h3.LatLngToCell(h3.NewLatLng(lat, lng), i.resolution)
	if err != nil {
		return nil, err
	}
	rings := int(radius)
	cells, err := h3.GridDisk(cell, rings)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(cells))
	for j, c := range cells {
		result[j] = h3.IndexToString(uint64(c))
	}
	return result, nil
}

func (i *H3IndexerStruct) GetCell(lat, lng float64) (string, error) {
	cell, err := h3.LatLngToCell(h3.NewLatLng(lat, lng), i.resolution)
	if err != nil {
		return "", err
	}
	return h3.IndexToString(uint64(cell)), nil
}

func (i *H3IndexerStruct) GetDistance(lat1, lng1, lat2, lng2 float64) float64 {
	return h3.GreatCircleDistanceKm(
		h3.NewLatLng(lat1, lng1),
		h3.NewLatLng(lat2, lng2),
	)
}

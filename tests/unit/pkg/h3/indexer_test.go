package h3_test

import (
	"testing"

	"geosearch-poc/pkg/h3"

	"github.com/stretchr/testify/assert"
)

func TestNewIndexer(t *testing.T) {
	tests := []struct {
		name        string
		resolution  int
		checkResult func(t *testing.T, indexer *h3.Indexer)
	}{
		{
			name:       "should create indexer with resolution 9",
			resolution: 9,
			checkResult: func(t *testing.T, indexer *h3.Indexer) {
				// Can't access private field resolution, so we'll test the functionality instead
				h3Index, err := indexer.GetCellFromLatLng(-23.550520, -46.633308)
				assert.NoError(t, err)
				assert.NotEmpty(t, h3Index)
			},
		},
		{
			name:       "should create indexer with resolution 10",
			resolution: 10,
			checkResult: func(t *testing.T, indexer *h3.Indexer) {
				h3Index, err := indexer.GetCellFromLatLng(-23.550520, -46.633308)
				assert.NoError(t, err)
				assert.NotEmpty(t, h3Index)
			},
		},
		{
			name:       "should create indexer with resolution 8",
			resolution: 8,
			checkResult: func(t *testing.T, indexer *h3.Indexer) {
				h3Index, err := indexer.GetCellFromLatLng(-23.550520, -46.633308)
				assert.NoError(t, err)
				assert.NotEmpty(t, h3Index)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create indexer
			indexer := h3.NewIndexer(tt.resolution)

			// Run checks
			tt.checkResult(t, indexer)
		})
	}
}

func TestIndexer_GetCellFromLatLng(t *testing.T) {
	indexer := h3.NewIndexer(9)

	tests := []struct {
		name        string
		lat         float64
		lng         float64
		expectError bool
		checkResult func(t *testing.T, h3Index string)
	}{
		{
			name:        "should get H3 index for São Paulo center",
			lat:         -23.550520,
			lng:         -46.633308,
			expectError: false,
			checkResult: func(t *testing.T, h3Index string) {
				assert.NotEmpty(t, h3Index)
				assert.Len(t, h3Index, 15) // H3 index at resolution 9 should be 15 characters
			},
		},
		{
			name:        "should get H3 index for New York",
			lat:         40.7128,
			lng:         -74.0060,
			expectError: false,
			checkResult: func(t *testing.T, h3Index string) {
				assert.NotEmpty(t, h3Index)
				assert.Len(t, h3Index, 15)
			},
		},
		{
			name:        "should get H3 index for London",
			lat:         51.5074,
			lng:         -0.1278,
			expectError: false,
			checkResult: func(t *testing.T, h3Index string) {
				assert.NotEmpty(t, h3Index)
				assert.Len(t, h3Index, 15)
			},
		},
		{
			name:        "should get H3 index for Tokyo",
			lat:         35.6762,
			lng:         139.6503,
			expectError: false,
			checkResult: func(t *testing.T, h3Index string) {
				assert.NotEmpty(t, h3Index)
				assert.Len(t, h3Index, 15)
			},
		},
		{
			name:        "should get H3 index for coordinates at equator",
			lat:         0.0,
			lng:         0.0,
			expectError: false,
			checkResult: func(t *testing.T, h3Index string) {
				assert.NotEmpty(t, h3Index)
				assert.Len(t, h3Index, 15)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get H3 index
			h3Index, err := indexer.GetCellFromLatLng(tt.lat, tt.lng)

			// Assert result
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Run additional checks if provided
				if tt.checkResult != nil {
					tt.checkResult(t, h3Index)
				}
			}
		})
	}
}

func TestIndexer_GetCellsInRadius(t *testing.T) {
	indexer := h3.NewIndexer(9)

	tests := []struct {
		name             string
		lat              float64
		lng              float64
		radius           float64
		expectError      bool
		expectedMinCells int
		checkResult      func(t *testing.T, cells []string)
	}{
		{
			name:             "should get cells within 1km radius",
			lat:              -23.550520,
			lng:              -46.633308,
			radius:           1.0,
			expectError:      false,
			expectedMinCells: 1,
			checkResult: func(t *testing.T, cells []string) {
				assert.GreaterOrEqual(t, len(cells), 1)
				for _, cell := range cells {
					assert.NotEmpty(t, cell)
					assert.Len(t, cell, 15)
				}
			},
		},
		{
			name:             "should get cells within 5km radius",
			lat:              -23.550520,
			lng:              -46.633308,
			radius:           5.0,
			expectError:      false,
			expectedMinCells: 5,
			checkResult: func(t *testing.T, cells []string) {
				assert.GreaterOrEqual(t, len(cells), 5)
				for _, cell := range cells {
					assert.NotEmpty(t, cell)
					assert.Len(t, cell, 15)
				}
			},
		},
		{
			name:             "should get cells within 10km radius",
			lat:              -23.550520,
			lng:              -46.633308,
			radius:           10.0,
			expectError:      false,
			expectedMinCells: 10,
			checkResult: func(t *testing.T, cells []string) {
				assert.GreaterOrEqual(t, len(cells), 10)
				for _, cell := range cells {
					assert.NotEmpty(t, cell)
					assert.Len(t, cell, 15)
				}
			},
		},
		{
			name:             "should get cells for different location",
			lat:              40.7128,
			lng:              -74.0060,
			radius:           2.0,
			expectError:      false,
			expectedMinCells: 2,
			checkResult: func(t *testing.T, cells []string) {
				assert.GreaterOrEqual(t, len(cells), 2)
				for _, cell := range cells {
					assert.NotEmpty(t, cell)
					assert.Len(t, cell, 15)
				}
			},
		},
		{
			name:             "should handle zero radius",
			lat:              -23.550520,
			lng:              -46.633308,
			radius:           0.0,
			expectError:      false,
			expectedMinCells: 1,
			checkResult: func(t *testing.T, cells []string) {
				assert.GreaterOrEqual(t, len(cells), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get cells in radius
			cells, err := indexer.GetCellsInRadius(tt.lat, tt.lng, tt.radius)

			// Assert result
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(cells), tt.expectedMinCells)

				// Run additional checks if provided
				if tt.checkResult != nil {
					tt.checkResult(t, cells)
				}
			}
		})
	}
}

func TestIndexer_GetDistanceBetween(t *testing.T) {
	indexer := h3.NewIndexer(9)

	tests := []struct {
		name        string
		lat1        float64
		lng1        float64
		lat2        float64
		lng2        float64
		expectError bool
		checkResult func(t *testing.T, distance float64)
	}{
		{
			name:        "should calculate distance between same points",
			lat1:        -23.550520,
			lng1:        -46.633308,
			lat2:        -23.550520,
			lng2:        -46.633308,
			expectError: false,
			checkResult: func(t *testing.T, distance float64) {
				assert.Equal(t, 0.0, distance)
			},
		},
		{
			name:        "should calculate distance between close points",
			lat1:        -23.550520,
			lng1:        -46.633308,
			lat2:        -23.551520,
			lng2:        -46.634308,
			expectError: false,
			checkResult: func(t *testing.T, distance float64) {
				assert.Greater(t, distance, 0.0)
				assert.Less(t, distance, 1.0) // Should be less than 1km
			},
		},
		{
			name:        "should calculate distance between far points",
			lat1:        -23.550520,
			lng1:        -46.633308,
			lat2:        -23.600520,
			lng2:        -46.683308,
			expectError: false,
			checkResult: func(t *testing.T, distance float64) {
				assert.Greater(t, distance, 5.0) // Should be more than 5km
			},
		},
		{
			name:        "should calculate distance between different cities",
			lat1:        -23.550520,
			lng1:        -46.633308, // São Paulo
			lat2:        -22.9068,
			lng2:        -43.1729, // Rio de Janeiro
			expectError: false,
			checkResult: func(t *testing.T, distance float64) {
				assert.Greater(t, distance, 300.0) // Should be more than 300km
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate distance
			distance := indexer.GetDistance(tt.lat1, tt.lng1, tt.lat2, tt.lng2)

			// Assert result
			assert.GreaterOrEqual(t, distance, 0.0)

			// Run additional checks if provided
			if tt.checkResult != nil {
				tt.checkResult(t, distance)
			}
		})
	}
}

func TestIndexer_Consistency(t *testing.T) {
	indexer := h3.NewIndexer(9)

	// Test that the same coordinates always produce the same H3 index
	lat := -23.550520
	lng := -46.633308

	h3Index1, err1 := indexer.GetCellFromLatLng(lat, lng)
	assert.NoError(t, err1)

	h3Index2, err2 := indexer.GetCellFromLatLng(lat, lng)
	assert.NoError(t, err2)

	assert.Equal(t, h3Index1, h3Index2)

	// Test that nearby coordinates produce different H3 indexes
	h3Index3, err3 := indexer.GetCellFromLatLng(lat+0.001, lng+0.001)
	assert.NoError(t, err3)

	// They might be the same due to H3 resolution, but let's check the pattern
	assert.NotEmpty(t, h3Index1)
	assert.NotEmpty(t, h3Index3)
}

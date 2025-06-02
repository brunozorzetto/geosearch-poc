package distance

import (
	"math"
)

// CalculateDistance calculates the distance between two points using the Haversine formula
// Returns the distance in meters
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth's radius in meters

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad
	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := R * c

	return distance
}

// IsWithinRadius checks if a point is within a specified radius of another point
func IsWithinRadius(lat1, lon1, lat2, lon2, radiusMeters float64) bool {
	distance := CalculateDistance(lat1, lon1, lat2, lon2)
	return distance <= radiusMeters
}

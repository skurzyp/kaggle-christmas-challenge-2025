package tree

import (
	"math"

	"github.com/engelsjk/polygol"
)

// Intersect checks if this tree intersects with another tree
func (t *ChristmasTree) Intersect(other *ChristmasTree) bool {
	poly1 := t.GetOrbPolygon()
	poly2 := other.GetOrbPolygon()

	// Convert orb.Polygon to polygol.Geom format
	geom1 := orbPolygonToGeom(poly1)
	geom2 := orbPolygonToGeom(poly2)

	// Use polygol to compute intersection
	intersection, err := polygol.Intersection(geom1, geom2)
	if err != nil {
		// If there's an error, assume no intersection
		return false
	}

	// Check if intersection is not empty
	// An empty intersection means no overlap
	return len(intersection) > 0 && len(intersection[0]) > 0
}

// IntersectionArea returns the area of overlap between two trees (0 if none)
func (t *ChristmasTree) IntersectionArea(other *ChristmasTree) float64 {
	poly1 := t.GetOrbPolygon()
	poly2 := other.GetOrbPolygon()

	geom1 := orbPolygonToGeom(poly1)
	geom2 := orbPolygonToGeom(poly2)

	intersection, err := polygol.Intersection(geom1, geom2)
	if err != nil {
		return 0
	}

	// Calculate area of intersection polygon(s)
	totalArea := 0.0
	for _, poly := range intersection {
		for _, ring := range poly {
			totalArea += calculateRingArea(ring)
		}
	}
	return totalArea
}

// calculateRingArea calculates the area of a polygon ring using the shoelace formula
func calculateRingArea(ring [][]float64) float64 {
	n := len(ring)
	if n < 3 {
		return 0
	}

	area := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += ring[i][0] * ring[j][1]
		area -= ring[j][0] * ring[i][1]
	}
	return math.Abs(area) / 2.0
}

// Package tree implements Christmas tree geometry and packing algorithms
// for the Kaggle Tree Packing Challenge.
package tree

import (
	"math"

	"github.com/engelsjk/polygol"
	"github.com/paulmach/orb"
)

// ChristmasTree represents a single tree with position and rotation
type ChristmasTree struct {
	ID    int
	X, Y  float64
	Angle float64 // Angle in DEGREES (Kaggle submission format)
}

// deg2rad converts degrees to radians
func deg2rad(deg float64) float64 {
	return deg * math.Pi / 180.0
}

// GetBoundingBox returns the axis-aligned bounding box of the rotated tree
func (t *ChristmasTree) GetBoundingBox() (float64, float64, float64, float64) {
	// Calculate bounding box from the actual orb polygon (same as used for intersection)
	poly := t.GetOrbPolygon()
	ring := poly[0]

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, pt := range ring {
		if pt[0] < minX {
			minX = pt[0]
		}
		if pt[0] > maxX {
			maxX = pt[0]
		}
		if pt[1] < minY {
			minY = pt[1]
		}
		if pt[1] > maxY {
			maxY = pt[1]
		}
	}

	return minX, minY, maxX, maxY
}

// GetOrbPolygon returns an orb.Polygon representing the tree outline
func (t *ChristmasTree) GetOrbPolygon() orb.Polygon {
	// Tree dimensions
	const (
		TrunkW       = 0.15
		TrunkH       = 0.2
		BaseW        = 0.7
		MidW         = 0.4
		TopW         = 0.25
		TipY         = 0.8
		Tier1Y       = 0.5
		Tier2Y       = 0.25
		BaseY        = 0.0
		TrunkBottomY = -TrunkH
	)

	// Create the outer ring of the polygon (COUNTER-CLOCKWISE for polygol)
	// CCW order: tip -> left side down -> trunk -> right side up -> tip
	ring := orb.Ring{
		// Start at Tip
		orb.Point{0.0, TipY},
		// Left side - Top Tier (going down left = CCW)
		orb.Point{-TopW / 2, Tier1Y},
		orb.Point{-TopW / 4, Tier1Y},
		// Left side - Middle Tier
		orb.Point{-MidW / 2, Tier2Y},
		orb.Point{-MidW / 4, Tier2Y},
		// Left side - Bottom Tier
		orb.Point{-BaseW / 2, BaseY},
		// Left Trunk
		orb.Point{-TrunkW / 2, BaseY},
		orb.Point{-TrunkW / 2, TrunkBottomY},
		// Right Trunk
		orb.Point{TrunkW / 2, TrunkBottomY},
		orb.Point{TrunkW / 2, BaseY},
		// Right side - Bottom Tier
		orb.Point{BaseW / 2, BaseY},
		// Right side - Middle Tier
		orb.Point{MidW / 4, Tier2Y},
		orb.Point{MidW / 2, Tier2Y},
		// Right side - Top Tier
		orb.Point{TopW / 4, Tier1Y},
		orb.Point{TopW / 2, Tier1Y},
		// Close the ring back to the tip
		orb.Point{0.0, TipY},
	}

	// Apply translation to tree position
	for i := range ring {
		ring[i][0] += t.X
		ring[i][1] += t.Y
	}

	// Apply rotation if needed
	if t.Angle != 0 {
		angleRad := deg2rad(t.Angle)
		cosAngle := math.Cos(angleRad)
		sinAngle := math.Sin(angleRad)

		for i := range ring {
			// Rotate around tree center (t.X, t.Y)
			x := ring[i][0] - t.X
			y := ring[i][1] - t.Y
			ring[i][0] = t.X + x*cosAngle - y*sinAngle
			ring[i][1] = t.Y + x*sinAngle + y*cosAngle
		}
	}

	return orb.Polygon{ring}
}

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

// orbPolygonToGeom converts an orb.Polygon to polygol.Geom format
func orbPolygonToGeom(poly orb.Polygon) polygol.Geom {
	geom := make(polygol.Geom, 1)            // One polygon
	geom[0] = make([][][]float64, len(poly)) // Rings (outer + holes)

	for i, ring := range poly {
		geom[0][i] = make([][]float64, len(ring))
		for j, point := range ring {
			geom[0][i][j] = []float64{point[0], point[1]}
		}
	}

	return geom
}

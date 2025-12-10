package main

import (
	"math"

	"github.com/engelsjk/polygol"
	"github.com/paulmach/orb"
	"github.com/solarlune/resolv"
)

// ChristmasTree represents a single tree with position and rotation
type ChristmasTree struct {
	ID    int
	X, Y  float64
	Angle float64
}

// GetPolygon returns a single polygon representing the entire tree outline
func (t *ChristmasTree) GetPolygon() *resolv.ConvexPolygon {

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

	// Full tree outline points (matching Python implementation)
	points := []float64{
		// Start at Tip
		0.0, TipY,
		// Right side - Top Tier
		TopW / 2, Tier1Y,
		TopW / 4, Tier1Y,
		// Right side - Middle Tier
		MidW / 2, Tier2Y,
		MidW / 4, Tier2Y,
		// Right side - Bottom Tier
		BaseW / 2, BaseY,
		// Right Trunk
		TrunkW / 2, BaseY,
		TrunkW / 2, TrunkBottomY,
		// Left Trunk
		-TrunkW / 2, TrunkBottomY,
		-TrunkW / 2, BaseY,
		// Left side - Bottom Tier
		-BaseW / 2, BaseY,
		// Left side - Middle Tier
		-MidW / 4, Tier2Y,
		-MidW / 2, Tier2Y,
		// Left side - Top Tier
		-TopW / 4, Tier1Y,
		-TopW / 2, Tier1Y,
	}

	poly := resolv.NewConvexPolygon(t.X, t.Y, points)

	poly.SetRotation(t.Angle)

	return poly
}

func (t *ChristmasTree) GetBoundingBox() (float64, float64, float64, float64) {
	bounds := t.GetPolygon().Bounds()
	return bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y
}

// GetOrbPolygon returns an orb.Polygon representing the tree outline
func (t *ChristmasTree) GetOrbPolygon() orb.Polygon {
	// Tree dimensions (same as GetPolygon)
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

	// Create the outer ring of the polygon
	ring := orb.Ring{
		// Start at Tip
		orb.Point{0.0, TipY},
		// Right side - Top Tier
		orb.Point{TopW / 2, Tier1Y},
		orb.Point{TopW / 4, Tier1Y},
		// Right side - Middle Tier
		orb.Point{MidW / 2, Tier2Y},
		orb.Point{MidW / 4, Tier2Y},
		// Right side - Bottom Tier
		orb.Point{BaseW / 2, BaseY},
		// Right Trunk
		orb.Point{TrunkW / 2, BaseY},
		orb.Point{TrunkW / 2, TrunkBottomY},
		// Left Trunk
		orb.Point{-TrunkW / 2, TrunkBottomY},
		orb.Point{-TrunkW / 2, BaseY},
		// Left side - Bottom Tier
		orb.Point{-BaseW / 2, BaseY},
		// Left side - Middle Tier
		orb.Point{-MidW / 4, Tier2Y},
		orb.Point{-MidW / 2, Tier2Y},
		// Left side - Top Tier
		orb.Point{-TopW / 4, Tier1Y},
		orb.Point{-TopW / 2, Tier1Y},
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
		cosAngle := math.Cos(t.Angle)
		sinAngle := math.Sin(t.Angle)

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
	// polygol.Geom is [][][][]float64
	// First level: array of polygons (for MultiPolygon, we have just one)
	// Second level: array of rings (outer ring + holes)
	// Third level: array of points
	// Fourth level: [x, y] coordinates

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

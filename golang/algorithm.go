package main

import (
	"math"
	"math/rand"

	"github.com/tidwall/rtree"
)

// generateWeightedAngle generates a random angle with distribution weighted by abs(sin(2*angle))
func generateWeightedAngle() float64 {
	for {
		angle := rand.Float64() * 2 * math.Pi
		if rand.Float64() < math.Abs(math.Sin(2*angle)) {
			return angle
		}
	}
}

// initializeTrees builds a greedy packing of n trees
func initializeTrees(numTrees int, existingTrees []ChristmasTree) ([]ChristmasTree, float64) {
	if numTrees == 0 {
		return []ChristmasTree{}, 0
	}

	placedTrees := make([]ChristmasTree, len(existingTrees))
	copy(placedTrees, existingTrees)

	// Spatial Index for fast collision detection
	tr := rtree.RTree{}
	for i, t := range placedTrees {
		minX, minY, maxX, maxY := t.GetBoundingBox()
		tr.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, i)
	}

	numToAdd := numTrees - len(placedTrees)
	if numToAdd > 0 {
		// If starting from scratch, place first tree at origin
		if len(placedTrees) == 0 {
			t := ChristmasTree{ID: 0, X: 0, Y: 0, Angle: rand.Float64() * 2 * math.Pi}
			placedTrees = append(placedTrees, t)
			minX, minY, maxX, maxY := t.GetBoundingBox()
			tr.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, 0)
			numToAdd--
		}

		for i := 0; i < numToAdd; i++ {
			newID := len(placedTrees)
			treeToPlace := ChristmasTree{ID: newID, Angle: rand.Float64() * 2 * math.Pi}

			var bestX, bestY float64
			minRadius := math.Inf(1)
			foundPlacement := false

			// Try 10 random starting attempts
			for attempt := 0; attempt < 10; attempt++ {
				angle := generateWeightedAngle()
				vx := math.Cos(angle)
				vy := math.Sin(angle)

				radius := 20.0
				stepIn := 0.5

				collisionFound := false

				// Move towards center
				for radius >= 0 {
					px := radius * vx
					py := radius * vy

					treeToPlace.X = px
					treeToPlace.Y = py

					// Check collision using spatial index
					candidateBoundsMinX, candidateBoundsMinY, candidateBoundsMaxX, candidateBoundsMaxY := treeToPlace.GetBoundingBox()

					// Query RTree for potential collisions
					possibleCollisions := []int{}
					tr.Search(
						[2]float64{candidateBoundsMinX, candidateBoundsMinY},
						[2]float64{candidateBoundsMaxX, candidateBoundsMaxY},
						func(min, max [2]float64, data interface{}) bool {
							possibleCollisions = append(possibleCollisions, data.(int))
							return true
						},
					)

					// Check for actual collisions
					isColliding := false
					for _, otherIdx := range possibleCollisions {
						if CheckCollision(&treeToPlace, &placedTrees[otherIdx]) {
							isColliding = true
							break
						}
					}

					if isColliding {
						collisionFound = true
						break
					}

					radius -= stepIn
				}

				// Back up if collision was found
				if collisionFound {
					stepOut := 0.05
					for {
						radius += stepOut
						px := radius * vx
						py := radius * vy

						treeToPlace.X = px
						treeToPlace.Y = py

						// Check collision again
						candidateBoundsMinX, candidateBoundsMinY, candidateBoundsMaxX, candidateBoundsMaxY := treeToPlace.GetBoundingBox()

						possibleCollisions := []int{}
						tr.Search(
							[2]float64{candidateBoundsMinX, candidateBoundsMinY},
							[2]float64{candidateBoundsMaxX, candidateBoundsMaxY},
							func(min, max [2]float64, data interface{}) bool {
								possibleCollisions = append(possibleCollisions, data.(int))
								return true
							},
						)

						isColliding := false
						for _, otherIdx := range possibleCollisions {
							if CheckCollision(&treeToPlace, &placedTrees[otherIdx]) {
								isColliding = true
								break
							}
						}

						if !isColliding {
							break
						}
					}
				} else {
					// No collision found even at center
					radius = 0
					treeToPlace.X = 0
					treeToPlace.Y = 0
				}

				if radius < minRadius {
					minRadius = radius
					bestX = treeToPlace.X
					bestY = treeToPlace.Y
					foundPlacement = true
				}
			}

			if foundPlacement {
				treeToPlace.X = bestX
				treeToPlace.Y = bestY
				placedTrees = append(placedTrees, treeToPlace)
				minX, minY, maxX, maxY := treeToPlace.GetBoundingBox()
				tr.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, newID)
			}
		}
	}

	// Calculate side length
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, t := range placedTrees {
		tMinX, tMinY, tMaxX, tMaxY := t.GetBoundingBox()
		if tMinX < minX {
			minX = tMinX
		}
		if tMinY < minY {
			minY = tMinY
		}
		if tMaxX > maxX {
			maxX = tMaxX
		}
		if tMaxY > maxY {
			maxY = tMaxY
		}
	}

	width := maxX - minX
	height := maxY - minY
	sideLength := math.Max(width, height)

	return placedTrees, sideLength
}

// CheckCollision checks if two trees collide
func CheckCollision(tree1, tree2 *ChristmasTree) bool {
	return tree1.Intersect(tree2)
}

package main

import (
	"math"
	"math/rand"
)

// generateWeightedAngle generates a random angle in DEGREES with distribution weighted by abs(sin(2*angle))
func generateWeightedAngle() float64 {
	for {
		angleDeg := rand.Float64() * 360.0
		angleRad := angleDeg * math.Pi / 180.0
		if rand.Float64() < math.Abs(math.Sin(2*angleRad)) {
			return angleDeg
		}
	}
}

// checkCollisionWithAll checks if tree collides with any tree in the list
func checkCollisionWithAll(tree *ChristmasTree, placedTrees []ChristmasTree) bool {
	for i := range placedTrees {
		if CheckCollision(tree, &placedTrees[i]) {
			return true
		}
	}
	return false
}

// initializeTrees builds a greedy packing of n trees
func initializeTrees(numTrees int, existingTrees []ChristmasTree) ([]ChristmasTree, float64) {
	if numTrees == 0 {
		return []ChristmasTree{}, 0
	}

	placedTrees := make([]ChristmasTree, len(existingTrees))
	copy(placedTrees, existingTrees)

	numToAdd := numTrees - len(placedTrees)
	if numToAdd > 0 {
		// If starting from scratch, place first tree at origin
		if len(placedTrees) == 0 {
			t := ChristmasTree{ID: 0, X: 0, Y: 0, Angle: rand.Float64() * 360.0}
			placedTrees = append(placedTrees, t)
			numToAdd--
		}

		for i := 0; i < numToAdd; i++ {
			newID := len(placedTrees)
			treeToPlace := ChristmasTree{ID: newID, Angle: rand.Float64() * 360.0}

			var bestX, bestY float64
			minRadius := math.Inf(1)
			foundPlacement := false

			// Try 10 random starting attempts
			for attempt := 0; attempt < 10; attempt++ {
				angle := generateWeightedAngle()
				angleRad := angle * math.Pi / 180.0
				vx := math.Cos(angleRad)
				vy := math.Sin(angleRad)

				radius := 20.0
				stepIn := 0.5

				collisionFound := false

				// Move towards center
				for radius >= 0 {
					px := radius * vx
					py := radius * vy

					treeToPlace.X = px
					treeToPlace.Y = py

					// Check collision with ALL placed trees (no spatial index)
					isColliding := checkCollisionWithAll(&treeToPlace, placedTrees)

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

						// Check collision with ALL placed trees
						isColliding := checkCollisionWithAll(&treeToPlace, placedTrees)

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

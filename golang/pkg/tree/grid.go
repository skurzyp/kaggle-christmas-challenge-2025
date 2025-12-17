package tree

import (
	"math"

	"github.com/tidwall/rtree"
)

// GridConfig holds configuration for grid-based tree placement
type GridConfig struct {
	HorizontalSpacing float64 // Spacing between trees in a row (default: 0.7)
	EvenRowY          float64 // Y spacing for even rows (default: 1.0)
	OddRowOffsetY     float64 // Y offset for odd rows (default: 0.8)
	OddRowOffsetX     float64 // X offset for odd rows (default: 0.35, which is 0.7/2)
}

// DefaultGridConfig returns the default grid configuration
func DefaultGridConfig() *GridConfig {
	return &GridConfig{
		HorizontalSpacing: 0.7,
		EvenRowY:          1.0,
		OddRowOffsetY:     0.8,
		OddRowOffsetX:     0.35, // 0.7/2 for staggered placement
	}
}

// GridSolution represents a grid-based solution attempt
type GridSolution struct {
	Trees []ChristmasTree
	Score float64
	NEven int // Number of trees per even row
	NOdd  int // Number of trees per odd row
}

// InitializeTreesGrid places trees in a grid pattern with alternating row orientations
// This is the Go equivalent of the Python find_best_trees_with_collision function
func InitializeTreesGrid(numTrees int, config *GridConfig) ([]ChristmasTree, float64) {
	if config == nil {
		config = DefaultGridConfig()
	}

	if numTrees == 0 {
		return []ChristmasTree{}, 0
	}

	var bestTrees []ChristmasTree
	bestScore := math.MaxFloat64

	// Try different combinations of even/odd row tree counts
	for nEven := 1; nEven <= numTrees; nEven++ {
		for nOdd := nEven; nOdd >= nEven-1 && nOdd >= 0; nOdd-- {
			trees := tryGridPlacement(numTrees, nEven, nOdd, config)

			// Check if we placed all trees
			if len(trees) != numTrees {
				continue
			}

			// Calculate score
			score := calculateGridScore(trees)

			if score < bestScore {
				bestScore = score
				bestTrees = trees
			}
		}
	}

	return bestTrees, bestScore
}

// tryGridPlacement attempts to place numTrees in a grid with nEven trees per even row
// and nOdd trees per odd row, using collision detection
func tryGridPlacement(numTrees, nEven, nOdd int, config *GridConfig) []ChristmasTree {
	var allTrees []ChristmasTree

	// Build RTree for collision detection
	tr := rtree.RTree{}

	remaining := numTrees
	rowIndex := 0

	for remaining > 0 {
		// Determine how many trees to place in this row
		var treesInRow int
		if rowIndex%2 == 0 {
			treesInRow = min(remaining, nEven)
		} else {
			treesInRow = min(remaining, nOdd)
		}
		remaining -= treesInRow

		// Calculate row parameters
		var angle float64
		var xOffset float64
		var y float64

		if rowIndex%2 == 0 {
			// Even row: angle 0, no x offset
			angle = 0
			xOffset = 0
			y = float64(rowIndex/2) * config.EvenRowY
		} else {
			// Odd row: angle 180, x offset, y offset
			angle = 180
			xOffset = config.OddRowOffsetX
			y = config.OddRowOffsetY + float64((rowIndex-1)/2)*config.EvenRowY
		}

		// Place trees in this row
		for i := 0; i < treesInRow; i++ {
			x := float64(i)*config.HorizontalSpacing + xOffset

			candidateTree := ChristmasTree{
				ID:    len(allTrees),
				X:     x,
				Y:     y,
				Angle: angle,
			}

			// Check for collision using RTree
			minX, minY, maxX, maxY := candidateTree.GetBoundingBox()

			// Query potential collisions
			hasCollision := false
			tr.Search(
				[2]float64{minX, minY},
				[2]float64{maxX, maxY},
				func(treeMin, treeMax [2]float64, data interface{}) bool {
					idx := data.(int)
					if candidateTree.Intersect(&allTrees[idx]) {
						hasCollision = true
						return false // Stop searching
					}
					return true
				},
			)

			if hasCollision {
				// Skip this tree if it collides
				continue
			}

			// Add tree to placement
			allTrees = append(allTrees, candidateTree)
			tr.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, len(allTrees)-1)
		}

		rowIndex++
	}

	return allTrees
}

// calculateGridScore calculates the score for a grid placement (max side squared)
func calculateGridScore(trees []ChristmasTree) float64 {
	if len(trees) == 0 {
		return 0
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for i := range trees {
		tMinX, tMinY, tMaxX, tMaxY := trees[i].GetBoundingBox()
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
	side := math.Max(width, height)

	return side * side
}

// FindBestGridSolution finds the best grid-based solution for n trees
// Returns the best score and trees
func FindBestGridSolution(numTrees int) (float64, []ChristmasTree) {
	config := DefaultGridConfig()
	trees, score := InitializeTreesGrid(numTrees, config)
	return score, trees
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

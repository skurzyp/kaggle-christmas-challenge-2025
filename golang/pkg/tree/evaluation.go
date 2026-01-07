package tree

import (
	"math"

	"github.com/tidwall/rtree"
)

// HasCollision checks if any trees in the list collide with each other
func HasCollision(trees []ChristmasTree) bool {
	if len(trees) < 2 {
		return false
	}

	// Build spatial index
	tr := rtree.RTree{}
	for i := range trees {
		minX, minY, maxX, maxY := trees[i].GetBoundingBox()
		tr.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, i)
	}

	// Check each tree against potential collisions
	for i := range trees {
		minX, minY, maxX, maxY := trees[i].GetBoundingBox()

		collision := false
		tr.Search(
			[2]float64{minX, minY},
			[2]float64{maxX, maxY},
			func(min, max [2]float64, data interface{}) bool {
				j := data.(int)
				if i != j && trees[i].Intersect(&trees[j]) {
					collision = true
					return false // Stop searching
				}
				return true
			},
		)

		if collision {
			return true
		}
	}

	return false
}

// CalculateSideLength calculates the bounding box side length for a list of trees
func CalculateSideLength(trees []ChristmasTree) float64 {
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
	return math.Max(width, height)
}

// CalculateScore calculates the score (same as side length for single group)
func CalculateScore(trees []ChristmasTree) float64 {
	return CalculateSideLength(trees)
}

// CalculateTotalOverlap computes the sum of all pairwise overlap areas
func CalculateTotalOverlap(trees []ChristmasTree) float64 {
	if len(trees) < 2 {
		return 0
	}

	// Build spatial index for broad-phase collision detection
	tr := rtree.RTree{}
	for i := range trees {
		minX, minY, maxX, maxY := trees[i].GetBoundingBox()
		tr.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, i)
	}

	totalOverlap := 0.0

	// Check each tree against potential collisions
	for i := range trees {
		minX, minY, maxX, maxY := trees[i].GetBoundingBox()

		tr.Search(
			[2]float64{minX, minY},
			[2]float64{maxX, maxY},
			func(min, max [2]float64, data interface{}) bool {
				j := data.(int)
				if j > i { // Only count each pair once
					area := trees[i].IntersectionArea(&trees[j])
					totalOverlap += area
				}
				return true
			},
		)
	}

	return totalOverlap
}

// CalculateTreeOverlap computes the total overlap area for a single tree with all others
// This is more efficient when only one tree has moved
func CalculateTreeOverlap(trees []ChristmasTree, treeIndex int) float64 {
	if len(trees) < 2 || treeIndex < 0 || treeIndex >= len(trees) {
		return 0
	}

	tree := &trees[treeIndex]
	minX, minY, maxX, maxY := tree.GetBoundingBox()

	totalOverlap := 0.0

	// Simple O(n) check against all other trees
	// Could be optimized further with R-tree if needed
	for j := range trees {
		if j != treeIndex {
			// Quick bounding box check first
			otherMinX, otherMinY, otherMaxX, otherMaxY := trees[j].GetBoundingBox()
			if minX <= otherMaxX && maxX >= otherMinX && minY <= otherMaxY && maxY >= otherMinY {
				area := tree.IntersectionArea(&trees[j])
				totalOverlap += area
			}
		}
	}

	return totalOverlap
}

// CalculatePenalizedScore returns BoundingBox + λ × TotalOverlap
func CalculatePenalizedScore(trees []ChristmasTree, overlapPenalty float64) float64 {
	bboxScore := CalculateSideLength(trees)
	overlap := CalculateTotalOverlap(trees)
	return bboxScore + overlapPenalty*overlap
}

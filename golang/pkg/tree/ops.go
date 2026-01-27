package tree

import (
	"math"
)

// HasOvl checks if the tree at index i overlaps with any other tree
func HasOvl(trees []ChristmasTree, i int) bool {
	if i < 0 || i >= len(trees) {
		return false
	}
	target := &trees[i]
	for j := range trees {
		if i == j {
			continue
		}
		if target.Intersect(&trees[j]) {
			return true
		}
	}
	return false
}

// AnyOvl checks if there is any overlap in the entire configuration
func AnyOvl(trees []ChristmasTree) bool {
	for i := range trees {
		for j := i + 1; j < len(trees); j++ {
			if trees[i].Intersect(&trees[j]) {
				return true
			}
		}
	}
	return false
}

// GetBounds calculates the bounding box of the entire configuration
func GetBounds(trees []ChristmasTree) (minX, minY, maxX, maxY float64) {
	if len(trees) == 0 {
		return 0, 0, 0, 0
	}
	minX, minY = math.MaxFloat64, math.MaxFloat64
	maxX, maxY = -math.MaxFloat64, -math.MaxFloat64

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
	return
}

// Side calculates the maximum dimension of the bounding box
func Side(trees []ChristmasTree) float64 {
	minX, minY, maxX, maxY := GetBounds(trees)
	width := maxX - minX
	height := maxY - minY
	return math.Max(width, height)
}

// Score calculates the score as side^2 / n
func Score(trees []ChristmasTree) float64 {
	if len(trees) == 0 {
		return 0
	}
	s := Side(trees)
	return (s * s) / float64(len(trees))
}

// GetBoundary returns indices of trees that are close to the bounding box boundary
func GetBoundary(trees []ChristmasTree) []int {
	var boundary []int
	if len(trees) == 0 {
		return boundary
	}

	gx0, gy0, gx1, gy1 := GetBounds(trees)
	eps := 0.01

	for i := range trees {
		// Calculate individual tree bounds to check distance to global bounds
		tx0, ty0, tx1, ty1 := trees[i].GetBoundingBox()

		if (tx0-gx0 < eps) || (gx1-tx1 < eps) || (ty0-gy0 < eps) || (gy1-ty1 < eps) {
			boundary = append(boundary, i)
		}
	}
	return boundary
}

// RemoveTree removes the tree at the specified index and returns a new slice
func RemoveTree(trees []ChristmasTree, removeIdx int) []ChristmasTree {
	if removeIdx < 0 || removeIdx >= len(trees) {
		return trees
	}
	// Create a new slice with capacity len-1
	newTrees := make([]ChristmasTree, 0, len(trees)-1)
	newTrees = append(newTrees, trees[:removeIdx]...)
	newTrees = append(newTrees, trees[removeIdx+1:]...)
	return newTrees
}

// SwapTrees attempts to swap two trees and returns true if valid AND no overlaps for involved trees
// Note: This modifies the input slice directly.
func SwapTrees(trees []ChristmasTree, i, j int) bool {
	if i == j || i < 0 || j < 0 || i >= len(trees) || j >= len(trees) {
		return false
	}

	// Swap trees
	trees[i].X, trees[j].X = trees[j].X, trees[i].X
	trees[i].Y, trees[j].Y = trees[j].Y, trees[i].Y
	trees[i].Angle, trees[j].Angle = trees[j].Angle, trees[i].Angle

	// Check for overlaps after swap
	if HasOvl(trees, i) || HasOvl(trees, j) {
		// Revert swap if invalid
		trees[i].X, trees[j].X = trees[j].X, trees[i].X
		trees[i].Y, trees[j].Y = trees[j].Y, trees[i].Y
		trees[i].Angle, trees[j].Angle = trees[j].Angle, trees[i].Angle
		return false
	}
	return true
}

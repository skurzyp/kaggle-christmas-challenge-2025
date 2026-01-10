package grid

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"tree-packing-challenge/pkg/tree"

	"github.com/tidwall/rtree"
)

// GridIndividual represents a candidate solution with simplified genome.
// The genome consists of:
//   - Angle (alpha): One tree is at alpha, the other at alpha+180 (trunks facing outward)
//   - Dx: Horizontal offset between the two trees in a block
//   - Dy: Vertical offset between the two trees in a block
//
// The number of rows and pairs per row are CALCULATED from the block dimensions
// and target number of trees - they are NOT part of the genome.
type GridIndividual struct {
	Angle float64 // Base angle (alpha). Tree A: alpha, Tree B: alpha+180
	Dx    float64 // Horizontal offset between trees in a pair
	Dy    float64 // Vertical offset between trees in a pair

	Score float64              // Cached score (SideLength)
	Trees []tree.ChristmasTree // Generated trees
}

// Config for GA
const (
	PopulationSize = 20
	Generations    = 50
	MutationRate   = 0.3
	CrossoverRate  = 0.7
	TournamentSize = 3
)

// FindBestGridGASolution runs the Genetic Algorithm to optimize block parameters
func FindBestGridGASolution(numTrees int) (float64, []tree.ChristmasTree) {
	rand.Seed(time.Now().UnixNano())
	fmt.Printf("Running Block-Based Grid GA Solver for N=%d...\n", numTrees)

	// Initialize Population
	pop := initPopulation()

	var bestInd GridIndividual
	bestInd.Score = math.MaxFloat64

	for gen := 0; gen < Generations; gen++ {
		// Evaluate fitness
		for i := range pop {
			evaluate(&pop[i], numTrees)
			if pop[i].Score < bestInd.Score {
				bestInd = pop[i]
				fmt.Printf("Gen %d: New Best Score=%.5f (Angle=%.1f°, Dx=%.3f, Dy=%.3f)\n",
					gen, bestInd.Score, bestInd.Angle, bestInd.Dx, bestInd.Dy)
			}
		}

		// Selection & Evolution
		newPop := make([]GridIndividual, 0, PopulationSize)

		// Elitism - keep the best
		newPop = append(newPop, bestInd)

		for len(newPop) < PopulationSize {
			p1 := tournamentSelection(pop)
			p2 := tournamentSelection(pop)

			child := p1 // Default clone
			if rand.Float64() < CrossoverRate {
				child = crossover(p1, p2)
			}

			if rand.Float64() < MutationRate {
				mutate(&child)
			}
			newPop = append(newPop, child)
		}
		pop = newPop
	}

	return bestInd.Score, bestInd.Trees
}

// FIXME: unused n param?
func initPopulation() []GridIndividual {
	pop := make([]GridIndividual, PopulationSize)
	for i := range pop {
		// Heuristic initialization for angles and offsets
		// Start from a known good configuration and add variance
		pop[i] = GridIndividual{
			Angle: 60.0 + (rand.Float64()-0.5)*40.0, // 40-80 degrees
			Dx:    -0.6 + (rand.Float64()-0.5)*0.4,  // -0.8 to -0.4
			Dy:    -0.1 + (rand.Float64()-0.5)*0.4,  // -0.3 to 0.1
		}
	}
	return pop
}

// checkPairCollision checks if two trees in a pair intersect
func checkPairCollision(angle, dx, dy float64) bool {
	tA := tree.ChristmasTree{X: 0, Y: 0, Angle: angle}
	tB := tree.ChristmasTree{X: dx, Y: dy, Angle: angle + 180.0}
	return tA.Intersect(&tB)
}

// findValidPairSpacing adjusts dx, dy to avoid collision within a pair
// Returns the adjusted dx, dy and whether a valid configuration was found
func findValidPairSpacing(angle, dx, dy float64) (float64, float64, bool) {
	// First check if current position is valid
	if !checkPairCollision(angle, dx, dy) {
		return dx, dy, true
	}

	// Try to find a valid position by expanding outward
	step := 0.05
	for scale := 1.0; scale <= 3.0; scale += 0.1 {
		// Try expanding in the current direction
		testDx := dx * scale
		testDy := dy * scale
		if !checkPairCollision(angle, testDx, testDy) {
			return testDx, testDy, true
		}

		// Try with small perturbations
		for _, pdx := range []float64{-step, 0, step} {
			for _, pdy := range []float64{-step, 0, step} {
				testDx := dx*scale + pdx
				testDy := dy*scale + pdy
				if !checkPairCollision(angle, testDx, testDy) {
					return testDx, testDy, true
				}
			}
		}
	}

	return dx, dy, false
}

// evaluate builds the solution from the genome and calculates the score
func evaluate(ind *GridIndividual, targetN int) {
	// First, ensure the pair configuration is valid (no intra-pair collision)
	validDx, validDy, found := findValidPairSpacing(ind.Angle, ind.Dx, ind.Dy)
	if !found {
		// Invalid pair configuration - heavily penalize
		ind.Score = 10000.0
		ind.Trees = nil
		return
	}

	// Use validated spacing
	dx, dy := validDx, validDy

	// Calculate the bounding box of a single 2-tree block
	blockWidth, blockHeight := calculateBlockDimensions(ind.Angle, dx, dy)

	// Add some spacing between blocks to prevent inter-block collisions
	blockSpacingX := blockWidth * 1.05 // 5% extra spacing
	blockSpacingY := blockHeight * 1.05

	// Calculate grid layout based on target number of trees and block size
	numBlocks := (targetN + 1) / 2 // Each block contains 2 trees

	// Find optimal number of blocks per row to minimize overall bounding box
	blocksPerRow, numRows := calculateOptimalLayout(numBlocks, blockSpacingX, blockSpacingY)

	// Generate all tree positions with collision checking
	trees := generateTreesWithCollisionCheck(ind.Angle, dx, dy, blocksPerRow, numRows, blockSpacingX, blockSpacingY, targetN)

	// Compact the trees: slide left and up as much as possible
	trees = compactTrees(trees, blocksPerRow)

	ind.Trees = trees

	// Fitness Calculation
	if len(trees) < targetN {
		missing := targetN - len(trees)
		ind.Score = 1000.0 + float64(missing)*10.0
	} else {
		// Check for any remaining collisions in the solution
		collisions := countCollisions(trees)
		if collisions > 0 {
			ind.Score = 500.0 + float64(collisions)*50.0
		} else {
			ind.Score = tree.CalculateScore(trees)
		}
	}
}

// countCollisions counts the number of intersecting tree pairs
func countCollisions(trees []tree.ChristmasTree) int {
	count := 0
	for i := 0; i < len(trees); i++ {
		for j := i + 1; j < len(trees); j++ {
			if trees[i].Intersect(&trees[j]) {
				count++
			}
		}
	}
	return count
}

// compactTrees slides entire blocks left (X) and entire rows up (Y).
// This exploits the grid structure: if the first tree of block N can move X,
// then all trees in blocks N and beyond can move X together.
// Similarly for rows: if the first block of row N can move Y, all rows N+ can move Y.
// Uses R-tree for efficient collision detection.
func compactTrees(trees []tree.ChristmasTree, blocksPerRow int) []tree.ChristmasTree {
	if len(trees) <= 2 || blocksPerRow <= 0 {
		return trees
	}

	step := 0.02
	maxIterations := 300

	// Make a copy to work with
	result := make([]tree.ChristmasTree, len(trees))
	copy(result, trees)

	treesPerRow := blocksPerRow * 2

	// Build R-tree for collision detection
	buildRTree := func(trees []tree.ChristmasTree) rtree.RTree {
		tr := rtree.RTree{}
		for i := range trees {
			minX, minY, maxX, maxY := trees[i].GetBoundingBox()
			tr.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, i)
		}
		return tr
	}

	// Check if a tree collides with non-affected trees using R-tree
	checkCollision := func(tr *rtree.RTree, candidate tree.ChristmasTree, affectedSet map[int]bool, allTrees []tree.ChristmasTree) bool {
		minX, minY, maxX, maxY := candidate.GetBoundingBox()
		collision := false

		tr.Search(
			[2]float64{minX, minY},
			[2]float64{maxX, maxY},
			func(min, max [2]float64, data interface{}) bool {
				j := data.(int)
				if !affectedSet[j] && candidate.Intersect(&allTrees[j]) {
					collision = true
					return false // Stop searching
				}
				return true
			},
		)
		return collision
	}

	// Compact columns (blocks) left - starting from block 1 (trees 2,3)
	for block := 1; block < blocksPerRow; block++ {
		// Get indices of all trees from this block onwards in all rows
		affectedIndices := []int{}
		affectedSet := make(map[int]bool)
		for row := 0; row*treesPerRow < len(result); row++ {
			for b := block; b < blocksPerRow; b++ {
				treeIdxA := row*treesPerRow + b*2
				treeIdxB := row*treesPerRow + b*2 + 1
				if treeIdxA < len(result) {
					affectedIndices = append(affectedIndices, treeIdxA)
					affectedSet[treeIdxA] = true
				}
				if treeIdxB < len(result) {
					affectedIndices = append(affectedIndices, treeIdxB)
					affectedSet[treeIdxB] = true
				}
			}
		}

		// Build R-tree with current positions
		tr := buildRTree(result)

		// Slide all affected trees left together
		for iter := 0; iter < maxIterations; iter++ {
			// Test if we can move all affected trees left
			canMove := true
			for _, idx := range affectedIndices {
				candidate := result[idx]
				candidate.X -= step

				if checkCollision(&tr, candidate, affectedSet, result) {
					canMove = false
					break
				}
			}

			if !canMove {
				break
			}

			// Apply the move to all affected trees
			for _, idx := range affectedIndices {
				result[idx].X -= step
			}

			// Since positions changed, rebuild R-tree
			tr = buildRTree(result)
		}
	}

	// Compact rows up - starting from row 1
	numRows := (len(result) + treesPerRow - 1) / treesPerRow
	for row := 1; row < numRows; row++ {
		// Get indices of all trees from this row onwards
		affectedIndices := []int{}
		affectedSet := make(map[int]bool)
		for r := row; r < numRows; r++ {
			startIdx := r * treesPerRow
			endIdx := startIdx + treesPerRow
			if endIdx > len(result) {
				endIdx = len(result)
			}
			for idx := startIdx; idx < endIdx; idx++ {
				affectedIndices = append(affectedIndices, idx)
				affectedSet[idx] = true
			}
		}

		// Build R-tree with current positions
		tr := buildRTree(result)

		// Slide all affected trees up together
		for iter := 0; iter < maxIterations; iter++ {
			// Test if we can move all affected trees up
			canMove := true
			for _, idx := range affectedIndices {
				candidate := result[idx]
				candidate.Y -= step

				if checkCollision(&tr, candidate, affectedSet, result) {
					canMove = false
					break
				}
			}

			if !canMove {
				break
			}

			// Apply the move to all affected trees
			for _, idx := range affectedIndices {
				result[idx].Y -= step
			}

			// Since positions changed, rebuild R-tree
			tr = buildRTree(result)
		}
	}

	return result
}

// calculateBlockDimensions returns the width and height of a single 2-tree block
func calculateBlockDimensions(angle, dx, dy float64) (float64, float64) {
	// Create the two trees of a block at origin
	tA := tree.ChristmasTree{X: 0, Y: 0, Angle: angle}
	tB := tree.ChristmasTree{X: dx, Y: dy, Angle: angle + 180.0}

	// Get bounding boxes
	minXA, minYA, maxXA, maxYA := tA.GetBoundingBox()
	minXB, minYB, maxXB, maxYB := tB.GetBoundingBox()

	// Combined bounding box
	minX := math.Min(minXA, minXB)
	minY := math.Min(minYA, minYB)
	maxX := math.Max(maxXA, maxXB)
	maxY := math.Max(maxYA, maxYB)

	return maxX - minX, maxY - minY
}

// calculateOptimalLayout determines the best number of blocks per row
func calculateOptimalLayout(numBlocks int, blockWidth, blockHeight float64) (int, int) {
	bestPerRow := 1
	bestRows := numBlocks
	bestAspectDiff := math.MaxFloat64

	for perRow := 1; perRow <= numBlocks; perRow++ {
		rows := (numBlocks + perRow - 1) / perRow
		totalWidth := float64(perRow) * blockWidth
		totalHeight := float64(rows) * blockHeight

		// We want roughly square, so minimize difference
		aspectDiff := math.Abs(totalWidth - totalHeight)

		// Also prefer configurations that minimize the max dimension
		maxDim := math.Max(totalWidth, totalHeight)

		// Combined metric
		metric := aspectDiff + maxDim*0.1

		if metric < bestAspectDiff {
			bestAspectDiff = metric
			bestPerRow = perRow
			bestRows = rows
		}
	}

	return bestPerRow, bestRows
}

// generateTreesWithCollisionCheck creates trees and verifies no collisions using R-tree
func generateTreesWithCollisionCheck(angle, dx, dy float64, blocksPerRow, numRows int, blockWidth, blockHeight float64, targetN int) []tree.ChristmasTree {
	trees := make([]tree.ChristmasTree, 0, targetN)
	tr := rtree.RTree{} // R-tree for fast collision detection
	cnt := 0

	for row := 0; row < numRows && cnt < targetN; row++ {
		baseY := float64(row) * blockHeight

		for col := 0; col < blocksPerRow && cnt < targetN; col++ {
			baseX := float64(col) * blockWidth

			// Tree A: angle = alpha
			tA := tree.ChristmasTree{
				ID:    cnt,
				X:     baseX,
				Y:     baseY,
				Angle: angle,
			}

			// Check collision with existing trees before adding
			if !checkTreeCollisionRTree(tA, trees, &tr) {
				trees = append(trees, tA)
				// Add to R-tree
				minX, minY, maxX, maxY := tA.GetBoundingBox()
				tr.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, len(trees)-1)
				cnt++
			}

			if cnt >= targetN {
				break
			}

			// Tree B: angle = alpha + 180 (rotated 180° from Tree A)
			tB := tree.ChristmasTree{
				ID:    cnt,
				X:     baseX + dx,
				Y:     baseY + dy,
				Angle: angle + 180.0,
			}

			// Check collision with existing trees before adding
			if !checkTreeCollisionRTree(tB, trees, &tr) {
				trees = append(trees, tB)
				// Add to R-tree
				minX, minY, maxX, maxY := tB.GetBoundingBox()
				tr.Insert([2]float64{minX, minY}, [2]float64{maxX, maxY}, len(trees)-1)
				cnt++
			}
		}
	}

	return trees
}

// checkTreeCollisionRTree checks if a tree collides with existing trees using R-tree
func checkTreeCollisionRTree(t tree.ChristmasTree, existing []tree.ChristmasTree, tr *rtree.RTree) bool {
	minX, minY, maxX, maxY := t.GetBoundingBox()
	collision := false

	tr.Search(
		[2]float64{minX, minY},
		[2]float64{maxX, maxY},
		func(min, max [2]float64, data interface{}) bool {
			idx := data.(int)
			if t.Intersect(&existing[idx]) {
				collision = true
				return false // Stop searching
			}
			return true
		},
	)
	return collision
}

func tournamentSelection(pop []GridIndividual) GridIndividual {
	best := pop[rand.Intn(len(pop))]
	for i := 0; i < TournamentSize-1; i++ {
		challenger := pop[rand.Intn(len(pop))]
		if challenger.Score < best.Score {
			best = challenger
		}
	}
	return best
}

func crossover(p1, p2 GridIndividual) GridIndividual {
	// Arithmetic crossover for all parameters
	alpha := rand.Float64()

	return GridIndividual{
		Angle: p1.Angle*alpha + p2.Angle*(1-alpha),
		Dx:    p1.Dx*alpha + p2.Dx*(1-alpha),
		Dy:    p1.Dy*alpha + p2.Dy*(1-alpha),
	}
}

func mutate(ind *GridIndividual) {
	// Mutate each gene with some probability
	if rand.Float64() < 0.5 {
		ind.Angle += rand.NormFloat64() * 10.0
		// Keep angle in reasonable range [0, 360)
		if ind.Angle < 0 {
			ind.Angle += 360.0
		} else if ind.Angle >= 360.0 {
			ind.Angle -= 360.0
		}
	}
	if rand.Float64() < 0.5 {
		ind.Dx += rand.NormFloat64() * 0.2
	}
	if rand.Float64() < 0.5 {
		ind.Dy += rand.NormFloat64() * 0.2
	}
}

package tree

import (
	"fmt"
	"math"
	"time"
)

// SimulatedAnnealingPenalty holds the state for the penalty-based SA solver
type SimulatedAnnealingPenalty struct {
	*SABase
}

// NewSimulatedAnnealingPenalty creates a new penalty-based SA solver
func NewSimulatedAnnealingPenalty(trees []ChristmasTree, config *SAConfig) *SimulatedAnnealingPenalty {
	return &SimulatedAnnealingPenalty{
		SABase: NewSABase(trees, config),
	}
}

// SolvePenalty runs the penalty-based simulated annealing algorithm
// All moves are allowed but penalized by overlap area
// Uses incremental overlap calculation for efficiency (only recalculates for the perturbed tree)
func (sa *SimulatedAnnealingPenalty) SolvePenalty() (float64, []ChristmasTree) {
	startTime := time.Now()

	T := sa.Config.Tmax
	currentTrees := CloneTrees(sa.Trees)

	// Calculate initial state
	currentBBox := CalculateSideLength(currentTrees)
	currentOverlap := CalculateTotalOverlap(currentTrees)
	currentScore := currentBBox + sa.Config.OverlapPenalty*currentOverlap

	bestBBoxScore := currentBBox
	bestScore := currentScore
	bestTrees := CloneTrees(currentTrees)

	// Initial best valid check
	if currentOverlap == 0 {
		bestScore = currentBBox
	}

	for step := 0; step < sa.Config.NSteps; step++ {
		for step1 := 0; step1 < sa.Config.NStepsPerT; step1++ {
			// Select random tree to perturb
			i := sa.Rng.Intn(len(currentTrees))

			// Calculate overlap BEFORE perturbation (only for tree i)
			oldTreeOverlap := CalculateTreeOverlap(currentTrees, i)

			// Perturb the tree
			oldX, oldY, oldAngle := sa.PerturbTree(&currentTrees[i])

			// Calculate overlap AFTER perturbation (only for tree i)
			newTreeOverlap := CalculateTreeOverlap(currentTrees, i)

			// Calculate new bounding box
			newBBox := CalculateSideLength(currentTrees)

			// Incremental overlap update: totalOverlap - oldContribution + newContribution
			newOverlap := currentOverlap - oldTreeOverlap + newTreeOverlap
			newScore := newBBox + sa.Config.OverlapPenalty*newOverlap

			delta := newScore - currentScore

			// Accept if better or with probability exp(-delta/T)
			if delta < 0 || sa.Rng.Float64() < math.Exp(-delta/T) {
				currentScore = newScore
				currentBBox = newBBox
				currentOverlap = newOverlap

				// Track best valid (collision-free) solution
				if newOverlap == 0 && newBBox < bestBBoxScore {
					bestBBoxScore = newBBox
					bestScore = newBBox
					bestTrees = CloneTrees(currentTrees)
					fmt.Printf("[n=%3d] NEW BEST SCORE (valid): %8.5f\n", len(currentTrees), bestBBoxScore)
				}
			} else {
				sa.RestoreTree(&currentTrees[i], oldX, oldY, oldAngle)
			}

			if step1%sa.Config.LogFreq == 0 || step1 == (sa.Config.NStepsPerT-1) {
				elapsed := FormatDuration(time.Since(startTime))
				fmt.Printf("[n=%3d] T: %.3e  Step: %6d  Score: %8.5f  Overlap: %6.4f  Best: %8.5f  Time: %s\n",
					len(currentTrees), T, step1, currentScore, currentOverlap, bestBBoxScore, elapsed)
			}
		}

		T = sa.CoolTemperature(T, step)
	}

	return bestScore, bestTrees
}

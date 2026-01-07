package sa

import (
	"fmt"
	"math"
	"time"

	"tree-packing-challenge/pkg/tree"
)

// SimulatedAnnealing holds the state for the collision-free SA solver
type SimulatedAnnealing struct {
	*Base
}

// NewSimulatedAnnealing creates a new collision-free SA solver
func NewSimulatedAnnealing(trees []tree.ChristmasTree, config *Config) *SimulatedAnnealing {
	return &SimulatedAnnealing{
		Base: NewBase(trees, config),
	}
}

// Solve runs the collision-free simulated annealing algorithm
// Moves that cause collisions are rejected
func (sa *SimulatedAnnealing) Solve() (float64, []tree.ChristmasTree) {
	startTime := time.Now()

	T := sa.Config.Tmax
	currentTrees := CloneTrees(sa.Trees)
	currentScore := tree.CalculateScore(currentTrees)
	bestScore := currentScore
	bestTrees := CloneTrees(currentTrees)

	for step := 0; step < sa.Config.NSteps; step++ {
		for step1 := 0; step1 < sa.Config.NStepsPerT; step1++ {
			// Select random tree to perturb
			i := sa.Rng.Intn(len(currentTrees))
			oldX, oldY, oldAngle := sa.PerturbTree(&currentTrees[i])

			// Check for collision - reject if collision detected
			if tree.HasCollision(currentTrees) {
				sa.RestoreTree(&currentTrees[i], oldX, oldY, oldAngle)
				if step1%sa.Config.LogFreq == 0 || step1 == (sa.Config.NStepsPerT-1) {
					elapsed := FormatDuration(time.Since(startTime))
					fmt.Printf("[Trees: %d]T: %.3f  Step: %6d  Score: %8.5f  Best: %8.5f  Time: %s\n",
						len(currentTrees), T, step1, currentScore, bestScore, elapsed)
				}
				continue
			}

			newScore := tree.CalculateScore(currentTrees)
			delta := newScore - currentScore

			// Accept if better or with probability exp(-delta/T)
			if delta < 0 || sa.Rng.Float64() < math.Exp(-delta/T) {
				currentScore = newScore
				if newScore < bestScore {
					bestScore = newScore
					bestTrees = CloneTrees(currentTrees)
					fmt.Printf("[n=%3d] NEW BEST SCORE: %8.5f\n", len(currentTrees), bestScore)
				}
			} else {
				sa.RestoreTree(&currentTrees[i], oldX, oldY, oldAngle)
			}

			if step1%sa.Config.LogFreq == 0 || step1 == (sa.Config.NStepsPerT-1) {
				elapsed := FormatDuration(time.Since(startTime))
				fmt.Printf("[n=%3d] T: %.3e  Step: %6d  Score: %8.5f  Best: %8.5f  Time: %s\n",
					len(currentTrees), T, step1, currentScore, bestScore, elapsed)
			}
		}

		T = sa.CoolTemperature(T, step)
	}

	return bestScore, bestTrees
}

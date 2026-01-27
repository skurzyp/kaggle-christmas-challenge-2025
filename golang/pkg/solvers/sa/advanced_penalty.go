package sa

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"tree-packing-challenge/pkg/tree"
)

// RunAdvancedSAPenalty runs the advanced Simulated Annealing optimization with penalty scoring.
// It allows overlaps but penalizes them, enabling traversal through invalid states.
func RunAdvancedSAPenalty(initialTrees []tree.ChristmasTree, config *Config) []tree.ChristmasTree {
	startTime := time.Now()
	rng := rand.New(rand.NewSource(config.RandomSeed))

	// Working copy
	cur := CloneTrees(initialTrees)
	n := len(cur)

	// Initial score
	curBBox := tree.CalculateSideLength(cur)
	curOverlap := tree.CalculateTotalOverlap(cur)
	curScore := curBBox + config.OverlapPenalty*curOverlap

	// Best VALID solution (overlap == 0)
	// Initialize with input if valid, otherwise keep best found so far
	bestValidTrees := CloneTrees(cur)
	bestValidScore := math.MaxFloat64

	if curOverlap == 0 {
		bestValidScore = curBBox
	}

	iter := config.NSteps * config.NStepsPerT
	T := config.Tmax

	// Helper to update best valid
	updateBest := func() {
		if curOverlap < 1e-9 { // Float tolerance for zero overlap
			if curBBox < bestValidScore {
				bestValidScore = curBBox
				bestValidTrees = CloneTrees(cur)
				fmt.Printf("[AdvPenalty] [n=%d] NEW BEST VALID: %.5f\n", n, bestValidScore)
			}
		}
	}

	updateBest()

	for it := 0; it < iter; it++ {
		mt := rng.Intn(11) // 0-10 move types
		sc := T / config.Tmax
		if sc > 1 {
			sc = 1
		}
		// Scale factor for perturbations

		accepted := false

		// Variables to store undo info for single/pair moves
		var undoIdx []int
		var undoTrees []tree.ChristmasTree

		switch mt {
		case 0: // Translate random
			i := rng.Intn(n)
			undoIdx = []int{i}
			undoTrees = []tree.ChristmasTree{cur[i]}

			cur[i].X += rng.NormFloat64() * 0.5 * sc
			cur[i].Y += rng.NormFloat64() * 0.5 * sc

		case 1: // Move towards center
			i := rng.Intn(n)
			undoIdx = []int{i}
			undoTrees = []tree.ChristmasTree{cur[i]}

			gx0, gy0, gx1, gy1 := tree.GetBounds(cur)
			dx := (gx0+gx1)/2.0 - cur[i].X
			dy := (gy0+gy1)/2.0 - cur[i].Y
			d := math.Sqrt(dx*dx + dy*dy)
			if d > 1e-6 {
				rf := rng.Float64()
				cur[i].X += dx / d * rf * 0.6 * sc
				cur[i].Y += dy / d * rf * 0.6 * sc
			}

		case 2: // Rotate
			i := rng.Intn(n)
			undoIdx = []int{i}
			undoTrees = []tree.ChristmasTree{cur[i]}

			cur[i].Angle += rng.NormFloat64() * 80.0 * sc
			cur[i].Angle = math.Mod(cur[i].Angle+360, 360)

		case 3: // Translate + Rotate
			i := rng.Intn(n)
			undoIdx = []int{i}
			undoTrees = []tree.ChristmasTree{cur[i]}

			rf2x := rng.Float64()*2 - 1
			rf2y := rng.Float64()*2 - 1
			rf2a := rng.Float64()*2 - 1
			cur[i].X += rf2x * 0.5 * sc
			cur[i].Y += rf2y * 0.5 * sc
			cur[i].Angle += rf2a * 60.0 * sc
			cur[i].Angle = math.Mod(cur[i].Angle+360, 360)

		case 4: // Boundary move
			boundary := tree.GetBoundary(cur)
			if len(boundary) > 0 {
				i := boundary[rng.Intn(len(boundary))]
				undoIdx = []int{i}
				undoTrees = []tree.ChristmasTree{cur[i]}

				gx0, gy0, gx1, gy1 := tree.GetBounds(cur)
				dx := (gx0+gx1)/2.0 - cur[i].X
				dy := (gy0+gy1)/2.0 - cur[i].Y
				d := math.Sqrt(dx*dx + dy*dy)
				if d > 1e-6 {
					rf := rng.Float64()
					cur[i].X += dx / d * rf * 0.7 * sc
					cur[i].Y += dy / d * rf * 0.7 * sc
				}
				rf2 := rng.Float64()*2 - 1
				cur[i].Angle += rf2 * 50.0 * sc
				cur[i].Angle = math.Mod(cur[i].Angle+360, 360)
			}
		// TODO: not fully implemented
		case 5: // Squeeze (global)
			savedCur := CloneTrees(cur)
			factor := 1.0 - rng.Float64()*0.004*sc
			gx0, gy0, gx1, gy1 := tree.GetBounds(cur)
			cx := (gx0 + gx1) / 2.0
			cy := (gy0 + gy1) / 2.0
			for i := range cur {
				cur[i].X = cx + (cur[i].X-cx)*factor
				cur[i].Y = cy + (cur[i].Y-cy)*factor
			}

			undoTrees = savedCur // using this as full backup
			undoIdx = nil        // nil implies full restore

		case 6: // Levy flight
			i := rng.Intn(n)
			undoIdx = []int{i}
			undoTrees = []tree.ChristmasTree{cur[i]}

			levy := math.Pow(rng.Float64()+0.001, -1.3) * 0.008
			rf2x := rng.Float64()*2 - 1
			rf2y := rng.Float64()*2 - 1
			cur[i].X += rf2x * levy
			cur[i].Y += rf2y * levy

		case 7: // Pair move
			if n > 1 {
				i := rng.Intn(n)
				j := (i + 1) % n
				undoIdx = []int{i, j}
				undoTrees = []tree.ChristmasTree{cur[i], cur[j]}

				rf2x := rng.Float64()*2 - 1
				rf2y := rng.Float64()*2 - 1
				dx := rf2x * 0.3 * sc
				dy := rf2y * 0.3 * sc
				cur[i].X += dx
				cur[i].Y += dy
				cur[j].X += dx
				cur[j].Y += dy
			}

		case 10: // Swap
			if n > 1 {
				i := rng.Intn(n)
				j := rng.Intn(n)
				if i != j {
					undoIdx = []int{i, j}
					undoTrees = []tree.ChristmasTree{cur[i], cur[j]}

					// Swap logic
					cur[i].X, cur[j].X = cur[j].X, cur[i].X
					cur[i].Y, cur[j].Y = cur[j].Y, cur[i].Y
					cur[i].Angle, cur[j].Angle = cur[j].Angle, cur[i].Angle
				}
			}

		default: // Small jitter
			i := rng.Intn(n)
			undoIdx = []int{i}
			undoTrees = []tree.ChristmasTree{cur[i]}

			rf2x := rng.Float64()*2 - 1
			rf2y := rng.Float64()*2 - 1
			cur[i].X += rf2x * 0.002 * sc
			cur[i].Y += rf2y * 0.002 * sc
		}

		newBBox := tree.CalculateSideLength(cur)
		newOverlap := tree.CalculateTotalOverlap(cur)

		newScore := newBBox + config.OverlapPenalty*newOverlap
		delta := newScore - curScore

		// Metropolis acceptance
		if delta < 0 || rng.Float64() < math.Exp(-delta/T) {
			accepted = true
		}

		if accepted {
			curBBox = newBBox
			curOverlap = newOverlap
			curScore = newScore
			updateBest()
		} else {
			// Revert
			if undoIdx != nil {
				for k, idx := range undoIdx {
					cur[idx] = undoTrees[k]
				}
			} else if len(undoTrees) > 0 {
				// Global revert
				cur = undoTrees // Restore full slice
			}
		}

		// Logging
		if it%config.LogFreq == 0 {
			elapsed := time.Since(startTime).Round(time.Millisecond)
			fmt.Printf("[AdvPenalty] T: %.3e  Step: %6d  Score: %8.5f  Overlap: %6.4f  BestValid: %8.5f  Time: %s\n",
				T, it, curScore, curOverlap, bestValidScore, elapsed)
		}

		// Cool temperature
		if (it+1)%config.NStepsPerT == 0 {
			step := it / config.NStepsPerT
			T = GetNextTemperature(config, T, step)
		}
	}

	return bestValidTrees
}

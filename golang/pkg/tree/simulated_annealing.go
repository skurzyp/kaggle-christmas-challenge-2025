package tree

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/tidwall/rtree"
	"gopkg.in/yaml.v3"
)

// CoolingSchedule represents the type of cooling schedule
type CoolingSchedule string

const (
	CoolingLinear      CoolingSchedule = "linear"
	CoolingExponential CoolingSchedule = "exponential"
	CoolingPolynomial  CoolingSchedule = "polynomial"
)

// SAConfig holds configuration parameters for simulated annealing
type SAConfig struct {
	Tmax          float64         `yaml:"Tmax"`
	Tmin          float64         `yaml:"Tmin"`
	NSteps        int             `yaml:"nsteps"`
	NStepsPerT    int             `yaml:"nsteps_per_T"`
	Cooling       CoolingSchedule `yaml:"cooling"`
	Alpha         float64         `yaml:"alpha"`
	N             float64         `yaml:"n"` // Polynomial exponent
	PositionDelta float64         `yaml:"position_delta"`
	AngleDelta    float64         `yaml:"angle_delta"`
	RandomSeed    int64           `yaml:"random_state"`
	LogFreq       int             `yaml:"log_freq"`
}

// LoadSAConfig loads SA configuration from a YAML file
func LoadSAConfig(path string) (*SAConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse wrapper structure (config.yaml has nested "params" key)
	var wrapper struct {
		Params SAConfig `yaml:"params"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		// Try parsing directly as SAConfig
		var config SAConfig
		if err2 := yaml.Unmarshal(data, &config); err2 != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
		return &config, nil
	}

	return &wrapper.Params, nil
}

// DefaultSAConfig returns a default SA configuration
func DefaultSAConfig() *SAConfig {
	return &SAConfig{
		Tmax:          0.0002,
		Tmin:          0.00001, // Lower floor allows more fine-tuning
		NSteps:        10,
		NStepsPerT:    100,
		Cooling:       CoolingExponential,
		Alpha:         0.99,
		N:             4,
		PositionDelta: 0.01,
		AngleDelta:    30.0,
		RandomSeed:    42,
		LogFreq:       100,
	}
}

// SimulatedAnnealing holds the state for the SA solver
type SimulatedAnnealing struct {
	Trees  []ChristmasTree
	Config *SAConfig
	rng    *rand.Rand
}

// NewSimulatedAnnealing creates a new SA solver
func NewSimulatedAnnealing(trees []ChristmasTree, config *SAConfig) *SimulatedAnnealing {
	if config == nil {
		config = DefaultSAConfig()
	}

	return &SimulatedAnnealing{
		Trees:  trees,
		Config: config,
		rng:    rand.New(rand.NewSource(config.RandomSeed)),
	}
}

// Clone creates a deep copy of a ChristmasTree
func (t *ChristmasTree) Clone() ChristmasTree {
	return ChristmasTree{
		ID:    t.ID,
		X:     t.X,
		Y:     t.Y,
		Angle: t.Angle,
	}
}

// perturbTree perturbs a tree's position and angle, returns old params
func (sa *SimulatedAnnealing) perturbTree(tree *ChristmasTree) (oldX, oldY, oldAngle float64) {
	oldX, oldY, oldAngle = tree.X, tree.Y, tree.Angle

	dx := (sa.rng.Float64()*2 - 1) * sa.Config.PositionDelta
	dy := (sa.rng.Float64()*2 - 1) * sa.Config.PositionDelta
	dAngle := sa.rng.NormFloat64() * sa.Config.AngleDelta

	tree.X += dx
	tree.Y += dy
	tree.Angle += dAngle
	// Normalize to [-180, 180]
	for tree.Angle > 180 {
		tree.Angle -= 360
	}
	for tree.Angle <= -180 {
		tree.Angle += 360
	}

	return oldX, oldY, oldAngle
}

// restoreTree restores a tree to its previous position
func (sa *SimulatedAnnealing) restoreTree(tree *ChristmasTree, x, y, angle float64) {
	tree.X = x
	tree.Y = y
	tree.Angle = angle
}

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

// formatDuration formats a duration in a readable format
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// Solve runs the simulated annealing algorithm
// GetCollidingTrees returns indices of trees that collide with the given tree (excluding itself)
func GetCollidingTrees(targetIdx int, trees []ChristmasTree) []int {
	var collisions []int
	if len(trees) < 2 {
		return collisions
	}

	// Use linear scan for finding all collisions of ONE tree against others.
	target := &trees[targetIdx]
	for i := range trees {
		if i == targetIdx {
			continue
		}
		if target.Intersect(&trees[i]) {
			collisions = append(collisions, i)
		}
	}
	return collisions
}

// resolveCollision attempts to move the tree to resolve collisions
// Returns true if resolved (no collisions), false otherwise
func (sa *SimulatedAnnealing) resolveCollision(treeIdx int, currentTrees []ChristmasTree) bool {
	maxSteps := 10 // Try to resolve in 10 steps

	for step := 0; step < maxSteps; step++ {
		collisions := GetCollidingTrees(treeIdx, currentTrees)
		if len(collisions) == 0 {
			return true
		}

		// Calculate repulsion vector
		var moveX, moveY float64
		targetTree := &currentTrees[treeIdx]

		for _, otherIdx := range collisions {
			otherTree := &currentTrees[otherIdx]

			// Vector from other -> target
			dx := targetTree.X - otherTree.X
			dy := targetTree.Y - otherTree.Y

			// If centers are identical, pick random direction
			if math.Abs(dx) < 1e-6 && math.Abs(dy) < 1e-6 {
				dx = sa.rng.Float64()*2 - 1
				dy = sa.rng.Float64()*2 - 1
			}

			// Normalize
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 0 {
				moveX += dx / dist
				moveY += dy / dist
			}
		}

		// Normalize total movement vector
		totalDist := math.Sqrt(moveX*moveX + moveY*moveY)
		if totalDist > 0 {
			// Move by a small step
			stepSize := sa.Config.PositionDelta
			targetTree.X += (moveX / totalDist) * stepSize
			targetTree.Y += (moveY / totalDist) * stepSize
		} else {
			// Random kick if stuck
			targetTree.X += (sa.rng.Float64()*2 - 1) * sa.Config.PositionDelta
			targetTree.Y += (sa.rng.Float64()*2 - 1) * sa.Config.PositionDelta
		}
	}

	// Check one last time
	if len(GetCollidingTrees(treeIdx, currentTrees)) > 0 {
		return false
	}
	return true
}

// Solve runs the simulated annealing algorithm
func (sa *SimulatedAnnealing) Solve() (float64, []ChristmasTree) {
	startTime := time.Now()

	T := sa.Config.Tmax

	// Deep copy current trees - this is the state we will mutate
	currentTrees := make([]ChristmasTree, len(sa.Trees))
	for i := range sa.Trees {
		currentTrees[i] = sa.Trees[i].Clone()
	}

	currentScore := CalculateScore(currentTrees)
	bestScore := currentScore

	// Deep copy best trees - this stores the best solution ever found
	// We need a separate copy because currentTrees will continue to change (heat up/cool down)
	// and might move away from the optimum.
	bestTrees := make([]ChristmasTree, len(currentTrees))
	for i := range currentTrees {
		bestTrees[i] = currentTrees[i].Clone()
	}

	for step := 0; step < sa.Config.NSteps; step++ {
		for step1 := 0; step1 < sa.Config.NStepsPerT; step1++ {
			// Select random tree to perturb
			i := sa.rng.Intn(len(currentTrees))

			// Save state before perturbation
			oldX, oldY, oldAngle := currentTrees[i].X, currentTrees[i].Y, currentTrees[i].Angle

			// 1. Perturb Angle ONLY (as requested)
			dAngle := sa.rng.NormFloat64() * sa.Config.AngleDelta
			currentTrees[i].Angle += dAngle

			// Normalize to [-180, 180]
			for currentTrees[i].Angle > 180 {
				currentTrees[i].Angle -= 360
			}
			for currentTrees[i].Angle <= -180 {
				currentTrees[i].Angle += 360
			}

			// 2. Resolve Collisions (shift position if needed)
			resolved := true
			// Check if rotation caused collision (cheaper check for just this tree)
			if len(GetCollidingTrees(i, currentTrees)) > 0 {
				resolved = sa.resolveCollision(i, currentTrees)
			}

			if !resolved {
				// Could not resolve, revert
				sa.restoreTree(&currentTrees[i], oldX, oldY, oldAngle)
				continue
			}

			newScore := CalculateScore(currentTrees)
			delta := newScore - currentScore

			// Accept if better or with probability exp(-delta/T)
			if delta < 0 || sa.rng.Float64() < math.Exp(-delta/T) {
				currentScore = newScore
				if newScore < bestScore {
					bestScore = newScore
					for j := range currentTrees {
						bestTrees[j] = currentTrees[j].Clone()
					}
					fmt.Printf("[n=%3d] NEW BEST SCORE: %8.5f\n", len(currentTrees), bestScore)
				}
			} else {
				// Revert if rejected by Metropolis criterion
				sa.restoreTree(&currentTrees[i], oldX, oldY, oldAngle)
			}

			if step1%sa.Config.LogFreq == 0 || step1 == (sa.Config.NStepsPerT-1) {
				elapsed := formatDuration(time.Since(startTime))
				fmt.Printf("[n=%3d] T: %.3e  Step: %6d  Score: %8.5f  Best: %8.5f  Time: %s\n",
					len(currentTrees), T, step1, currentScore, bestScore, elapsed)
			}
		}

		// Cool temperature
		switch sa.Config.Cooling {
		case CoolingLinear:
			T -= (sa.Config.Tmax - sa.Config.Tmin) / float64(sa.Config.NSteps)
		case CoolingExponential:
			Tfactor := -math.Log(sa.Config.Tmax / sa.Config.Tmin)
			T = sa.Config.Tmax * math.Exp(Tfactor*float64(step+1)/float64(sa.Config.NSteps))
		case CoolingPolynomial:
			progress := float64(sa.Config.NSteps-step-1) / float64(sa.Config.NSteps)
			T = sa.Config.Tmin + (sa.Config.Tmax-sa.Config.Tmin)*math.Pow(progress, sa.Config.N)
		}
	}

	return bestScore, bestTrees
}

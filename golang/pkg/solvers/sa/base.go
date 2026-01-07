// Package sa provides simulated annealing algorithms and shared infrastructure
// for solving tree packing optimization problems.
package sa

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"tree-packing-challenge/pkg/tree"
)

// Base provides shared functionality for SA algorithm variants
type Base struct {
	Trees  []tree.ChristmasTree
	Config *Config
	Rng    *rand.Rand
}

// NewBase creates a new base SA solver with shared setup
func NewBase(trees []tree.ChristmasTree, config *Config) *Base {
	if config == nil {
		config = DefaultConfig()
	}

	return &Base{
		Trees:  trees,
		Config: config,
		Rng:    rand.New(rand.NewSource(config.RandomSeed)),
	}
}

// PerturbTree perturbs a tree's position and angle, returns old params
func (sa *Base) PerturbTree(t *tree.ChristmasTree) (oldX, oldY, oldAngle float64) {
	oldX, oldY, oldAngle = t.X, t.Y, t.Angle

	dx := (sa.Rng.Float64()*2 - 1) * sa.Config.PositionDelta
	dy := (sa.Rng.Float64()*2 - 1) * sa.Config.PositionDelta
	// Gaussian-distributed angle perturbation, clamped to [-180, 180]
	dAngle := sa.Rng.NormFloat64() * sa.Config.AngleDelta
	dAngle = math.Max(-180, math.Min(180, dAngle))

	t.X += dx
	t.Y += dy
	t.Angle = math.Mod(t.Angle+dAngle+360, 360)

	return oldX, oldY, oldAngle
}

// RestoreTree restores a tree to its previous position
func (sa *Base) RestoreTree(t *tree.ChristmasTree, x, y, angle float64) {
	t.X = x
	t.Y = y
	t.Angle = angle
}

// CoolTemperature applies the cooling schedule and returns the new temperature
func (sa *Base) CoolTemperature(T float64, step int) float64 {
	switch sa.Config.Cooling {
	case CoolingLinear:
		return T - (sa.Config.Tmax-sa.Config.Tmin)/float64(sa.Config.NSteps)
	case CoolingExponential:
		Tfactor := -math.Log(sa.Config.Tmax / sa.Config.Tmin)
		return sa.Config.Tmax * math.Exp(Tfactor*float64(step+1)/float64(sa.Config.NSteps))
	case CoolingPolynomial:
		progress := float64(sa.Config.NSteps-step-1) / float64(sa.Config.NSteps)
		return sa.Config.Tmin + (sa.Config.Tmax-sa.Config.Tmin)*math.Pow(progress, sa.Config.N)
	}
	return T
}

// CloneTrees creates a deep copy of a slice of trees
func CloneTrees(trees []tree.ChristmasTree) []tree.ChristmasTree {
	cloned := make([]tree.ChristmasTree, len(trees))
	for i := range trees {
		cloned[i] = trees[i].Clone()
	}
	return cloned
}

// FormatDuration formats a duration in a readable format
func FormatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

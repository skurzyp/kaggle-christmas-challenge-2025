package tree

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// SABase provides shared functionality for SA algorithm variants
type SABase struct {
	Trees  []ChristmasTree
	Config *SAConfig
	Rng    *rand.Rand
}

// NewSABase creates a new base SA solver with shared setup
func NewSABase(trees []ChristmasTree, config *SAConfig) *SABase {
	if config == nil {
		config = DefaultSAConfig()
	}

	return &SABase{
		Trees:  trees,
		Config: config,
		Rng:    rand.New(rand.NewSource(config.RandomSeed)),
	}
}

// PerturbTree perturbs a tree's position and angle, returns old params
func (sa *SABase) PerturbTree(tree *ChristmasTree) (oldX, oldY, oldAngle float64) {
	oldX, oldY, oldAngle = tree.X, tree.Y, tree.Angle

	dx := (sa.Rng.Float64()*2 - 1) * sa.Config.PositionDelta
	dy := (sa.Rng.Float64()*2 - 1) * sa.Config.PositionDelta
	// Gaussian-distributed angle perturbation, clamped to [-180, 180]
	dAngle := sa.Rng.NormFloat64() * sa.Config.AngleDelta
	dAngle = math.Max(-180, math.Min(180, dAngle))

	tree.X += dx
	tree.Y += dy
	tree.Angle = math.Mod(tree.Angle+dAngle+360, 360)

	return oldX, oldY, oldAngle
}

// RestoreTree restores a tree to its previous position
func (sa *SABase) RestoreTree(tree *ChristmasTree, x, y, angle float64) {
	tree.X = x
	tree.Y = y
	tree.Angle = angle
}

// CoolTemperature applies the cooling schedule and returns the new temperature
func (sa *SABase) CoolTemperature(T float64, step int) float64 {
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
func CloneTrees(trees []ChristmasTree) []ChristmasTree {
	cloned := make([]ChristmasTree, len(trees))
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

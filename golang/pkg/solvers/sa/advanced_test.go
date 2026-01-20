package sa

import (
	"math/rand"
	"testing"

	"tree-packing-challenge/pkg/tree"
)

func TestSqueeze(t *testing.T) {
	// Create two trees that are far apart
	trees := []tree.ChristmasTree{
		{ID: 1, X: 0, Y: 0, Angle: 0},
		{ID: 2, X: 100, Y: 100, Angle: 0},
	}

	// Squeeze should bring them closer
	squeezed := Squeeze(trees)

	// Check if bounds are smaller
	_, _, gx1Orig, _ := tree.GetBounds(trees)
	_, _, gx1New, _ := tree.GetBounds(squeezed)

	origSide := gx1Orig // Assuming started at 0,0
	newSide := gx1New

	if newSide > origSide {
		t.Errorf("Squeeze failed to reduce bounds: got %f, want < %f", newSide, origSide)
	}
}

func TestCompaction(t *testing.T) {
	// Create trees loosely packed
	trees := []tree.ChristmasTree{
		{ID: 1, X: 0, Y: 0, Angle: 0},
		{ID: 2, X: 10, Y: 10, Angle: 0},
		{ID: 3, X: 20, Y: 20, Angle: 0},
	}

	compacted := Compaction(trees, 100)

	origSide := tree.Side(trees)
	newSide := tree.Side(compacted)

	if newSide > origSide {
		t.Errorf("Compaction failed to improve or maintain side: got %f, want <= %f", newSide, origSide)
	}
}

func TestRunAdvancedSA(t *testing.T) {
	// Setup small problem
	trees := []tree.ChristmasTree{
		{ID: 1, X: 0, Y: 0, Angle: 0},
		{ID: 2, X: 5, Y: 5, Angle: 0},
	}

	// Run SA for a few iterations
	conf := &Config{
		Tmax:       1.0,
		Tmin:       0.1,
		RandomSeed: 42,
		NSteps:     2,
		NStepsPerT: 5,
		Cooling:    CoolingExponential,
	}
	result := RunAdvancedSA(trees, conf)

	if len(result) != len(trees) {
		t.Errorf("Expected %d trees, got %d", len(trees), len(result))
	}

	if tree.AnyOvl(result) {
		t.Errorf("RunAdvancedSA returned invalid solution with overlaps")
	}
}

func TestPerturbAdvanced(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	trees := []tree.ChristmasTree{
		{ID: 1, X: 0, Y: 0, Angle: 0},
		{ID: 2, X: 5, Y: 5, Angle: 0},
	}

	perturbed := PerturbAdvanced(trees, 1.0, rng)

	if len(perturbed) != len(trees) {
		t.Errorf("PerturbAdvanced changed number of trees")
	}

	// It's possible for Perturb to return original if overlaps can't be resolved,
	// so we mainly check for basic validity (no panics, correct count).
}

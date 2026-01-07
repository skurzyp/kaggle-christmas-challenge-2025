// Package tree defines the core data structures for the Christmas tree packing challenge.
package tree

// ChristmasTree represents a single tree with position and rotation
type ChristmasTree struct {
	ID    int
	X, Y  float64
	Angle float64 // Angle in DEGREES (Kaggle submission format)
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

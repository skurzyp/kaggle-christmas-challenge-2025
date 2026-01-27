package sa

import (
	"math"
	"math/rand"

	"tree-packing-challenge/pkg/tree"
)

// Squeeze reduces the bounding box size as long as no overlaps occur
func Squeeze(trees []tree.ChristmasTree) []tree.ChristmasTree {
	c := CloneTrees(trees)
	gx0, gy0, gx1, gy1 := tree.GetBounds(c)
	cx := (gx0 + gx1) / 2.0
	cy := (gy0 + gy1) / 2.0

	// Try scaling down from 0.9995 to 0.98
	for scale := 0.9995; scale >= 0.98; scale -= 0.0005 {
		trial := CloneTrees(c)
		for i := range trial {
			trial[i].X = cx + (c[i].X-cx)*scale
			trial[i].Y = cy + (c[i].Y-cy)*scale
		}

		if !tree.AnyOvl(trial) {
			c = trial // Accepted
		} else {
			break // Cannot squeeze further
		}
	}
	return c
}

// Compaction attempts to move trees towards the center to reduce bounds
func Compaction(trees []tree.ChristmasTree, iters int) []tree.ChristmasTree {
	c := CloneTrees(trees)
	bs := tree.Side(c)

	for it := 0; it < iters; it++ {
		gx0, gy0, gx1, gy1 := tree.GetBounds(c)
		cx := (gx0 + gx1) / 2.0
		cy := (gy0 + gy1) / 2.0
		improved := false

		for i := range c {
			ox, oy := c[i].X, c[i].Y
			dx := cx - c[i].X
			dy := cy - c[i].Y
			d := math.Sqrt(dx*dx + dy*dy)

			if d < 1e-6 {
				continue
			}

			// Try different step sizes
			steps := []float64{0.02, 0.008, 0.003, 0.001, 0.0004}
			for _, step := range steps {
				c[i].X = ox + dx/d*step
				c[i].Y = oy + dy/d*step

				if !tree.HasOvl(c, i) {
					newSide := tree.Side(c)
					if newSide < bs-1e-12 {
						bs = newSide
						improved = true
						ox, oy = c[i].X, c[i].Y // Update current best pos for this tree
					} else {
						// Revert if no improvement in global score
						c[i].X, c[i].Y = ox, oy
					}
				} else {
					// Revert if overlap
					c[i].X, c[i].Y = ox, oy
				}
			}
		}

		if !improved {
			break
		}
	}

	return c
}

// LocalSearch performs local optimization by small moves and rotations
func LocalSearch(trees []tree.ChristmasTree, maxIter int) []tree.ChristmasTree {
	c := CloneTrees(trees)
	bs := tree.Side(c)

	steps := []float64{0.01, 0.004, 0.0015, 0.0006, 0.00025, 0.0001}
	rots := []float64{5.0, 2.0, 0.8, 0.3, 0.1}
	dxDir := []float64{1, -1, 0, 0, 1, 1, -1, -1}
	dyDir := []float64{0, 0, 1, -1, 1, -1, 1, -1}

	for iter := 0; iter < maxIter; iter++ {
		improved := false
		for i := range c {
			gx0, gy0, gx1, gy1 := tree.GetBounds(c)
			cx := (gx0 + gx1) / 2.0
			cy := (gy0 + gy1) / 2.0

			ddx := cx - c[i].X
			ddy := cy - c[i].Y
			dist := math.Sqrt(ddx*ddx + ddy*ddy)

			// Move towards center
			if dist > 1e-6 {
				for _, st := range steps {
					ox, oy := c[i].X, c[i].Y
					c[i].X += ddx / dist * st
					c[i].Y += ddy / dist * st

					if !tree.HasOvl(c, i) {
						newSide := tree.Side(c)
						if newSide < bs-1e-12 {
							bs = newSide
							improved = true
						} else {
							c[i].X, c[i].Y = ox, oy
						}
					} else {
						c[i].X, c[i].Y = ox, oy
					}
				}
			}

			// Move in 8 directions
			for _, st := range steps {
				for d := 0; d < 8; d++ {
					ox, oy := c[i].X, c[i].Y
					c[i].X += dxDir[d] * st
					c[i].Y += dyDir[d] * st

					if !tree.HasOvl(c, i) {
						newSide := tree.Side(c)
						if newSide < bs-1e-12 {
							bs = newSide
							improved = true
						} else {
							c[i].X, c[i].Y = ox, oy
						}
					} else {
						c[i].X, c[i].Y = ox, oy
					}
				}
			}

			// Rotate
			for _, rt := range rots {
				for _, da := range []float64{rt, -rt} {
					oa := c[i].Angle
					c[i].Angle += da
					c[i].Angle = math.Mod(c[i].Angle, 360)
					if c[i].Angle < 0 {
						c[i].Angle += 360
					}

					if !tree.HasOvl(c, i) {
						newSide := tree.Side(c)
						if newSide < bs-1e-12 {
							bs = newSide
							improved = true
						} else {
							c[i].Angle = oa
						}
					} else {
						c[i].Angle = oa
					}
				}
			}
		}
		if !improved {
			break
		}
	}
	return c
}

// PerturbAdvanced perturbs the configuration based on strength
func PerturbAdvanced(trees []tree.ChristmasTree, str float64, rng *rand.Rand) []tree.ChristmasTree {
	c := CloneTrees(trees)
	original := CloneTrees(trees) // Keep original in case we fail to fix overlaps

	n := len(c)
	np := int(float64(n)*0.08 + str*3.0)
	if np < 1 {
		np = 1
	}

	for k := 0; k < np; k++ {
		i := rng.Intn(n)
		c[i].X += rng.NormFloat64() * str * 0.5
		c[i].Y += rng.NormFloat64() * str * 0.5
		c[i].Angle += rng.NormFloat64() * 30.0
		c[i].Angle = math.Mod(c[i].Angle+360, 360)
	}

	// Try to fix overlaps
	for iter := 0; iter < 150; iter++ {
		fixed := true
		for i := 0; i < n; i++ {
			if tree.HasOvl(c, i) {
				fixed = false
				gx0, gy0, gx1, gy1 := tree.GetBounds(c)
				cx := (gx0 + gx1) / 2.0
				cy := (gy0 + gy1) / 2.0

				dx := c[i].X - cx
				dy := c[i].Y - cy
				d := math.Sqrt(dx*dx + dy*dy)
				if d > 1e-6 {
					c[i].X += dx / d * 0.02
					c[i].Y += dy / d * 0.02
				}
				// random float between -1 and 1
				rf2 := rng.Float64()*2 - 1
				c[i].Angle += rf2 * 15.0
				c[i].Angle = math.Mod(c[i].Angle+360, 360)
			}
		}
		if fixed {
			break
		}
	}

	if tree.AnyOvl(c) {
		return original
	}
	return c
}

// RunAdvancedSA runs the advanced Simulated Annealing optimization
func RunAdvancedSA(initialTrees []tree.ChristmasTree, config *Config) []tree.ChristmasTree {
	rng := rand.New(rand.NewSource(config.RandomSeed))
	c := CloneTrees(initialTrees)
	best := CloneTrees(c)
	cur := CloneTrees(c)

	bs := tree.Side(best)
	cs := bs
	T := config.Tmax
	noImp := 0

	n := len(c)
	if n == 0 {
		return c
	}

	iter := config.NSteps * config.NStepsPerT

	// Track total steps for cooling schedule
	step := 0
	for it := 0; it < iter; it++ {
		step++
		mt := rng.Intn(11) // 0-10 move types
		sc := T / config.Tmax
		valid := true
		savedCur := CloneTrees(cur) // Save state before mutation

		// Select move type
		switch mt {
		case 0:
			i := rng.Intn(n)
			cur[i].X += rng.NormFloat64() * 0.5 * sc
			cur[i].Y += rng.NormFloat64() * 0.5 * sc
			if tree.HasOvl(cur, i) {
				valid = false
			}
		case 1:
			i := rng.Intn(n)
			gx0, gy0, gx1, gy1 := tree.GetBounds(cur)
			dx := (gx0+gx1)/2.0 - cur[i].X
			dy := (gy0+gy1)/2.0 - cur[i].Y
			d := math.Sqrt(dx*dx + dy*dy)
			if d > 1e-6 {
				rf := rng.Float64()
				cur[i].X += dx / d * rf * 0.6 * sc
				cur[i].Y += dy / d * rf * 0.6 * sc
			}
			if tree.HasOvl(cur, i) {
				valid = false
			}
		case 2:
			i := rng.Intn(n)
			cur[i].Angle += rng.NormFloat64() * 80.0 * sc
			cur[i].Angle = math.Mod(cur[i].Angle+360, 360)
			if tree.HasOvl(cur, i) {
				valid = false
			}
		case 3:
			i := rng.Intn(n)
			rf2x := rng.Float64()*2 - 1
			rf2y := rng.Float64()*2 - 1
			rf2a := rng.Float64()*2 - 1
			cur[i].X += rf2x * 0.5 * sc
			cur[i].Y += rf2y * 0.5 * sc
			cur[i].Angle += rf2a * 60.0 * sc
			cur[i].Angle = math.Mod(cur[i].Angle+360, 360)
			if tree.HasOvl(cur, i) {
				valid = false
			}
		case 4:
			boundary := tree.GetBoundary(cur)
			if len(boundary) > 0 {
				i := boundary[rng.Intn(len(boundary))]
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
				if tree.HasOvl(cur, i) {
					valid = false
				}
			} else {
				valid = false
			}
		case 5:
			factor := 1.0 - rng.Float64()*0.004*sc
			gx0, gy0, gx1, gy1 := tree.GetBounds(cur)
			cx := (gx0 + gx1) / 2.0
			cy := (gy0 + gy1) / 2.0

			// We need to apply this to all trees, so we can't just modify cur in place without backup
			for i := range cur {
				cur[i].X = cx + (cur[i].X-cx)*factor
				cur[i].Y = cy + (cur[i].Y-cy)*factor
			}
			if tree.AnyOvl(cur) {
				valid = false
			}
		case 6:
			i := rng.Intn(n)
			// Levy flight: pow(rng + 0.001, -1.3) * 0.008
			levy := math.Pow(rng.Float64()+0.001, -1.3) * 0.008
			rf2x := rng.Float64()*2 - 1
			rf2y := rng.Float64()*2 - 1
			cur[i].X += rf2x * levy
			cur[i].Y += rf2y * levy
			if tree.HasOvl(cur, i) {
				valid = false
			}
		case 7:
			if n > 1 {
				i := rng.Intn(n)
				j := (i + 1) % n
				rf2x := rng.Float64()*2 - 1
				rf2y := rng.Float64()*2 - 1
				dx := rf2x * 0.3 * sc
				dy := rf2y * 0.3 * sc
				cur[i].X += dx
				cur[i].Y += dy
				cur[j].X += dx
				cur[j].Y += dy
				if tree.HasOvl(cur, i) || tree.HasOvl(cur, j) {
					valid = false
				}
			}
			// Case 8, 9 to be implemented
		case 10:
			if n > 1 {
				i := rng.Intn(n)
				j := rng.Intn(n)
				if !tree.SwapTrees(cur, i, j) {
					valid = false
				}
			}
		default: // mt 8, 9 fall here
			i := rng.Intn(n)
			rf2x := rng.Float64()*2 - 1
			rf2y := rng.Float64()*2 - 1
			cur[i].X += rf2x * 0.002
			cur[i].Y += rf2y * 0.002
			if tree.HasOvl(cur, i) {
				valid = false
			}
		}

		if !valid {
			cur = savedCur // Revert
			noImp++

			// Cool temperature if step reached
			if (it+1)%config.NStepsPerT == 0 {
				step := it / config.NStepsPerT
				T = GetNextTemperature(config, T, step)
			}
			continue
		}

		// Calculate stats
		ns := tree.Side(cur)
		delta := ns - cs

		if delta < 0 || rng.Float64() < math.Exp(-delta/T) {
			cs = ns
			if ns < bs {
				bs = ns
				best = CloneTrees(cur)
				noImp = 0
			} else {
				noImp++
			}
		} else {
			cur = CloneTrees(best) // Reset to best
			cs = bs
			noImp++
		}

		// Cool temperature
		if (it+1)%config.NStepsPerT == 0 {
			step := it / config.NStepsPerT
			T = GetNextTemperature(config, T, step)
		}
	}

	return best
}

# Tree Packing Challenge - Go Implementation

Greedy algorithm implementation for the tree packing challenge in Go.

## Structure

- **`main.go`** - Entry point, handles CSV output
- **`tree.go`** - Christmas tree geometry and collision detection
- **`algorithm.go`** - Greedy placement algorithm with spatial indexing
- **`go.mod`** - Dependencies

## Dependencies

- [`github.com/solarlune/resolv`](https://github.com/solarlune/resolv) - 2D collision detection using SAT
- [`github.com/tidwall/rtree`](https://github.com/tidwall/rtree) - Spatial index for fast collision queries

## How It Works

The algorithm:

1. **Progressive Packing**: Builds configurations from 1 to 200 trees, reusing previous placements
2. **Greedy Placement**: For each new tree:
   - Try 10 random angles (weighted by `abs(sin(2θ))` to favor corners)
   - Start at radius 20 and move inward until collision
   - Back off slightly to find valid placement
   - Keep the placement with smallest radius
3. **Collision Detection**: 
   - Trees decomposed into 4 convex parts (trunk + 3 tiers)
   - R-tree spatial index for fast bounding box queries
   - SAT (Separating Axis Theorem) for precise collision checks
4. **Output**: CSV with tree positions for each configuration

## Running

```bash
# Build
go build -o tree-packing .

# Run
./tree-packing
```

Output: `sample_submission.csv`

## Algorithm Details

**Tree Shape**: Each tree consists of 4 convex polygons:
- Trunk (rectangle)
- Bottom tier (trapezoid)
- Middle tier (trapezoid)
- Top tier (triangle)

**Weighted Angle**: Trees placed along vectors with angle θ where P(θ) ∝ |sin(2θ)|, creating a less circular packing.

**Spatial Optimization**: R-tree index reduces collision checks from O(n) to O(log n) on average.
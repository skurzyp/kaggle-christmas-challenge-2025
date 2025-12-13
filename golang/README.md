# Tree Packing Challenge - Go Implementation

Greedy and Simulated Annealing algorithms for the Kaggle tree packing challenge.

## Structure

```
golang/
├── cmd/packer/main.go         # CLI entry point
├── pkg/tree/
│   ├── tree.go                # Christmas tree geometry
│   ├── algorithm.go           # Greedy placement algorithm
│   └── simulated_annealing.go # SA optimization
├── sa_config.yaml             # SA configuration file
└── go.mod
```

## Quick Start

```bash
# Build
cd golang
go build -o packer ./cmd/packer

# Run with greedy algorithm (default)
./packer -algorithm greedy -n 200 -output submission.csv

# Run with simulated annealing
./packer -algorithm sa -config sa_config.yaml -n 200 -output submission.csv
```

## CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-algorithm` | `greedy` | Algorithm: `greedy` or `sa` |
| `-config` | *(none)* | Path to SA config YAML file |
| `-n` | `200` | Number of trees to pack |
| `-output` | `submission.csv` | Output CSV file path |
| `-seed` | `0` | Random seed (0 = use current time) |

## Algorithms

### Greedy Placement
1. Progressive packing from 1 to N trees
2. For each new tree: try 10 random angles, move inward until collision
3. R-tree spatial index for O(log n) collision queries

### Simulated Annealing
1. Start with greedy solution
2. Perturb random tree's position/angle
3. Accept better solutions or worse ones with probability exp(-Δ/T)
4. Cool temperature using linear/exponential/polynomial schedule

## SA Configuration

Edit `sa_config.yaml`:

```yaml
params:
  Tmax: 0.0002          # Starting temperature
  Tmin: 0.00005         # Final temperature
  nsteps: 15            # Outer temperature steps
  nsteps_per_T: 500     # Inner iterations per temperature
  cooling: "exponential"  # linear, exponential, polynomial
  position_delta: 0.01  # Position perturbation range
  angle_delta: 30.0     # Angle perturbation range (degrees)
  random_state: 42
  log_freq: 250
```

## Dependencies

- [`github.com/engelsjk/polygol`](https://github.com/engelsjk/polygol) - Polygon intersection
- [`github.com/tidwall/rtree`](https://github.com/tidwall/rtree) - Spatial index
- [`gopkg.in/yaml.v3`](https://gopkg.in/yaml.v3) - YAML config parsing
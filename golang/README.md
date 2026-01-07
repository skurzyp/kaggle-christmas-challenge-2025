# Tree Packing Challenge - Go Implementation

Greedy, Grid, and Simulated Annealing algorithms for the Kaggle tree packing challenge.

## Structure

```
golang/
├── cmd/packer/main.go           # CLI entry point
├── cmd/packer/main.go           # CLI entry point
├── pkg/
│   ├── tree/                    # Domain model
│   │   ├── model.go             # ChristmasTree struct
│   │   ├── geometry.go          # Geometry calculations
│   │   ├── intersection.go      # Intersection logic
│   │   ├── defaults.go          # Constants
│   │   ├── ops.go               # Tree operations (Overlap, Bounds)
│   │   └── evaluation.go        # Scoring functions
│   └── solvers/                 # Optimization algorithms
│       ├── greedy/              # Greedy placement
│       ├── grid/                # Grid-based placement
│       └── sa/                  # Simulated Annealing variants
├── sa_config.yaml               # SA configuration file
└── go.mod
```

## Quick Start

### Linux / macOS

```bash
# Build
cd golang
go build -o packer ./cmd/packer

# Run with greedy algorithm (default)
./packer -algorithm greedy -n 200 -output submission.csv

# Run with simulated annealing
./packer -algorithm sa -config sa_config.yaml -n 200 -output submission.csv

# Run with grid placement
./packer -algorithm grid -n 200 -output submission.csv

# Run with grid + SA optimization
./packer -algorithm grid-sa -n 200 -output submission.csv

# Run with penalty-based SA
./packer -algorithm sa-penalty -config sa_config.yaml -n 200 -output submission.csv

# Run with advanced SA
./packer -algorithm sa-advanced -config sa_config.yaml -n 200 -output submission.csv

# Run with advanced SA with Penalty
./packer -algorithm sa-advanced-penalty -config sa_config.yaml -n 200 -output submission.csv

# Run with grid + penalty-based SA
./packer -algorithm grid-sa-penalty -config sa_config.yaml -n 200 -output submission.csv
```

### Windows (PowerShell)

```powershell
# Build
cd golang
go build -o packer.exe ./cmd/packer

# Run with greedy algorithm (default)
.\packer.exe -algorithm greedy -n 200 -output submission.csv

# Run with simulated annealing
.\packer.exe -algorithm sa -config sa_config.yaml -n 200 -output submission.csv

# Run with grid placement
.\packer.exe -algorithm grid -n 200 -output submission.csv

# Run with grid + SA optimization
.\packer.exe -algorithm grid-sa -n 200 -output submission.csv
```

## CLI Flags

| Flag         | Default                                    | Description                                     |
| ------------ | ------------------------------------------ | ----------------------------------------------- |
| `-algorithm` | `greedy`                                   | `greedy`, `sa`, `sa-penalty`, `sa-advanced`, `grid`, `grid-sa`, `grid-sa-penalty`, `sa-advanced-penalty` |
| `-config`    | _(none)_                                   | Path to SA config YAML file                     |
| `-n`         | `200`                                      | Number of trees to pack                         |
| `-output`    | `../../results/submissions/submission.csv` | Output CSV file path                            |
| `-seed`      | `0`                                        | Random seed (0 = use current time)              |

## Algorithms

### Greedy Placement (`pkg/solvers/greedy/greedy.go`)

1. Progressive packing from 1 to N trees
2. For each new tree: try 10 random angles, move inward until collision
3. R-tree spatial index for O(log n) collision queries

### Grid Placement (`pkg/solvers/grid/grid.go`)

1. Places trees in alternating rows with staggered X offsets
2. Even rows: angle 0°, odd rows: angle 180° (inverted trees)
3. Horizontal spacing: 0.7 units, odd row X offset: 0.35
4. Tries different row configurations to find optimal packing

### Simulated Annealing - Collision Free (`pkg/solvers/sa/collision_free.go`)

1. Start with greedy or grid solution
2. Perturb random tree's position/angle
3. **Reject moves that cause collisions**
4. Accept better solutions or worse ones with probability exp(-Δ/T)
5. Cool temperature using linear/exponential/polynomial schedule

### Simulated Annealing - Penalty Based (`pkg/solvers/sa/penalty.go`)

1. Start with greedy or grid solution
2. Perturb random tree's position/angle
3. **Allow overlaps but penalize by overlap area**: `Score = BoundingBox + λ × OverlapArea`
4. Accept moves based on Metropolis criterion
5. Track best **valid** (collision-free) solution found

## SA Configuration

Edit `sa_config.yaml`:

```yaml
params:
  Tmax: 0.0002 # Starting temperature
  Tmin: 0.00005 # Final temperature
  nsteps: 15 # Outer temperature steps
  nsteps_per_T: 500 # Inner iterations per temperature
  cooling: "exponential" # linear, exponential, polynomial
  position_delta: 0.01 # Position perturbation range
  angle_delta: 30.0 # Angle perturbation range (degrees)
  random_state: 42
  log_freq: 250
  overlap_penalty: 10.0 # λ for penalty-based SA
```

## Dependencies

- [`github.com/engelsjk/polygol`](https://github.com/engelsjk/polygol) - Polygon intersection
- [`github.com/tidwall/rtree`](https://github.com/tidwall/rtree) - Spatial index
- [`gopkg.in/yaml.v3`](https://gopkg.in/yaml.v3) - YAML config parsing

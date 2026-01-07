# Solvers Overview

This repository contains several solver implementations for the Christmas Tree Packing Challenge. The solvers are located in `golang/pkg/solvers`.

## 1. Greedy Solver (`pkg/solvers/greedy`)

The **Greedy Solver** places trees sequentially, attempting to minimize the distance from the center.

- **Strategy**: 
    - For each tree, it generates candidate positions using random angles and distances.
    - It uses a **Monte Carlo** approach, trying 10 random starts and moving them inwards towards the center until a collision occurs or the center is reached.
    - It selects the valid position closest to the center.
- **Key Features**:
    - **Spatial Indexing**: Uses `rtree` for efficient collision detection.
    - **Weighted Angles**: Generates angles with a bias using `sin(2*angle)` distribution.

## 2. Grid Solver (`pkg/solvers/grid`)

The **Grid Solver** places trees in a structured, lattice-like pattern.

- **Strategy**:
    - Organizes trees in rows.
    - **Alternating Orientation**: Even rows have trees at 0°, Odd rows at 180°.
    - **Staggered Layout**: Odd rows are offset horizontally to fit into the gaps of even rows.
- **Optimization**:
    - Iterates through all possible combinations of trees per even/odd row to find the configuration that yields the smallest bounding box.

## 3. Simulated Annealing (`pkg/solvers/sa`)

The **Simulated Annealing (SA)** package provides a robust framework for optimization with several variants.

### Base Infrastructure
- Provides shared configuration (`Config`), random number generation, and basic tree perturbation/restoration utilities.
- Implements different cooling schedules: Linear, Exponential, and Polynomial.

### Variants

#### A. Collision-Free SA (`collision_free.go`)
- **Logic**: Standard SA where **any move causing an overlap is immediately rejected**.
- **Use Case**: Good for maintaining a strictly valid state throughout, but can get stuck in local optima easily due to the strict validity constraint.

#### B. Penalty-Based SA (`penalty.go`)
- **Logic**: Allows trees to overlap during the process.
- **Score Function**: `Score = SideLength + (PenaltyFactor * TotalOverlapArea)`.
- **Mechanism**:
    - Moves are accepted based on the Metropolis criterion applied to the penalized score.
    - **Incremental Updates**: efficiently calculates change in overlap area to speed up iterations.
    - Tracks the **Best Valid Solution** (0 overlap) found during the process.
- **Benefit**: Can traverse "invalid" states to reach better "valid" states, avoiding local optima better than the collision-free approach.

#### C. Advanced SA (`advanced.go`)
- **Logic**: Implements complex move sets and specialized operators to explore the solution space more effectively.
- **Operators**:
    1.  **Squeeze**: Linearly scales down the entire packing until a collision occurs.
    2.  **Compaction**: Individually moves trees towards the center with varying step sizes.
    3.  **Local Search**: Tries small moves and rotations for each tree to escape local optima.
    4.  **Perturb**: Applies stronger random changes to shake up the configuration.
- **Move Types**: The `RunAdvancedSA` function utilizes 11 different move types, including:
    - Random translations/rotations.
    - Moving towards the center.
    - Boundary-focused moves.
    - "Levy flight" (long-distance jumps).
    - Pairwise moves (moving two trees together).
    - Swapping two trees.

#### D. Mixed Strategy SA (`advanced_penalty.go`)
- **Overview**: This solver combines the robust move sets of the **Advanced SA** with the flexibility of the **Penalty-Based SA**.
- **Logic**: 
    - It uses the sophisticated mutation operators (squeeze, compaction, levy flights, etc.) from the Advanced SA.
    - It evaluates states using the penalty function from the Penalty SA, allowing it to temporarily accept invalid states (overlaps) to traverse through barriers in the search landscape.
- **Benefit**: This "mixed" approach offers the highest potential for finding global optima by leveraging both advanced exploration techniques and the ability to tunnel through invalid regions.

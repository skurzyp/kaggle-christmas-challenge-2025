# Mutation Operators

This document describes the various mutation operators and move types used in the Advanced Simulated Annealing solver (`pkg/solvers/sa/advanced.go`). These operators are designed to explore the search space effectively by combining local exploitation with global exploration.

## High-Level Operators

These are complex, compound operations that may affect multiple trees or run for multiple iterations.

### 1. Squeeze
- **Description**: Linearly scales down the positions of **all trees** towards the center of the packing.
- **Mechanism**: Iteratively reduces the scale factor (e.g., from 0.9995 down to 0.98) and checks for collisions. It keeps the smallest scaling that is still valid (collision-free).
- **Goal**: To rapidly tighten a loose packing.

### 2. Compaction
- **Description**: Individually attempts to move each tree towards the geometric center of the packing.
- **Mechanism**: 
    - Iterates through all trees.
    - Tries to move each tree towards the center using a sequence of decreasing step sizes (e.g., 0.02 down to 0.0004).
    - Accepts the move if it improves the bounding box and maintains validity.
- **Goal**: To remove gaps in the packing by "gravitating" trees inwards.

### 3. Local Search
- **Description**: Fine-tunes the position and rotation of each tree to escape local optima.
- **Mechanism**:
    - For each tree, tries:
        - Moving towards the center.
        - Moving in 8 cardinal/intercardinal directions.
        - Rotating by small angles (e.g., ±5°, ±2°... down to ±0.1°).
    - Uses a series of decreasing step sizes.
    - Moves are accepted only if they improve the solution and do not cause collisions.

### 4. Perturb (Advanced)
- **Description**: Applies a "kick" to the system to escape deep local optima.
- **Mechanism**:
    - Selects a subset of trees (approx 8% + strength factor).
    - Applies random Gaussian noise to their X, Y positions and Angle.
    - Attempts to resolve any resulting overlaps using a short localized optimization loop.
    - Reverts to the original state if overlaps cannot be resolved.

## Atomic Move Types (Transition Kernel)

The `RunAdvancedSA` function utilizes a probabilistic transition kernel that selects one of the following move types at each step.

| ID | Name | Description |
| :--- | :--- | :--- |
| **0** | **Random Translation** | Moves a single random tree by a random offset (Gaussian distribution). |
| **1** | **Center Bias** | Moves a single random tree specifically **towards the center** of the bounding box. effective for compaction. |
| **2** | **Rotation** | Rotates a single random tree by a random angle. |
| **3** | **Translate + Rotate** | Applies both random translation and rotation to a single tree simultaneously. |
| **4** | **Boundary Move** | Selects a tree that is on the **boundary** of the packing (defining the current bounding box) and moves it inwards. This directly targets the objective function. |
| **5** | **Global Squeeze** | Scales the positions of **all trees** towards the center. Similar to the Squeeze operator but applied as a single atomic step in the SA chain. |
| **6** | **Levy Flight** | Moves a single tree by a distance drawn from a **Levy distribution** (heavy-tailed). This allows for occasional long jumps to explore distant regions of the box. |
| **7** | **Pair Move** | Selects two adjacent trees (in the list) and moves them together by the same offset. Helps preserve local substructures. |
| **10** | **Swap** | Swaps the positions and rotations of two randomly selected trees. Useful for changing the relative topology of the packing. |
| **Def**| **Small Jitter** | (Default case) Applies a very small random translation to a single tree. Used for fine-tuning. |

**Note**: Move probabilities are uniform in the current implementation (random int 0-10), though some case handlers might reject invalid moves or valid moves depending on the SA logic (Collision-Free vs Penalty).

### Move 0: Random Translation
- **Description**: Moves a single random tree by a random offset.
- **Mechanism**:
    - Selects a random tree index `i`.
    - Updates `X` and `Y` by adding Gaussian noise (`NormFloat64`) scaled by `0.5 * sc` (where `sc` is the current temperature scale factor).
    - Effectiveness decreases as temperature cools.

### Move 1: Center Bias
- **Description**: Moves a single random tree specifically **towards the center** of the bounding box.
- **Mechanism**:
    - Calculates the center of the current bounding box (`gx0, gy0, gx1, gy1`).
    - Determines the vector from the tree's current position to the center.
    - Moves the tree along this vector by a factor of `0.6 * sc` multiplied by a random float.
    - Highly effective for "packing" functionality.

### Move 2: Rotation
- **Description**: Rotates a single random tree by a random angle.
- **Mechanism**:
    - Selects a random tree.
    - Adds Gaussian noise to the `Angle` scaled by `80.0 * sc`.
    - Normalizes the result to `[0, 360)`.

### Move 3: Translate + Rotate
- **Description**: Applies both random translation and rotation to a single tree simultaneously.
- **Mechanism**:
    - Selects a random tree.
    - Translates `X` and `Y` by uniform random values (`Float64()*2 - 1`) scaled by `0.5 * sc`.
    - Rotates `Angle` by uniform random value scaled by `60.0 * sc`.
    - Provides a more chaotic perturbation than separate moves.

### Move 4: Boundary Move
- **Description**: Target trees that define the current bounding box.
- **Mechanism**:
    - Identifies trees on the boundary using `tree.GetBoundary(cur)`.
    - Selects one boundary tree at random.
    - Moves it towards the center of the packing, scaled by `0.7 * sc`.
    - Also applies a random rotation scaled by `50.0 * sc`.
    - Directly attempts to reduce the objective function (side length).

### Move 5: Global Squeeze
- **Description**: Scales the positions of **all trees** towards the center.
- **Mechanism**:
    - Calculates the dataset center.
    - Multiplies the distance of **every tree** from the center by a factor slightly less than 1 (`1.0 - rng * 0.004 * sc`).
    - Effectively shrinks the entire layout.
    - In collision-free mode, this is often rejected if tight, but useful in penalty mode.

### Move 6: Levy Flight
- **Description**: Moves a single tree by a distance drawn from a heavy-tailed distribution.
- **Mechanism**:
    - Calculates a step size using `Pow(rng + 0.001, -1.3) * 0.008`.
    - This generates mostly small moves but occasionally very large jumps ("flights").
    - Moves the tree in a random direction `(rf2x, rf2y)` by this step size.
    - Helps trees "teleport" out of crowded areas to open spaces.

### Move 7: Pair Move
- **Description**: Moves two adjacent trees (in the list) together.
- **Mechanism**:
    - Selects tree `i` and `j = (i + 1) % n`.
    - Generates a single random translation vector scaled by `0.3 * sc`.
    - Applies this vector to **both** trees.
    - Preserves the relative spatial relationship between the pair if they are close (though list adjacency doesn't guarantee spatial proximity, it often correlates in initial stages or if pre-sorted).

### Move 10: Swap
- **Description**: Swaps the configuration of two trees.
- **Mechanism**:
    - Selects two distinct random trees `i` and `j`.
    - Swaps their `X`, `Y`, and `Angle` values entirely.
    - Useful for topological changes (e.g., swapping a large tree in the center with a small tree on the edge).

### Default: Small Jitter (Moves 8, 9)
- **Description**: Applies a tiny fine-tuning translation.
- **Mechanism**:
    - Fallback for unused move IDs (8, 9).
    - Translates a random tree by a very small uniform amount (`0.002` constant, not scaled by temperature).
    - Acts as a "shake" to settle trees into better local fits.

import shapely
print(f'Using shapely {shapely.__version__}')

import os
from decimal import Decimal, getcontext
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
from matplotlib.patches import Rectangle
from shapely import affinity, touches


from tqdm import tqdm
from concurrent.futures import ProcessPoolExecutor

# Set precision for Decimal
getcontext().prec = 25
scale_factor = Decimal('1e15')

class ChristmasTree:
    """Represents a single, rotatable Christmas tree of a fixed size."""

    def __init__(self, center_x='0', center_y='0', angle='0'):
        """Initializes the Christmas tree with a specific position and rotation."""
        self.center_x = Decimal(center_x)
        self.center_y = Decimal(center_y)
        self.angle = Decimal(angle)

        trunk_w = Decimal('0.15')
        trunk_h = Decimal('0.2')
        base_w = Decimal('0.7')
        mid_w = Decimal('0.4')
        top_w = Decimal('0.25')
        tip_y = Decimal('0.8')
        tier_1_y = Decimal('0.5')
        tier_2_y = Decimal('0.25')
        base_y = Decimal('0.0')
        trunk_bottom_y = -trunk_h

        initial_polygon = Polygon(
            [
                # Start at Tip
                (Decimal('0.0') * scale_factor, tip_y * scale_factor),
                # Right side - Top Tier
                (top_w / Decimal('2') * scale_factor, tier_1_y * scale_factor),
                (top_w / Decimal('4') * scale_factor, tier_1_y * scale_factor),
                # Right side - Middle Tier
                (mid_w / Decimal('2') * scale_factor, tier_2_y * scale_factor),
                (mid_w / Decimal('4') * scale_factor, tier_2_y * scale_factor),
                # Right side - Bottom Tier
                (base_w / Decimal('2') * scale_factor, base_y * scale_factor),
                # Right Trunk
                (trunk_w / Decimal('2') * scale_factor, base_y * scale_factor),
                (trunk_w / Decimal('2') * scale_factor, trunk_bottom_y * scale_factor),
                # Left Trunk
                (-(trunk_w / Decimal('2')) * scale_factor, trunk_bottom_y * scale_factor),
                (-(trunk_w / Decimal('2')) * scale_factor, base_y * scale_factor),
                # Left side - Bottom Tier
                (-(base_w / Decimal('2')) * scale_factor, base_y * scale_factor),
                # Left side - Middle Tier
                (-(mid_w / Decimal('4')) * scale_factor, tier_2_y * scale_factor),
                (-(mid_w / Decimal('2')) * scale_factor, tier_2_y * scale_factor),
                # Left side - Top Tier
                (-(top_w / Decimal('4')) * scale_factor, tier_1_y * scale_factor),
                (-(top_w / Decimal('2')) * scale_factor, tier_1_y * scale_factor),
            ]
        )
        rotated = affinity.rotate(initial_polygon, float(self.angle), origin=(0, 0))
        self.polygon = affinity.translate(rotated,
                                          xoff=float(self.center_x * scale_factor),
                                          yoff=float(self.center_y * scale_factor))

from shapely.strtree import STRtree
from shapely.geometry import Polygon

def find_best_trees_with_collision(n: int) -> tuple[float, list[ChristmasTree]]:
    best_score, best_trees = float("inf"), None

    for n_even in range(1, n + 1):
        for n_odd in [n_even, n_even - 1]:
            all_trees = []
            rest = n
            r = 0
            while rest > 0:
                m = min(rest, n_even if r % 2 == 0 else n_odd)
                rest -= m

                angle = 0 if r % 2 == 0 else 180
                x_offset = 0 if r % 2 == 0 else Decimal("0.7") / 2
                y = r // 2 * Decimal("1.0") if r % 2 == 0 else (Decimal("0.8") + (r - 1) // 2 * Decimal("1.0"))

                row_trees = []
                # Tworzymy STRtree tylko z Polygonów
                placed_polygons = [t.polygon for t in all_trees if isinstance(t.polygon, Polygon)]
                tree_index = STRtree(placed_polygons) if placed_polygons else None

                for i in range(m):
                    candidate_tree = ChristmasTree(center_x=Decimal("0.7") * i + x_offset, center_y=y, angle=angle)
                    candidate_poly = candidate_tree.polygon

                    # upewniamy się, że candidate_poly jest Polygonem
                    if not isinstance(candidate_poly, Polygon):
                        continue

                    # szybkie sprawdzenie kolizji przy użyciu STRtree
                    collision = False
                    if tree_index:
                        possible_hits = tree_index.query(candidate_poly)
                        collision = any(
                            isinstance(p, Polygon) and candidate_poly.intersects(p) and not candidate_poly.touches(p)
                            for p in possible_hits
                        )

                    if collision:
                        # jeśli kolizja → nie dodawaj tego drzewa do rzędu
                        continue

                    row_trees.append(candidate_tree)

                all_trees.extend(row_trees)
                r += 1

            # jeśli w tym układzie udało się postawić wszystkie drzewa
            if len(all_trees) != n:
                continue

            # policz score
            xys = np.concatenate([np.asarray(t.polygon.exterior.xy).T / 1e15 for t in all_trees])
            min_x, min_y = xys.min(axis=0)
            max_x, max_y = xys.max(axis=0)
            score = max(max_x - min_x, max_y - min_y) ** 2

            if score < best_score:
                best_score = score
                best_trees = all_trees

    return best_score, best_trees



def plot_grid_solution(n, score, trees):
    """Plot arrangement produced by find_best_trees()."""
    fig, ax = plt.subplots(figsize=(7, 7))

    colors = plt.cm.viridis(np.linspace(0, 1, len(trees)))

    for i, tree in enumerate(trees):
        x, y = tree.polygon.exterior.xy
        x = np.array(x) / 1e15
        y = np.array(y) / 1e15
        ax.fill(x, y, alpha=0.5, color=colors[i])
        ax.plot(x, y, color='black', linewidth=0.5)

    # compute bounding box
    all_xy = np.concatenate([np.asarray(t.polygon.exterior.xy).T / 1e15 for t in trees])
    min_x, min_y = all_xy.min(axis=0)
    max_x, max_y = all_xy.max(axis=0)

    side = max(max_x - min_x, max_y - min_y)

    # draw bounding square
    ax.add_patch(
        Rectangle(
            (min_x, min_y),
            side,
            side,
            fill=False,
            edgecolor='red',
            linewidth=2,
            linestyle='--'
        )
    )

    ax.set_aspect("equal")
    ax.set_title(f"{n} trees | score={score:.6f}")
    ax.axis("off")

    os.makedirs("plots", exist_ok=True)
    plt.savefig(f"plots/{n:03d}_trees.png", dpi=150)
    plt.close()

solutions = []


with ProcessPoolExecutor(max_workers=4) as executor:
    for n, (score, trees) in enumerate(tqdm(executor.map(find_best_trees_with_collision, range(1, 201)), total=200), start=1):
        solutions.append((score, trees))


overall_score = sum(score / n for n, (score, _) in enumerate(solutions, 1))
print(f'Overall score: {overall_score:.6f}')


def to_str(x: Decimal):
    return f"s{round(float(x), 6)}"

rows = []
for n, (_, all_trees) in enumerate(solutions, 1):
    assert len(all_trees) == n
    for i_t, tree in enumerate(all_trees):
        rows.append({
            "id": f"{n:03d}_{i_t}",
            "x": to_str(tree.center_x),
            "y": to_str(tree.center_y),
            "deg": to_str(tree.angle),
        })



df = pd.DataFrame(rows)
df.to_csv("submission.csv", index=False)

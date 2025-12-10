"""
Visualization script for the Tree Packing Challenge sample submission.
Displays Christmas trees from the sample_submission.csv file.
"""

import matplotlib
matplotlib.use('Agg')  # Non-interactive backend for headless rendering

import math
import os
from decimal import Decimal, getcontext

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
from matplotlib.patches import Rectangle
from shapely import affinity
from shapely.geometry import Polygon
from shapely.ops import unary_union

pd.set_option('display.float_format', '{:.12f}'.format)

# Set precision for Decimal
getcontext().prec = 25
scale_factor = Decimal('1e15')

# Build the index of the submission, in the format:
#  <trees_in_problem>_<tree_index>
index = [f'{n:03d}_{t}' for n in range(1, 201) for t in range(n)]


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


def load_submission(filepath: str, num_trees: int) -> pd.DataFrame:
    """
    Load and process the sample submission for a specific number of trees.
    Returns a DataFrame with x, y, deg columns (without 's' prefix).
    """
    df = pd.read_csv(filepath, index_col='id')
    
    # Filter for the specific tree count
    prefix = f"{num_trees:03d}_"
    mask = df.index.str.startswith(prefix)
    subset = df[mask].copy()
    
    # Remove 's' prefix and convert to float
    for col in ['x', 'y', 'deg']:
        subset[col] = subset[col].str.replace('s', '').astype(str)
    
    return subset


def plot_trees(num_trees: int, submission_path: str):
    """Plot the tree arrangement for a given number of trees."""
    # Load submission data
    df = load_submission(submission_path, num_trees)
    
    if df.empty:
        print(f"No data found for {num_trees} trees")
        return
    
    # Create figure
    fig, ax = plt.subplots(figsize=(10, 10))
    colors = plt.cm.viridis(np.linspace(0, 1, len(df)))
    
    placed_trees = []
    
    for i, (idx, row) in enumerate(df.iterrows()):
        # Create tree using the proper ChristmasTree class
        tree = ChristmasTree(
            center_x=row['x'],
            center_y=row['y'],
            angle=row['deg']
        )
        placed_trees.append(tree)
    
    # Get bounding box
    all_polygons = [t.polygon for t in placed_trees]
    bounds = unary_union(all_polygons).bounds
    
    # Plot each tree
    for i, tree in enumerate(placed_trees):
        # Rescale for plotting
        x_scaled, y_scaled = tree.polygon.exterior.xy
        x = [Decimal(val) / scale_factor for val in x_scaled]
        y = [Decimal(val) / scale_factor for val in y_scaled]
        ax.plot(x, y, color=colors[i])
        ax.fill(x, y, alpha=0.5, color=colors[i])
    
    minx = Decimal(bounds[0]) / scale_factor
    miny = Decimal(bounds[1]) / scale_factor
    maxx = Decimal(bounds[2]) / scale_factor
    maxy = Decimal(bounds[3]) / scale_factor

    width = maxx - minx
    height = maxy - miny
    side_length = max(width, height)

    square_x = minx if width >= height else minx - (side_length - width) / 2
    square_y = miny if height >= width else miny - (side_length - height) / 2
    
    bounding_square = Rectangle(
        (float(square_x), float(square_y)),
        float(side_length),
        float(side_length),
        fill=False,
        edgecolor='red',
        linewidth=2,
        linestyle='--',
    )
    ax.add_patch(bounding_square)

    padding = Decimal('0.5')
    ax.set_xlim(
        float(square_x - padding),
        float(square_x + side_length + padding))
    ax.set_ylim(
        float(square_y - padding),
        float(square_y + side_length + padding))
    ax.set_aspect('equal', adjustable='box')
    ax.axis('off')
    plt.title(f'{num_trees} Trees: {side_length:.12f}')
    
    plt.tight_layout()
    plt.savefig(f'trees_{num_trees}.png', dpi=150, bbox_inches='tight', facecolor='white')
    print(f"Saved: trees_{num_trees}.png")
    plt.close()


def main():
    """Main function to visualize sample submissions."""
    submission_path = '../golang/sample_submission.csv'
    
    # Visualize different tree counts
    tree_counts = [10, 20, 50, 100]
    
    for n in tree_counts:
        print(f"Plotting {n} trees...")
        try:
            plot_trees(n, submission_path)
        except Exception as e:
            print(f"Error plotting {n} trees: {e}")
            import traceback
            traceback.print_exc()


if __name__ == "__main__":
    main()

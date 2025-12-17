# Python Visualization

This directory contains the Python logic for the Tree Packing Challenge, primarily for plotting results.

## Prerequisites

- **Python 3.13+**
- **Poetry**

## Setup

1.  Navigate to this directory:
    ```bash
    cd python
    ```

2.  Install dependencies using Poetry:
    ```bash
    poetry install
    ```

## How to Run

To run the main script:

```bash
poetry run python main.py
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `../results/submissions/submission.csv` | Path to submission CSV file |
| `-n, --trees` | `10 12 14 16 18` | Number of trees to visualize (space-separated) |

### Examples

```bash
# Use default submission path
poetry run python main.py

# Specify custom submission file
poetry run python main.py -o ../results/submissions/my-submission.csv

# Specify tree counts to visualize
poetry run python main.py -n 5 10 15 20
```

## Development

### Formatting

We use `black` for code formatting. To format your code:

```bash
poetry run black .
```
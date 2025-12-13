package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"tree-packing-challenge/pkg/tree"
)

func main() {
	// CLI flags
	algorithm := flag.String("algorithm", "greedy", "Algorithm to use: 'greedy' or 'sa'")
	configPath := flag.String("config", "", "Path to SA config YAML file (optional, uses defaults if not provided)")
	numTrees := flag.Int("n", 200, "Number of trees to pack")
	output := flag.String("output", "submission.csv", "Output CSV file path")
	seed := flag.Int64("seed", 0, "Random seed (0 = use current time)")

	flag.Parse()

	// Set random seed
	if *seed == 0 {
		rand.Seed(time.Now().UnixNano())
	} else {
		rand.Seed(*seed)
	}

	fmt.Printf("Tree Packing - Algorithm: %s, Trees: %d\n", *algorithm, *numTrees)

	var treeData [][]string

	switch *algorithm {
	case "greedy":
		treeData = runGreedy(*numTrees)
	case "sa":
		treeData = runSimulatedAnnealing(*numTrees, *configPath)
	default:
		fmt.Fprintf(os.Stderr, "Unknown algorithm: %s (use 'greedy' or 'sa')\n", *algorithm)
		os.Exit(1)
	}

	// Write CSV output
	if err := writeCSV(*output, treeData); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done! Output written to: %s\n", *output)
}

// runGreedy runs the greedy placement algorithm
func runGreedy(numTrees int) [][]string {
	var currentPlacedTrees []tree.ChristmasTree
	var treeData [][]string

	// Pack trees progressively from 1 to numTrees
	for n := 1; n <= numTrees; n++ {
		currentPlacedTrees, _ = tree.InitializeTrees(n, currentPlacedTrees)
		if n%10 == 0 {
			fmt.Printf("Greedy: Packed %d trees\n", n)
		}

		// Record each tree's position for this configuration
		for tIdx, t := range currentPlacedTrees {
			treeData = append(treeData, formatTree(n, tIdx, t))
		}
	}

	return treeData
}

// runSimulatedAnnealing runs SA optimization on each tree count
func runSimulatedAnnealing(numTrees int, configPath string) [][]string {
	// Load config
	var config *tree.SAConfig
	var err error

	if configPath != "" {
		config, err = tree.LoadSAConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v, using defaults\n", err)
			config = tree.DefaultSAConfig()
		}
	} else {
		config = tree.DefaultSAConfig()
	}

	var treeData [][]string
	var currentTrees []tree.ChristmasTree

	// For each configuration size
	for n := 1; n <= numTrees; n++ {
		// Initialize with greedy first
		currentTrees, _ = tree.InitializeTrees(n, currentTrees)

		// Run SA to optimize
		sa := tree.NewSimulatedAnnealing(currentTrees, config)
		bestScore, bestTrees := sa.Solve()

		// Use best trees as starting point for next iteration
		currentTrees = bestTrees

		fmt.Printf("SA: n=%d, score=%.5f\n", n, bestScore)

		// Record each tree's position for this configuration
		for tIdx, t := range bestTrees {
			treeData = append(treeData, formatTree(n, tIdx, t))
		}
	}

	return treeData
}

// formatTree formats a tree for CSV output
func formatTree(n, idx int, t tree.ChristmasTree) []string {
	return []string{
		fmt.Sprintf("%03d_%d", n, idx),
		fmt.Sprintf("s%.6f", t.X),
		fmt.Sprintf("s%.6f", t.Y),
		fmt.Sprintf("s%.6f", t.Angle),
	}
}

// writeCSV writes tree data to a CSV file
func writeCSV(path string, data [][]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"id", "x", "y", "deg"}); err != nil {
		return err
	}

	return writer.WriteAll(data)
}

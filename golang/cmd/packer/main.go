package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"tree-packing-challenge/pkg/tree"
)

// saResult holds the result of a single SA run for a specific n
type saResult struct {
	n        int
	score    float64
	trees    []tree.ChristmasTree
	treeData [][]string
}

func main() {
	// CLI flags
	algorithm := flag.String("algorithm", "greedy", "Algorithm to use: 'greedy', 'sa', 'grid', or 'grid-sa'")
	configPath := flag.String("config", "", "Path to SA config YAML file (optional, uses defaults if not provided)")
	numTrees := flag.Int("n", 200, "Number of trees to pack")
	output := flag.String("output", "../../results/submissions/submission.csv", "Output CSV file path")
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
	case "grid":
		treeData = runGrid(*numTrees)
	case "grid-sa":
		treeData = runGridSA(*numTrees, *configPath)
	default:
		fmt.Fprintf(os.Stderr, "Unknown algorithm: %s (use 'greedy', 'sa', 'grid', or 'grid-sa')\n", *algorithm)
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

// runSimulatedAnnealing runs SA optimization on each tree count in parallel
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

	numWorkers := runtime.NumCPU()
	fmt.Printf("Running SA in parallel with %d workers\n", numWorkers)

	// Channel for jobs (tree counts to process)
	jobs := make(chan int, numTrees)
	// Channel for results
	results := make(chan saResult, numTrees)

	// Start worker pool
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Go(func() {
			for n := range jobs {
				// Initialize trees for this n (each worker creates its own)
				initialTrees, _ := tree.InitializeTrees(n, nil)

				// Run SA to optimize
				sa := tree.NewSimulatedAnnealing(initialTrees, config)
				bestScore, bestTrees := sa.Solve()

				// Format tree data for this n
				var data [][]string
				for tIdx, t := range bestTrees {
					data = append(data, formatTree(n, tIdx, t))
				}

				results <- saResult{
					n:        n,
					score:    bestScore,
					trees:    bestTrees,
					treeData: data,
				}
			}
		})
	}

	// Send all jobs
	for n := 1; n <= numTrees; n++ {
		jobs <- n
	}
	close(jobs)

	// Wait for all workers in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	var allResults []saResult
	for result := range results {
		fmt.Printf("SA: n=%d, score=%.5f\n", result.n, result.score)
		allResults = append(allResults, result)
	}

	// Sort results by n to maintain order
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].n < allResults[j].n
	})

	// Combine all tree data in order
	var treeData [][]string
	for _, result := range allResults {
		treeData = append(treeData, result.treeData...)
	}

	return treeData
}

// runGrid runs the grid-based placement algorithm (Python-style)
func runGrid(numTrees int) [][]string {
	var treeData [][]string

	// For each configuration size
	for n := 1; n <= numTrees; n++ {
		score, trees := tree.FindBestGridSolution(n)

		if n%10 == 0 {
			fmt.Printf("Grid: n=%d, score=%.5f\n", n, score)
		}

		// Record each tree's position for this configuration
		for tIdx, t := range trees {
			treeData = append(treeData, formatTree(n, tIdx, t))
		}
	}

	return treeData
}

// runGridSA runs grid-based initialization followed by SA optimization in parallel
func runGridSA(numTrees int, configPath string) [][]string {
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

	numWorkers := runtime.NumCPU()
	fmt.Printf("Running Grid+SA in parallel with %d workers\n", numWorkers)

	// Channel for jobs (tree counts to process)
	jobs := make(chan int, numTrees)
	// Channel for results
	results := make(chan saResult, numTrees)

	// Start worker pool
	var wg sync.WaitGroup
	for range numWorkers {
		wg.Go(func() {
			for n := range jobs {
				// Initialize with grid-based placement
				_, gridTrees := tree.FindBestGridSolution(n)

				// Run SA to optimize
				sa := tree.NewSimulatedAnnealing(gridTrees, config)
				bestScore, bestTrees := sa.Solve()

				// Format tree data for this n
				var data [][]string
				for tIdx, t := range bestTrees {
					data = append(data, formatTree(n, tIdx, t))
				}

				results <- saResult{
					n:        n,
					score:    bestScore,
					trees:    bestTrees,
					treeData: data,
				}
			}
		})
	}

	// Send all jobs
	for n := 1; n <= numTrees; n++ {
		jobs <- n
	}
	close(jobs)

	// Wait for all workers in a separate goroutine
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	var allResults []saResult
	for result := range results {
		fmt.Printf("Grid+SA: n=%d, score=%.5f\n", result.n, result.score)
		allResults = append(allResults, result)
	}

	// Sort results by n to maintain order
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].n < allResults[j].n
	})

	// Combine all tree data in order
	var treeData [][]string
	for _, result := range allResults {
		treeData = append(treeData, result.treeData...)
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

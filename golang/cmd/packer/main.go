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

	"tree-packing-challenge/pkg/solvers/greedy"
	"tree-packing-challenge/pkg/solvers/grid"
	"tree-packing-challenge/pkg/solvers/sa"
	"tree-packing-challenge/pkg/tree"
)

// Result holds the result of a single SA run for a specific n
type Result struct {
	N        int
	Score    float64
	Trees    []tree.ChristmasTree
	TreeData [][]string
}

// SolverFunc defines the signature for a single-instance solver
type SolverFunc func(n int, config *sa.Config) (float64, []tree.ChristmasTree)

func main() {
	// CLI flags
	algorithm := flag.String("algorithm", "greedy", "Algorithm: greedy, sa, sa-penalty, sa-advanced, grid, grid-sa, grid-sa-penalty")
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
		treeData = runSimulatedAnnealing(*numTrees, *configPath, false)
	case "sa-penalty":
		treeData = runSimulatedAnnealing(*numTrees, *configPath, true)
	case "grid":
		treeData = runGrid(*numTrees)
	case "grid-sa":
		treeData = runGridSA(*numTrees, *configPath, false)
	case "grid-sa-penalty":
		treeData = runGridSA(*numTrees, *configPath, true)
	case "sa-advanced":
		treeData = runAdvancedSA(*numTrees, *configPath)
	case "sa-advanced-penalty":
		treeData = runAdvancedSAPenalty(*numTrees, *configPath)
	default:
		fmt.Fprintf(os.Stderr, "Unknown algorithm: %s\n", *algorithm)
		os.Exit(1)
	}

	// Write CSV output
	if err := writeCSV(*output, treeData); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Done! Output written to: %s\n", *output)
}

// runParallel executes the given solver in parallel for all n from 1 to numTrees
func runParallel(numTrees int, configPath string, algoName string, solver SolverFunc) [][]string {
	config := loadConfig(configPath)
	numWorkers := runtime.NumCPU()
	fmt.Printf("Running %s in parallel with %d workers\n", algoName, numWorkers)

	jobs := make(chan int, numTrees)
	results := make(chan Result, numTrees)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Go(func() {
			for n := range jobs {
				score, trees := solver(n, config)

				var data [][]string
				for tIdx, t := range trees {
					data = append(data, formatTree(n, tIdx, t))
				}

				results <- Result{
					N:        n,
					Score:    score,
					Trees:    trees,
					TreeData: data,
				}
			}
		})
	}

	for n := 1; n <= numTrees; n++ {
		jobs <- n
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []Result
	for result := range results {
		fmt.Printf("%s: n=%d, score=%.5f\n", algoName, result.N, result.Score)
		allResults = append(allResults, result)
	}

	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].N < allResults[j].N
	})

	var treeData [][]string
	for _, result := range allResults {
		treeData = append(treeData, result.TreeData...)
	}

	return treeData
}

// runGreedy runs the greedy placement algorithm in parallel
func runGreedy(numTrees int) [][]string {
	return runParallel(numTrees, "", "Greedy", func(n int, _ *sa.Config) (float64, []tree.ChristmasTree) {
		trees, sideLength := greedy.InitializeTrees(n, nil)
		return sideLength, trees
	})
}

// loadConfig loads SA config from path or returns defaults
func loadConfig(configPath string) *sa.Config {
	if configPath != "" {
		config, err := sa.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v, using defaults\n", err)
			return sa.DefaultConfig()
		}
		return config
	}
	return sa.DefaultConfig()
}

// runSimulatedAnnealing runs SA optimization in parallel
func runSimulatedAnnealing(numTrees int, configPath string, usePenalty bool) [][]string {
	algoName := "SA"
	if usePenalty {
		algoName = "SA-Penalty"
	}

	return runParallel(numTrees, configPath, algoName, func(n int, config *sa.Config) (float64, []tree.ChristmasTree) {
		initialTrees, _ := greedy.InitializeTrees(n, nil)
		if usePenalty {
			solver := sa.NewSimulatedAnnealingPenalty(initialTrees, config)
			return solver.SolvePenalty()
		}
		solver := sa.NewSimulatedAnnealing(initialTrees, config)
		return solver.Solve()
	})
}

// runGrid runs the grid-based placement algorithm in parallel
func runGrid(numTrees int) [][]string {
	return runParallel(numTrees, "", "Grid", func(n int, _ *sa.Config) (float64, []tree.ChristmasTree) {
		score, trees := grid.FindBestSolution(n)
		return score, trees
	})
}

// runGridSA runs grid-based initialization followed by SA optimization in parallel
func runGridSA(numTrees int, configPath string, usePenalty bool) [][]string {
	algoName := "Grid+SA"
	if usePenalty {
		algoName = "Grid+SA-Penalty"
	}

	return runParallel(numTrees, configPath, algoName, func(n int, config *sa.Config) (float64, []tree.ChristmasTree) {
		_, gridTrees := grid.FindBestSolution(n)
		if usePenalty {
			solver := sa.NewSimulatedAnnealingPenalty(gridTrees, config)
			return solver.SolvePenalty()
		}
		solver := sa.NewSimulatedAnnealing(gridTrees, config)
		return solver.Solve()
	})
}

// runAdvancedSA runs the advanced SA algorithm in parallel
func runAdvancedSA(numTrees int, configPath string) [][]string {
	return runParallel(numTrees, configPath, "Advanced SA", func(n int, config *sa.Config) (float64, []tree.ChristmasTree) {
		initialTrees, _ := greedy.InitializeTrees(n, nil)
		iter := config.NSteps * config.NStepsPerT
		bestTrees := sa.RunAdvancedSA(initialTrees, iter, config.Tmax, config.Tmin, config.RandomSeed)
		return tree.CalculateScore(bestTrees), bestTrees
	})
}

// runAdvancedSAPenalty runs the advanced SA algorithm with penalty
func runAdvancedSAPenalty(numTrees int, configPath string) [][]string {
	return runParallel(numTrees, configPath, "Advanced SA Penalty", func(n int, config *sa.Config) (float64, []tree.ChristmasTree) {
		initialTrees, _ := greedy.InitializeTrees(n, nil)
		bestTrees := sa.RunAdvancedSAPenalty(initialTrees, config)
		return tree.CalculateScore(bestTrees), bestTrees
	})
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

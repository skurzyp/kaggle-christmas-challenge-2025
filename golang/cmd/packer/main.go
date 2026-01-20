package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
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
type SolverFunc func(n int, config *sa.Config, startNodes []tree.ChristmasTree) (float64, []tree.ChristmasTree)

func main() {
	// CLI flags
	algorithm := flag.String("algorithm", "greedy", "Algorithm: greedy, sa, sa-penalty, sa-advanced, grid, grid-sa, grid-sa-penalty")
	configPath := flag.String("config", "", "Path to SA config YAML file (optional, uses defaults if not provided)")
	numTrees := flag.Int("n", 200, "Number of trees to pack")
	output := flag.String("output", "../results/submissions/submission.csv", "Output CSV file path")
	seed := flag.Int64("seed", 0, "Random seed (0 = use current time)")
	startFrom := flag.String("start-from", "", "Path to submission CSV to use as starting point")

	flag.Parse()

	// Set random seed
	if *seed == 0 {
		rand.Seed(time.Now().UnixNano())
	} else {
		rand.Seed(*seed)
	}

	fmt.Printf("Tree Packing - Algorithm: %s, Trees: %d\n", *algorithm, *numTrees)

	var startingPoints map[int][]tree.ChristmasTree
	if *startFrom != "" {
		var err error
		startingPoints, err = loadStartingPoints(*startFrom)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading starting points: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Loaded starting points from %s for %d layouts\n", *startFrom, len(startingPoints))
	}

	var treeData [][]string

	switch *algorithm {
	case "greedy":
		treeData = runGreedy(*numTrees, startingPoints)
	case "sa":
		treeData = runSimulatedAnnealing(*numTrees, *configPath, false, startingPoints)
	case "sa-penalty":
		treeData = runSimulatedAnnealing(*numTrees, *configPath, true, startingPoints)
	case "grid":
		treeData = runGrid(*numTrees, startingPoints)
	case "grid-sa":
		treeData = runGridSA(*numTrees, *configPath, false, startingPoints)
	case "grid-sa-penalty":
		treeData = runGridSA(*numTrees, *configPath, true, startingPoints)
	case "sa-advanced":
		treeData = runAdvancedSA(*numTrees, *configPath, startingPoints)
	case "sa-advanced-penalty":
		treeData = runAdvancedSAPenalty(*numTrees, *configPath, startingPoints)
	case "grid-ga":
		treeData = runGridGA(*numTrees, startingPoints)
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
func runParallel(numTrees int, configPath string, algoName string, startingPoints map[int][]tree.ChristmasTree, solver SolverFunc) [][]string {
	config := loadConfig(configPath)
	numWorkers := runtime.NumCPU()
	fmt.Printf("Running %s in parallel with %d workers\n", algoName, numWorkers)

	jobs := make(chan int, numTrees)
	results := make(chan Result, numTrees)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Go(func() {
			for n := range jobs {
				var startNodes []tree.ChristmasTree
				if startingPoints != nil {
					startNodes = startingPoints[n]
				}
				score, trees := solver(n, config, startNodes)

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
func runGreedy(numTrees int, startingPoints map[int][]tree.ChristmasTree) [][]string {
	return runParallel(numTrees, "", "Greedy", startingPoints, func(n int, _ *sa.Config, _ []tree.ChristmasTree) (float64, []tree.ChristmasTree) {
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
func runSimulatedAnnealing(numTrees int, configPath string, usePenalty bool, startingPoints map[int][]tree.ChristmasTree) [][]string {
	algoName := "SA"
	if usePenalty {
		algoName = "SA-Penalty"
	}

	return runParallel(numTrees, configPath, algoName, startingPoints, func(n int, config *sa.Config, startNodes []tree.ChristmasTree) (float64, []tree.ChristmasTree) {
		var initialTrees []tree.ChristmasTree
		if len(startNodes) > 0 {
			initialTrees = startNodes // copy? usually safe to use as is if solver doesn't mutate in place blindly
		} else {
			initialTrees, _ = greedy.InitializeTrees(n, nil)
		}

		if usePenalty {
			solver := sa.NewSimulatedAnnealingPenalty(initialTrees, config)
			return solver.SolvePenalty()
		}
		solver := sa.NewSimulatedAnnealing(initialTrees, config)
		return solver.Solve()
	})
}

// runGrid runs the grid-based placement algorithm in parallel
func runGrid(numTrees int, startingPoints map[int][]tree.ChristmasTree) [][]string {
	return runParallel(numTrees, "", "Grid", startingPoints, func(n int, _ *sa.Config, startNodes []tree.ChristmasTree) (float64, []tree.ChristmasTree) {
		if len(startNodes) > 0 {
			// If provided, just evaluate them
			return tree.CalculateScore(startNodes), startNodes
		}
		score, trees := grid.FindBestSolution(n)
		return score, trees
	})
}

// runGridSA runs grid-based initialization followed by SA optimization in parallel
func runGridSA(numTrees int, configPath string, usePenalty bool, startingPoints map[int][]tree.ChristmasTree) [][]string {
	algoName := "Grid+SA"
	if usePenalty {
		algoName = "Grid+SA-Penalty"
	}

	return runParallel(numTrees, configPath, algoName, startingPoints, func(n int, config *sa.Config, startNodes []tree.ChristmasTree) (float64, []tree.ChristmasTree) {
		var gridTrees []tree.ChristmasTree
		if len(startNodes) > 0 {
			gridTrees = startNodes
		} else {
			_, gridTrees = grid.FindBestSolution(n)
		}

		if usePenalty {
			solver := sa.NewSimulatedAnnealingPenalty(gridTrees, config)
			return solver.SolvePenalty()
		}
		solver := sa.NewSimulatedAnnealing(gridTrees, config)
		return solver.Solve()
	})
}

// runAdvancedSA runs the advanced SA algorithm in parallel
func runAdvancedSA(numTrees int, configPath string, startingPoints map[int][]tree.ChristmasTree) [][]string {
	return runParallel(numTrees, configPath, "Advanced SA", startingPoints, func(n int, config *sa.Config, startNodes []tree.ChristmasTree) (float64, []tree.ChristmasTree) {
		var initialTrees []tree.ChristmasTree
		if len(startNodes) > 0 {
			initialTrees = startNodes
		} else {
			initialTrees, _ = greedy.InitializeTrees(n, nil)
		}

		bestTrees := sa.RunAdvancedSA(initialTrees, config)
		return tree.CalculateScore(bestTrees), bestTrees
	})
}

// runAdvancedSAPenalty runs the advanced SA algorithm with penalty
func runAdvancedSAPenalty(numTrees int, configPath string, startingPoints map[int][]tree.ChristmasTree) [][]string {
	return runParallel(numTrees, configPath, "Advanced SA Penalty", startingPoints, func(n int, config *sa.Config, startNodes []tree.ChristmasTree) (float64, []tree.ChristmasTree) {
		var initialTrees []tree.ChristmasTree
		if len(startNodes) > 0 {
			initialTrees = startNodes
		} else {
			initialTrees, _ = greedy.InitializeTrees(n, nil)
		}
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
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

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

// runGridGA runs the genetic algorithm grid placement in parallel
func runGridGA(numTrees int, startingPoints map[int][]tree.ChristmasTree) [][]string {
	return runParallel(numTrees, "", "Grid GA", startingPoints, func(n int, _ *sa.Config, _ []tree.ChristmasTree) (float64, []tree.ChristmasTree) {
		score, trees := grid.FindBestGridGASolution(n)
		return score, trees
	})
}

func loadStartingPoints(path string) (map[int][]tree.ChristmasTree, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	result := make(map[int][]tree.ChristmasTree)
	startIdx := 0
	if len(records) > 0 && len(records[0]) > 0 && strings.ToLower(records[0][0]) == "id" {
		startIdx = 1
	}

	for _, record := range records[startIdx:] {
		if len(record) < 4 {
			continue
		}
		// Parse ID: "005_2" -> N=5
		parts := strings.Split(record[0], "_")
		if len(parts) != 2 {
			continue
		}
		n, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		// Parse X, Y, Angle (remove 's')
		parseVal := func(s string) float64 {
			s = strings.TrimPrefix(s, "s")
			v, _ := strconv.ParseFloat(s, 64)
			return v
		}

		t := tree.ChristmasTree{
			ID:    len(result[n]) + 1,
			X:     parseVal(record[1]),
			Y:     parseVal(record[2]),
			Angle: parseVal(record[3]),
		}
		result[n] = append(result[n], t)
	}
	return result, nil
}

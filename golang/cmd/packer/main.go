package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"tree-packing-challenge/pkg/tree"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Starting Tree Packing...")

	var currentPlacedTrees []tree.ChristmasTree
	var treeData [][]string

	// Pack trees progressively from 1 to 200
	for n := 1; n <= 200; n++ {
		currentPlacedTrees, _ = tree.InitializeTrees(n, currentPlacedTrees)
		if n%10 == 0 {
			fmt.Printf("Packed %d trees\n", n)
		}

		// Record each tree's position for this configuration
		for tIdx, t := range currentPlacedTrees {
			idStr := fmt.Sprintf("%03d_%d", n, tIdx)
			xStr := fmt.Sprintf("s%.6f", t.X)
			yStr := fmt.Sprintf("s%.6f", t.Y)
			degStr := fmt.Sprintf("s%.6f", t.Angle)

			treeData = append(treeData, []string{idStr, xStr, yStr, degStr})
		}
	}

	// Write CSV output to results folder
	outputPath := filepath.Join("..", "results", "submissions", "sample_submission.csv")
	file, err := os.Create(outputPath)
	if err != nil {
		// Fallback to current directory
		file, err = os.Create("sample_submission.csv")
		if err != nil {
			panic(err)
		}
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"id", "x", "y", "deg"})
	writer.WriteAll(treeData)

	fmt.Printf("Done! Output written to: %s\n", file.Name())
}

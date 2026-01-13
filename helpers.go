package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
)

func inExplored(needle Point, haystack []Point) bool {
	for _, x := range haystack {
		if x.Row == needle.Row && x.Col == needle.Col {
			return true
		}
	}

	return false
}

func EmptyContents(filePath string) error {
	directory := filePath
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		fullPath := filepath.Join(directory, entry.Name())
		err := os.RemoveAll(fullPath)
		if err != nil {
			fmt.Printf("failed to delete %s: %v\n", fullPath, err)
			continue
		}
	}

	return nil
}

func EnsureDir(path string) error {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("could not create directory %s: %w", path, err)
	}
	return nil
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}

	return x
}

func EuclideanDistance(p, goal Point) float64 {
	dx := float64(p.Row - goal.Row)
	dy := float64(p.Col - goal.Col)
	return math.Sqrt((dx * dx) + (dy * dy))
}

// might be better for larger mazes to be fair
// untested properly yet
func EuclideanDistanceSquared(p, goal Point) float64 {
	dx := float64(p.Row - goal.Row)
	dy := float64(p.Col - goal.Col)
	return (dx * dx) + (dy * dy)
}

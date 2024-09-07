package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// Function to run an external utility asynchronously with specified input and output files
func runExternalUtilityAsync(inputFile, outputFile string) error {
	dir := "../../utils/realesrgan-ncnn-vulkan-20220424-ubuntu/"
	cmd := exec.Command(dir+"realesrgan-ncnn-vulkan", "-i", dir+inputFile,
		"-o", dir+outputFile, "-j 1:1:1")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run external utility: %w", err)
	}

	return nil
}

func measureExecutionTime(numProcesses int, inputFile, outputFilePrefix string) time.Duration {
	var wg sync.WaitGroup
	wg.Add(numProcesses)

	start := time.Now()

	for i := 0; i < numProcesses; i++ {
		go func(index int) {
			defer wg.Done()
			outputFile := fmt.Sprintf("%s_%d.png", outputFilePrefix, index)
			if err := runExternalUtilityAsync(inputFile, outputFile); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}(i)
	}

	wg.Wait()
	return time.Since(start)
}

func main() {
	inputFile := "input2.jpg"
	outputFilePrefix := "output2"
	numProcessesList := []int{1, 2, 4, 8}

	for _, numProcesses := range numProcessesList {
		fmt.Printf("Running test with %d processes...\n", numProcesses)
		duration := measureExecutionTime(numProcesses, inputFile, outputFilePrefix)
		fmt.Printf("Execution time with %d processes: %v\n", numProcesses, duration)
		time.Sleep(1 * time.Second) // Sleep to avoid overlap and give time for process cleanup
	}
}

// Package fileprocessor provides methods for processing files using an external utility.
// It runs the external command synchronously and captures its output and errors for logging.
package fileprocessor

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
)

// FileProcessor provides a method for processing files using an external utility.
type FileProcessor struct {
	utilityDir string
	log        *slog.Logger
}

// loggerWriter is a helper type that implements the io.Writer interface
// and logs the data written to it using the provided logger.
type loggerWriter struct {
	log *slog.Logger
}

func (lw *loggerWriter) Write(p []byte) (n int, err error) {
	lw.log.Info(string(p))
	return len(p), nil
}

// NewFileProcessor creates a new instance of FileProcessor with the specified utility directory.
//
// Parameters:
//   - utilityDir: The directory where the external utility is located.
//   - log: A logger for capturing and outputting log messages.
//
// Returns:
//   - A pointer to a FileProcessor instance.
func NewFileProcessor(utilityDir string, log *slog.Logger) *FileProcessor {
	return &FileProcessor{
		utilityDir: utilityDir,
		log:        log,
	}
}

// Process runs an external utility synchronously and waits for its completion.
//
// Parameters:
//   - inputFile: The input file to be processed.
//   - outputFile: The output file where the result will be stored.
//
// Returns:
//   - An error if the external utility fails to run or encounters an error during execution.
func (fp *FileProcessor) Process(inputFile, outputFile string) error {
	inputPath := inputFile
	outputPath := outputFile

	path := filepath.Join(fp.utilityDir, "realesrgan-ncnn-vulkan")
	cmd := exec.Command(path, "-i", inputPath, "-o", outputPath, "-n", "realesrgan-x4plus")

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutMulti := io.MultiWriter(&stdoutBuf, &loggerWriter{log: fp.log})
	stderrMulti := io.MultiWriter(&stderrBuf, &loggerWriter{log: fp.log})

	cmd.Stdout = stdoutMulti
	cmd.Stderr = stderrMulti

	err := cmd.Run()

	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	fp.log.Info("Command output", "stdout", stdoutStr, "stderr", stderrStr)

	if err != nil {
		fp.log.Error("Failed to run external utility", "error", err, "stdout", stdoutStr, "stderr", stderrStr)
		return fmt.Errorf("failed to run external utility: %w", err)
	}

	return nil
}

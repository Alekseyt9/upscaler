package fileprocessor

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileProcessor_Process_Success(t *testing.T) {
	utilityDir := "./test-utility"
	inputFile := filepath.Join("testdata", "input.png")
	outputFile := filepath.Join("testdata", "output.png")

	logBuf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(logBuf, nil))

	fp := NewFileProcessor(utilityDir, logger)

	_, err := os.Stat(inputFile)
	require.NoError(t, err, "Input file does not exist")

	err = fp.Process(inputFile, outputFile)
	require.NoError(t, err, "Expected no error from Process")

	_, err = os.Stat(outputFile)
	require.NoError(t, err, "Expected output file to be created")

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "Command output", "Expected 'Command output' log in logger")

	err = os.Remove(outputFile)
	require.NoError(t, err, "Failed to clean up output file")
}

func TestFileProcessor_Process_Failure(t *testing.T) {
	utilityDir := "./test-utility"
	inputFile := filepath.Join("testdata", "nonexistent.png")
	outputFile := filepath.Join("testdata", "output.png")

	logBuf := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(logBuf, nil))

	fp := NewFileProcessor(utilityDir, logger)

	err := fp.Process(inputFile, outputFile)
	require.Error(t, err, "Expected an error from Process due to non-existent input file")

	logOutput := logBuf.String()
	assert.Contains(t, logOutput, "Failed to run external utility", "Expected error log for failed external utility")

	_, err = os.Stat(outputFile)
	assert.True(t, os.IsNotExist(err), "Output file should not exist")

	if _, err := os.Stat(outputFile); err == nil {
		os.Remove(outputFile)
	}
}

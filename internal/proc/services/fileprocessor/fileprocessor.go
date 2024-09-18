package fileprocessor

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
)

// FileProcessor предоставляет метод для обработки файлов с помощью внешней утилиты
type FileProcessor struct {
	utilityDir string
	log        *slog.Logger
}

type loggerWriter struct {
	log *slog.Logger
}

func (lw *loggerWriter) Write(p []byte) (n int, err error) {
	lw.log.Info(string(p))
	return len(p), nil
}

// NewFileProcessor создает новый экземпляр FileProcessor с указанием директории утилиты
func NewFileProcessor(utilityDir string, log *slog.Logger) *FileProcessor {
	return &FileProcessor{
		utilityDir: utilityDir,
		log:        log,
	}
}

// Process запускает внешнюю утилиту синхронно и ожидает завершения
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

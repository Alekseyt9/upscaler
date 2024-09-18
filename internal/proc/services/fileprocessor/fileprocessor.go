package fileprocessor

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
)

// FileProcessor предоставляет метод для обработки файлов с помощью внешней утилиты
type FileProcessor struct {
	utilityDir string
}

// NewFileProcessor создает новый экземпляр FileProcessor с указанием директории утилиты
func NewFileProcessor(utilityDir string) *FileProcessor {
	return &FileProcessor{
		utilityDir: utilityDir,
	}
}

// Process запускает внешнюю утилиту синхронно и ожидает завершения
func (fp *FileProcessor) Process(inputFile, outputFile string) error {
	inputPath := inputFile
	outputPath := outputFile

	path := filepath.Join(fp.utilityDir, "realesrgan-ncnn-vulkan")
	cmd := exec.Command(path, "-i", inputPath, "-o", outputPath)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdoutStr := stdoutBuf.String()
	stderrStr := stderrBuf.String()

	if err != nil {
		return fmt.Errorf("failed to run external utility: %w %s %s", err, stdoutStr, stderrStr)
	}

	return nil
}

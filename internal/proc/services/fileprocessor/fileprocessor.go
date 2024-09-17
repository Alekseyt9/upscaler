package fileprocessor

import (
	"fmt"
	"os"
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
	inputPath := filepath.Join(fp.utilityDir, inputFile)
	outputPath := filepath.Join(fp.utilityDir, outputFile)

	cmd := exec.Command(filepath.Join(fp.utilityDir, "realesrgan-ncnn-vulkan"), "-i", inputPath, "-o", outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run external utility: %w", err)
	}

	return nil
}

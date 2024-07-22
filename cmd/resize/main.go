package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nfnt/resize"
	_ "golang.org/x/image/webp"
)

func decodeImage(file *os.File) (image.Image, string, error) {
	ext := strings.ToLower(filepath.Ext(file.Name()))

	switch ext {
	case ".jpg", ".jpeg":
		img, err := jpeg.Decode(file)
		return img, "jpeg", err
	case ".png":
		img, err := png.Decode(file)
		return img, "png", err
	case ".webp":
		img, str, err := image.Decode(file)
		return img, str, err
	default:
		return nil, "", image.ErrFormat
	}
}

func resizeImage(inputPath string, outputPath string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Открываем файл изображения
	file, err := os.Open(inputPath)
	if err != nil {
		log.Printf("Error opening file %s: %v\n", inputPath, err)
		return
	}
	defer file.Close()

	// Декодируем изображение
	img, format, err := decodeImage(file)
	if err != nil {
		log.Printf("Error decoding file %s: %v\n", inputPath, err)
		return
	}
	log.Printf("Decoded format: %s\n", format)

	// Ресайзим изображение до 30x30
	m := resize.Resize(30, 30, img, resize.Lanczos3)

	// Сохраняем результат в буфер в формате PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, m)
	if err != nil {
		log.Printf("Error encoding file %s: %v\n", inputPath, err)
		return
	}

	// Сохраняем буфер в новый файл
	out, err := os.Create(outputPath)
	if err != nil {
		log.Printf("Error creating output file %s: %v\n", outputPath, err)
		return
	}
	defer out.Close()

	_, err = buf.WriteTo(out)
	if err != nil {
		log.Printf("Error writing to output file %s: %v\n", outputPath, err)
		return
	}

	log.Printf("Image resized and saved to %s\n", outputPath)
}

func main() {
	// Список входных и выходных файлов
	files := []struct {
		input  string
		output string
	}{
		{"pics/in1.webp", "pics/output1.jpg"},
		{"pics/in2.png", "pics/output2.jpg"},
		{"pics/in3.jpg", "pics/output3.jpg"},
	}

	var wg sync.WaitGroup

	// Запускаем ресайзинг в отдельных горутинах
	for _, file := range files {
		wg.Add(1)
		go resizeImage(file.input, file.output, &wg)
	}

	// Ожидаем завершения всех горутин
	wg.Wait()
	log.Println("All images resized.")
}

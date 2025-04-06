package utils

import (
	"image/jpeg"
	"log"
	"os"

	"github.com/disintegration/imaging"
)

// ProcessThumbnail открывает исходное изображение, изменяет его размер и сохраняет оптимизированную версию
func ProcessThumbnail(inputPath, outputPath string) error {
	// Открываем исходное изображение
	src, err := imaging.Open(inputPath)
	if err != nil {
		return err
	}

	// Изменяем размер до 320x320, сохраняя пропорции
	thumbnail := imaging.Fit(src, 320, 320, imaging.Lanczos)

	// Создаем файл для сохранения
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Сохраняем изображение в формате JPEG с качеством 90
	opts := jpeg.Options{Quality: 90}
	err = jpeg.Encode(outFile, thumbnail, &opts)
	if err != nil {
		return err
	}

	log.Printf("Обработанное изображение сохранено: %s", outputPath)
	return nil
}

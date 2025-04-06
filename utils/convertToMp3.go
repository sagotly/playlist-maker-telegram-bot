package utils

import (
	"log"
	"os/exec"
	"path/filepath"
)

// convertToMP3 конвертирует файл m4a в mp3 с использованием ffmpeg и возвращает путь к новому файлу
func ConvertToMP3(inputFile string) (string, error) {
	// Заменяем расширение на .mp3
	outputFile := inputFile[:len(inputFile)-len(filepath.Ext(inputFile))] + ".mp3"
	log.Printf("Начинаю конвертацию файла %s в %s", inputFile, outputFile)
	// ffmpeg -y -i inputFile -c:a libmp3lame -q:a 2 outputFile
	cmd := exec.Command("ffmpeg", "-y", "-i", inputFile, "-c:a", "libmp3lame", "-q:a", "2", outputFile)
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка выполнения ffmpeg для конвертации: %v", err)
		log.Printf("Вывод ffmpeg: %s", string(outBytes))
		return "", err
	}
	log.Printf("Конвертация завершена, выходной файл: %s", outputFile)
	return outputFile, nil
}

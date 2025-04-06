package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// mergeFiles создает временный список файлов для ffmpeg и склеивает аудиофайлы с перекодировкой
func MergeFiles(files []string, downloadDir string) (string, error) {
	listFile := filepath.Join(downloadDir, "files.txt")
	f, err := os.Create(listFile)
	if err != nil {
		log.Printf("Ошибка создания файла списка: %v", err)
		return "", err
	}
	defer f.Close()

	for _, file := range files {
		line := fmt.Sprintf("file '%s'\n", file)
		log.Printf("Добавляю строку в список: %s", line)
		_, err := f.WriteString(line)
		if err != nil {
			log.Printf("Ошибка записи в файл списка: %v", err)
			return "", err
		}
	}

	outputFile := filepath.Join(downloadDir, "output.mp3")
	log.Printf("Запускаю ffmpeg для склейки файлов. Выходной файл: %s", outputFile)
	// Флаг -y перезаписывает выходной файл без запроса.
	// Перекодировка всех файлов в mp3 для совместимости
	cmd := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", listFile, "-c:a", "libmp3lame", "-q:a", "2", outputFile)
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка выполнения ffmpeg: %v", err)
		log.Printf("Вывод ffmpeg: %s", string(outBytes))
		return "", err
	}
	log.Printf("ffmpeg завершился успешно. Выходной файл: %s", outputFile)
	return outputFile, nil
}

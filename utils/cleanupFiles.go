package utils

import (
	"log"
	"os"
)

// cleanupFiles удаляет временные файлы
func CleanupFiles(files []string) {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			log.Printf("Не удалось удалить файл %s: %v", file, err)
		} else {
			log.Printf("Файл %s удалён", file)
		}
	}
}

package utils

import (
	"io"
	"log"
	"net/http"
	"os"
)

// downloadFile скачивает файл по заданному URL и сохраняет его по пути localPath
func DownloadFile(url, localPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("HTTP GET ошибка: %v", err)
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(localPath)
	if err != nil {
		log.Printf("Ошибка создания файла %s: %v", localPath, err)
		return err
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("Ошибка записи в файл %s: %v", localPath, err)
		return err
	}
	log.Printf("Скачано %d байт в файл %s", written, localPath)
	return nil
}

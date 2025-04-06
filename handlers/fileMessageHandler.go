package handlers

import (
	"log"
	"os"
	"path/filepath"
	"playlist-maker-bot/entities"
	"playlist-maker-bot/utils"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
)

func ProcessFileMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update, chatID int64, downloadDir string) {
	var fileID, fileName, ext string

	if update.Message.Document != nil {
		fileID = update.Message.Document.FileID

		// Берем расширение из оригинального имени
		origName := update.Message.Document.FileName
		ext = strings.ToLower(filepath.Ext(origName))
		if ext == "" {
			ext = ".mp3" // значение по умолчанию
		}

		// Генерируем новое имя на основе UUID
		fileName = uuid.New().String() + ext

	} else if update.Message.Audio != nil {
		fileID = update.Message.Audio.FileID

		// Определяем расширение по MIME-типу
		ext = ".m4a"
		if update.Message.Audio.MimeType == "audio/mpeg" {
			ext = ".mp3"
		}

		fileName = uuid.New().String() + ext
	}

	if fileID == "" {
		return
	}

	if !strings.HasSuffix(strings.ToLower(fileName), ".mp3") && !strings.HasSuffix(strings.ToLower(fileName), ".m4a") {
		msg := tgbotapi.NewMessage(chatID, "Файл не является mp3 или m4a.")
		bot.Send(msg)
		return
	}

	log.Printf("Получен файл %s с fileID %s", fileName, fileID)

	fileConfig := tgbotapi.FileConfig{FileID: fileID}

	file, err := bot.GetFile(fileConfig)
	if err != nil {
		log.Printf("Ошибка получения информации о файле: %v", err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка получения файла: "+err.Error())
		bot.Send(msg)
		return
	}

	fileURL := file.Link(bot.Token)
	localPath := filepath.Join(downloadDir, fileName)
	log.Printf("Начинаю скачивание файла %s по URL %s", fileName, fileURL)

	err = utils.DownloadFile(fileURL, localPath)
	if err != nil {
		log.Printf("Ошибка скачивания файла %s: %v", fileName, err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка скачивания файла "+fileName+": "+err.Error())
		bot.Send(msg)
		return
	}

	log.Printf("Файл %s успешно скачан в %s", fileName, localPath)

	// Если файл m4a, конвертируем его в mp3
	if strings.HasSuffix(strings.ToLower(localPath), ".m4a") {
		convertedPath, err := utils.ConvertToMP3(localPath)
		if err != nil {
			log.Printf("Ошибка конвертации файла %s: %v", localPath, err)
			msg := tgbotapi.NewMessage(chatID, "Ошибка конвертации файла "+fileName+": "+err.Error())
			bot.Send(msg)
			return
		}

		log.Printf("Конвертация завершена, получен файл %s", convertedPath)

		// Удаляем оригинальный m4a файл
		err = os.Remove(localPath)
		if err != nil {
			log.Printf("Не удалось удалить исходный файл %s: %v", localPath, err)
		}

		localPath = convertedPath
	}

	entities.Files[chatID] = append(entities.Files[chatID], localPath)

	msg := tgbotapi.NewMessage(chatID, "Файл принят!")
	bot.Send(msg)
}

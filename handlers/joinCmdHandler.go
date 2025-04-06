package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"playlist-maker-bot/entities"
	"playlist-maker-bot/utils"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/google/uuid"
)

func HandleJoinCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, chatID int64, downloadDir string) {
	files, exists := entities.Files[chatID]
	if !exists || len(files) == 0 {
		msg := tgbotapi.NewMessage(chatID, "Нет файлов для склейки.")
		bot.Send(msg)
		return
	}
	log.Printf("Получены файлы для чата %d: %v", chatID, files)
	msg := tgbotapi.NewMessage(chatID, "Начинаю склеивать файлы...")
	bot.Send(msg)

	// Обработка прикреплённого изображения (обложки) для плейлиста
	var thumbPath string
	if update.Message.Photo != nil && len(*update.Message.Photo) > 0 {
		// Выбираем последний элемент массива фото (наилучшее качество)
		photo := (*update.Message.Photo)[len(*update.Message.Photo)-1]
		photoFileID := photo.FileID

		photoFile, err := bot.GetFile(tgbotapi.FileConfig{FileID: photoFileID})
		if err != nil {
			log.Printf("Ошибка получения информации о фото: %v", err)
		} else {
			photoURL := photoFile.Link(bot.Token)
			thumbPath = filepath.Join(downloadDir, "thumb_"+uuid.New().String()+".jpg")
			err = utils.DownloadFile(photoURL, thumbPath)
			if err != nil {
				log.Printf("Ошибка скачивания фото: %v", err)
				thumbPath = ""
			}
			err = utils.ProcessThumbnail(thumbPath, thumbPath)
			if err != nil {
				log.Printf("Ошибка обработки фото: %v", err)
				thumbPath = ""
			}
			log.Printf("Получена обложка для плейлиста: %s", thumbPath)
		}
	}

	// Используем аргумент команды как имя выходного файла.
	// Например: /join myPlaylist.mp3
	fileNameArg := strings.TrimSpace(update.Message.CommandArguments())
	if fileNameArg == "" && update.Message.Caption != "" {
		parts := strings.SplitN(update.Message.Caption, " ", 2)
		if len(parts) == 2 {
			fileNameArg = strings.TrimSpace(parts[1])
		}
	}

	if fileNameArg == "" {
		msg := tgbotapi.NewMessage(chatID, "Пожалуйста, укажите имя итогового файла после команды join, например: /join myPlaylist.mp3")
		bot.Send(msg)
		return
	}

	// Если имя файла не заканчивается на .mp3, добавляем расширение
	if !strings.HasSuffix(strings.ToLower(fileNameArg), ".mp3") {
		fileNameArg += ".mp3"
	}

	log.Printf("Начинаю склеивать файлы для чата %d: %v", chatID, files)

	outputFile, err := utils.MergeFiles(files, downloadDir)
	if err != nil {
		log.Printf("Ошибка при склейивании файлов для чата %d: %v", chatID, err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при склейивании файлов: "+err.Error())
		bot.Send(msg)
		return
	}

	// Переименовываем полученный файл в указанное имя
	finalOutputPath := filepath.Join(downloadDir, fileNameArg)
	err = os.Rename(outputFile, finalOutputPath)
	if err != nil {
		log.Printf("Ошибка при переименовании файла %s: %v", outputFile, err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при переименовании файла: "+err.Error())
		bot.Send(msg)
		return
	}

	log.Printf("Отправляю склеенный аудио файл %s в чат %d", finalOutputPath, chatID)
	err = SendAudioWithThumb(bot, chatID, finalOutputPath, thumbPath, "Вот ваш плейлист!")
	if err != nil {
		log.Printf("Ошибка при отправке файла %s в чат %d: %v", finalOutputPath, chatID, err)
		msg := tgbotapi.NewMessage(chatID, "Ошибка при отправке файла: "+err.Error())
		bot.Send(msg)
		return
	}
	// Чистим временные файлы
	allFiles := append(files, finalOutputPath)
	if thumbPath != "" {
		allFiles = append(allFiles, thumbPath)
	}
	log.Printf("Очищаю файлы для чата %d: %v", chatID, allFiles)
	utils.CleanupFiles(allFiles)
	entities.Files[chatID] = []string{}
}

// SendAudioWithThumb отправляет аудио с обложкой, используя кастомный HTTP-запрос.
func SendAudioWithThumb(bot *tgbotapi.BotAPI, chatID int64, audioPath, thumbPath, caption string) error {
	log.Printf("SendAudioWithThumb: Начало отправки аудио '%s', обложка '%s'", audioPath, thumbPath)

	// Формируем URL для метода sendAudio
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendAudio", bot.Token)
	log.Printf("SendAudioWithThumb: URL для запроса: %s", url)

	// Создаем multipart/form-data тело запроса
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Поле chat_id
	log.Printf("SendAudioWithThumb: Добавление поля chat_id: %d", chatID)
	if err := writer.WriteField("chat_id", strconv.FormatInt(chatID, 10)); err != nil {
		log.Printf("SendAudioWithThumb: Ошибка записи поля chat_id: %v", err)
		return err
	}

	// Поле caption, если задано
	if caption != "" {
		log.Printf("SendAudioWithThumb: Добавление поля caption: %s", caption)
		if err := writer.WriteField("caption", caption); err != nil {
			log.Printf("SendAudioWithThumb: Ошибка записи поля caption: %v", err)
			return err
		}
	}

	// Прикрепляем аудио файл
	log.Printf("SendAudioWithThumb: Открытие аудио файла: %s", audioPath)
	audioFile, err := os.Open(audioPath)
	if err != nil {
		errMsg := fmt.Errorf("ошибка открытия аудио файла: %v", err)
		log.Printf("SendAudioWithThumb: %v", errMsg)
		return errMsg
	}
	defer audioFile.Close()

	log.Printf("SendAudioWithThumb: Создание части для аудио файла")
	audioPart, err := writer.CreateFormFile("audio", filepath.Base(audioPath))
	if err != nil {
		errMsg := fmt.Errorf("ошибка создания части для аудио: %v", err)
		log.Printf("SendAudioWithThumb: %v", errMsg)
		return errMsg
	}

	_, err = io.Copy(audioPart, audioFile)
	if err != nil {
		errMsg := fmt.Errorf("ошибка копирования аудио файла: %v", err)
		log.Printf("SendAudioWithThumb: %v", errMsg)
		return errMsg
	}
	log.Printf("SendAudioWithThumb: Аудио файл успешно добавлен")

	// Прикрепляем обложку, если thumbPath не пустой
	if thumbPath != "" {
		log.Printf("SendAudioWithThumb: Открытие файла обложки: %s", thumbPath)
		thumbFile, err := os.Open(thumbPath)
		if err != nil {
			errMsg := fmt.Errorf("ошибка открытия файла обложки: %v", err)
			log.Printf("SendAudioWithThumb: %v", errMsg)
			return errMsg
		}
		defer thumbFile.Close()

		log.Printf("SendAudioWithThumb: Создание части для обложки")
		thumbPart, err := writer.CreateFormFile("thumb", filepath.Base(thumbPath))
		if err != nil {
			errMsg := fmt.Errorf("ошибка создания части для обложки: %v", err)
			log.Printf("SendAudioWithThumb: %v", errMsg)
			return errMsg
		}

		_, err = io.Copy(thumbPart, thumbFile)
		if err != nil {
			errMsg := fmt.Errorf("ошибка копирования файла обложки: %v", err)
			log.Printf("SendAudioWithThumb: %v", errMsg)
			return errMsg
		}
		log.Printf("SendAudioWithThumb: Файл обложки успешно добавлен")
	}

	// Завершаем формирование тела запроса
	log.Printf("SendAudioWithThumb: Закрытие multipart writer")
	if err := writer.Close(); err != nil {
		errMsg := fmt.Errorf("ошибка закрытия multipart writer: %v", err)
		log.Printf("SendAudioWithThumb: %v", errMsg)
		return errMsg
	}

	// Создаем HTTP-запрос
	log.Printf("SendAudioWithThumb: Создание HTTP-запроса")
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		errMsg := fmt.Errorf("ошибка создания HTTP-запроса: %v", err)
		log.Printf("SendAudioWithThumb: %v", errMsg)
		return errMsg
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос через стандартного HTTP-клиента бота
	log.Printf("SendAudioWithThumb: Отправка HTTP-запроса")
	resp, err := bot.Client.Do(req)
	if err != nil {
		errMsg := fmt.Errorf("ошибка отправки запроса: %v", err)
		log.Printf("SendAudioWithThumb: %v", errMsg)
		return errMsg
	}
	defer resp.Body.Close()
	log.Printf("SendAudioWithThumb: Получен ответ с кодом статуса %d", resp.StatusCode)

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMsg := fmt.Errorf("ошибка чтения ответа: %v", err)
		log.Printf("SendAudioWithThumb: %v", errMsg)
		return errMsg
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("ошибка Telegram API: %s", body)
		log.Printf("SendAudioWithThumb: %v", errMsg)
		return errMsg
	}

	log.Printf("SendAudioWithThumb: Аудио успешно отправлено")
	return nil
}

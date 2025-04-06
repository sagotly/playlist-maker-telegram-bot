package main

import (
	"log"
	"os"
	"path/filepath"
	"playlist-maker-bot/config"
	"playlist-maker-bot/handlers"
	"playlist-maker-bot/utils"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// downloadDir — директория для скачивания файлов (../downloads)
var downloadDir string

func init() {
	utils.ProcessThumbnail("downloads/input.jpg", "downloads/input.jpg")
	// Получаем рабочую директорию
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Не удалось получить рабочую директорию: %v", err)
	}

	// Родительская директория от cwd
	parentDir := filepath.Dir(cwd)
	downloadDir = filepath.Join(parentDir, "playlistMakerTelegramBot", "downloads")

	// Создаем директорию, если её нет
	err = os.MkdirAll(downloadDir, os.ModePerm)
	if err != nil {
		log.Fatalf("Не удалось создать директорию %s: %v", downloadDir, err)
	}

	log.Printf("Файлы будут сохраняться в директории: %s", downloadDir)
}

func main() {
	config, err := config.LoadConfig("./config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	// Чтение токена из переменной окружения
	botToken := config.TgBotToken
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Инициализация бота
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Настройка получения обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatalf("Ошибка получения обновлений: %v", err)
	}

	for update := range updates {
		if update.Message == nil { // игнорируем обновления без сообщения
			continue
		}

		chatID := update.Message.Chat.ID

		// Определяем текст команды из Message.Text или Message.Caption
		var commandText string
		if update.Message.Text != "" {
			commandText = update.Message.Text
		} else if update.Message.Caption != "" {
			commandText = update.Message.Caption
		}

		if strings.HasPrefix(commandText, "/") {
			processCommand(bot, &update, chatID)
		} else {
			handlers.ProcessFileMessage(bot, &update, chatID, downloadDir)
		}
	}

}

func processCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update, chatID int64) {
	var command string

	if update.Message.Text != "" {
		command = update.Message.Command()
	} else if update.Message.Caption != "" {
		// Если команда пришла вместе с фото или документом,
		// проверяем поле Caption на наличие команды.
		if strings.HasPrefix(update.Message.Caption, "/") {
			parts := strings.SplitN(update.Message.Caption, " ", 2)
			command = strings.TrimPrefix(parts[0], "/")
		}
	}

	switch command {
	case "start":
		handlers.HandleStartCommand(bot, chatID)
	case "join":
		handlers.HandleJoinCommand(bot, update, chatID, downloadDir)
	}
}

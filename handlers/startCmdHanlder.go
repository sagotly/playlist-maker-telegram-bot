package handlers

import (
	"playlist-maker-bot/entities"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func HandleStartCommand(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Отправьте mp3 или m4a файлы, чтобы добавить их в плейлист.\n\n"+"Используйте команду /join название плейлиста, чтобы склеить файлы в один плейлист, для обложки плейлиста отправтье /join вместе с вложеной фоткой.")
	bot.Send(msg)

	// Обнуляем список файлов для данного чата
	entities.Files[chatID] = []string{}
}

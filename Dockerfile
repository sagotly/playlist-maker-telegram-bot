# Используем образ Go версии 1.24.1 на Alpine
FROM golang:1.24.1-alpine

# Обновляем apk и устанавливаем ffmpeg
RUN apk update && apk add --no-cache ffmpeg

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app/playlist-maker

# Копируем исходный код сервиса в контейнер
COPY . .

# Запускаем Go-приложение
CMD ["go", "run", "main.go"]

# Этап сборки
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Копируем только go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальные исходники
COPY . .

# Сборка бинарника (без ненужных файлов)
RUN go build -o server ./cmd/main.go


# Минимальный образ для запуска
FROM alpine:3.19

# Переменная окружения с дефолтным путем
ENV CONFIG_PATH=/config/local.json

WORKDIR /app

# Копируем только бинарник
COPY --from=builder /app/server .

# Открываем порт (если сервер его использует)

# Возможность передавать флаг вручную:
# docker run goserver ./server --config=/config/local.json
ENTRYPOINT ["./server"]

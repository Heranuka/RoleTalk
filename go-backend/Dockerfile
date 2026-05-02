# Stage 1: Build
FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOMAXPROCS=1 go build -p 1 -o /app/main cmd/app/main.go

# Stage 2: Final
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/main .
# Копируем конфиги
COPY --from=builder /app/configs ./configs
# Копируем шаблоны
COPY --from=builder /app/internal/infra/mailer/templates ./internal/infra/mailer/templates

# ВАЖНО: Если ты хочешь, чтобы приложение читало .env как файл,
# его нужно скопировать (но лучше прокидывать через docker-compose)
# COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./main"]
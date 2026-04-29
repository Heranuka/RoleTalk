FROM golang:1.25.6 AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o voice-service ./cmd/app

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /app/voice-service .
EXPOSE 8080

CMD ["./voice-service"]
# Сборка
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/main.go

# Запуск
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

RUN mkdir config
COPY config/config.example.yaml ./config/config.example.yaml

COPY --from=builder /app/main .
CMD ["./main"]
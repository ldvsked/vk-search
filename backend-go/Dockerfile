FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Указываем путь к твоему main.go
RUN go build -o main cmd/app/main.go

FROM alpine:latest
WORKDIR /root/
# Копируем скомпилированный файл
COPY --from=builder /app/main .
# Мы копируем папку с промптами из билдера в финальный образ
COPY --from=builder /app/internal/infrastructure/llm/prompts ./internal/infrastructure/llm/prompts/
EXPOSE 8080
CMD ["./main"]
FROM golang:1.22.0-alpine

WORKDIR /app

COPY . .

RUN go mod download

# Запуск main.go напрямую
CMD ["go", "run", "./cmd/server/main.go", "-port", "${PORT}", "-storage", "${STORAGE}"]
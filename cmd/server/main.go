package main

import (
	"BASProject/config"
	"BASProject/internal/handlers"
	"BASProject/internal/services"
	"BASProject/internal/storage"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const maxConnections = 10

func main() {
	// Параметры командной строки для порта и пути к хранилищу
	port := flag.Int("port", 0, "Port for the server (overrides config)")
	storagePath := flag.String("storage", "", "Path to storage (overrides config)")
	flag.Parse()

	// Загрузка конфигурации
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Переопределение значений конфигурации, если заданы аргументы
	if *port != 0 {
		cfg.Server.Port = *port
	}
	if *storagePath != "" {
		cfg.Storage.Path = *storagePath
	}

	// Инициализация сервисов и обработчиков
	redisClient := storage.NewRedisClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	fileService := services.NewFileService(redisClient, cfg.Storage.Path)
	sessionService := services.NewSessionService(redisClient, fileService)
	startHandler := handlers.NewStartHandler(sessionService)
	uploadChunkHandler := handlers.NewUploadChunkHandler(sessionService)
	statusHandler := handlers.NewStatusHandler(sessionService)
	deleteHandler := handlers.NewDeleteHandler(sessionService)

	// Создаем канал для управления количеством подключений
	sem := make(chan struct{}, 10)

	// Функция-обертка для ограничения количества одновременных подключений
	limitConnections := func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sem <- struct{}{}        // Отправляем значение в канал (если канал заполнен, ожидать освобождения места)
			defer func() { <-sem }() // Освобождаем место в канале после завершения обработки
			handlerFunc(w, r)        // Вызываем оригинальный обработчик
		}
	}

	// Настройка маршрутов
	router := mux.NewRouter()
	router.HandleFunc("/upload/start", limitConnections(startHandler.StartSession)).Methods("POST")
	router.HandleFunc("/upload/{session_id}/chunk", limitConnections(uploadChunkHandler.UploadChunk)).Methods("POST")
	router.HandleFunc("/upload/complete/{session_id}", limitConnections(uploadChunkHandler.CompleteUpload)).Methods("POST")
	router.HandleFunc("/upload/status/{session_id}", limitConnections(statusHandler.GetUploadStatus)).Methods("GET")
	router.HandleFunc("/upload/{session_id}", limitConnections(deleteHandler.DeleteSession)).Methods("DELETE")

	// Запуск сервера
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server is running on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

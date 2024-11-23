package main

import (
	"BASProject/config"
	"BASProject/internal/handlers"
	"BASProject/internal/services"
	"BASProject/internal/storage"
	"BASProject/internal/utils"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Параметры командной строки для порта и пути к хранилищу
	port := flag.Int("port", 0, "Port for the server (overrides config)")
	storageFlag := flag.String("storage", "", "Path to storage (overrides config, default: 'data')")
	flag.Parse()

	// Загрузка конфигурации
	cfgPath := "config/config.yaml"
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// После обработки флагов командной строки в сервере
	if *port == 0 {
		// Если порт не указан, используем порт из конфига
		*port = cfg.Server.Port
	} else {
		cfg.Server.Port = *port
	}

	// После обработки флагов командной строки в сервере
	if *storageFlag == "" {
		// Если путь не указан, используем путь из конфига
		*storageFlag = cfg.Storage.Path
		log.Printf("No storage path provided, defaulting to 'data' directory.")
	}
	storagePath, err := utils.ExpandPath(*storageFlag)
	if err != nil {
		log.Fatalf("Error expanding storage path: %v", err)
	}

	// Проверяем, существует ли путь
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		log.Fatalf("Storage path does not exist: %s", *storageFlag)
	}

	cfg.Storage.Path = storagePath

	// Инициализация сервисов и обработчиков
	redisClient := storage.NewRedisClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	fileService := services.NewFileService(redisClient, cfg.Storage.Path)
	sessionService := services.NewSessionService(redisClient, fileService)
	startHandler := handlers.NewStartHandler(sessionService)
	uploadChunkHandler := handlers.NewUploadChunkHandler(sessionService)
	statusHandler := handlers.NewStatusHandler(sessionService)
	deleteHandler := handlers.NewDeleteHandler(sessionService)

	// Настройка маршрутов
	router := mux.NewRouter()
	router.HandleFunc("/upload/start", startHandler.StartSession).Methods("POST")
	router.HandleFunc("/upload/{session_id}/chunk", uploadChunkHandler.UploadChunk).Methods("POST")
	router.HandleFunc("/upload/complete/{session_id}", uploadChunkHandler.CompleteUpload).Methods("POST")
	router.HandleFunc("/upload/status/{session_id}", statusHandler.GetUploadStatus).Methods("GET")
	router.HandleFunc("/upload/{session_id}", deleteHandler.DeleteSession).Methods("DELETE")

	// Запуск сервера
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server is running on %s", addr)
	log.Printf("Storage path: %s", cfg.Storage.Path)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

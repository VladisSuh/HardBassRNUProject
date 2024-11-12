package services

import (
	"BASProject/internal/storage"
	"BASProject/internal/utils"
	"errors"
	"fmt"
	"strconv"
)

type SessionService struct {
	Storage     *storage.RedisClient
	FileService *FileService
}

func NewSessionService(storage *storage.RedisClient, fileService *FileService) *SessionService {
	return &SessionService{
		Storage:     storage,
		FileService: fileService,
	}
}

// Создание сессии загрузки файла с проверкой на существование файла
func (s *SessionService) CreateSession(fileName string, fileSize int64) (string, int64, error) {
	if fileName == "" || fileSize <= 0 {
		return "", 0, errors.New("invalid file name or file size")
	}

	// Проверка на существование сессии для этого файла
	// existingSessionID, err := s.Storage.GetSessionIDByFileName(fileName)
	// if err != nil {
	// 	// Сессия не найдена, создаем новую
	// } else {
	// 	// Возвращаем существующую сессию, если она есть
	// 	return existingSessionID, 0, nil
	// }

	sessionID := utils.GenerateSessionID()
	maxChunkSize := int64(1024 * 1024 * 1024) // 1GB
	chunkSize := s.FileService.CalculateChunkSize(fileSize, maxChunkSize)

	sessionData := map[string]interface{}{
		"file_name":     fileName,
		"file_size":     fileSize,
		"chunk_size":    chunkSize,
		"uploaded_size": 0,
		"status":        "in_progress", // Статус сессии
	}

	err := s.Storage.SaveSession(sessionID, sessionData)
	if err != nil {
		return "", 0, err
	}

	return sessionID, chunkSize, nil
}

func (s *SessionService) UpdateProgress(sessionID string, uploadedSize int64) error {
	// Получаем данные сессии из Redis
	sessionData, err := s.Storage.GetSessionData(sessionID)
	if err != nil {
		return fmt.Errorf("failed to retrieve session data: %w", err)
	}

	// Преобразуем "file_size" и "uploaded_size" к int64
	fileSize, err := toInt64(sessionData["file_size"])
	if err != nil {
		return fmt.Errorf("failed to parse file_size: %w", err)
	}

	// Обновляем "uploaded_size" с учетом приведения
	sessionData["uploaded_size"] = uploadedSize
	err = s.Storage.SaveSession(sessionID, sessionData)
	if err != nil {
		return fmt.Errorf("failed to update session data: %w", err)
	}

	// Проверяем, завершена ли загрузка
	if uploadedSize >= fileSize {
		sessionData["status"] = "completed"
		err = s.Storage.SaveSession(sessionID, sessionData)
		if err != nil {
			return fmt.Errorf("failed to mark session as completed: %w", err)
		}
	}

	return nil
}

// Функция приведения для безопасного преобразования interface{} в int64
func toInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case string:
		return strconv.ParseInt(v, 10, 64)
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("unsupported type for conversion to int64: %T", v)
	}
}

// Получение состояния загрузки для сессии
func (s *SessionService) GetUploadStatus(sessionID string) (map[string]interface{}, error) {
	// Получаем данные сессии из Redis
	sessionData, err := s.Storage.GetSessionData(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session data: %w", err)
	}

	fileSize, _ := sessionData["file_size"].(int64)
	uploadedSize, _ := sessionData["uploaded_size"].(int64)
	isComplete := uploadedSize >= fileSize

	status := map[string]interface{}{
		"file_size":      fileSize,
		"uploaded_size":  uploadedSize,
		"completed":      isComplete,
		"remaining_size": fileSize - uploadedSize,
		"status":         sessionData["status"].(string),
	}

	return status, nil
}

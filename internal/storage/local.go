package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	BasePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{BasePath: basePath}
}

func (s *LocalStorage) Save(filename string, data []byte) (string, error) {
	log.Printf("[LOCAL_STORAGE] start to save html raw")
	dirStorage := s.BasePath
	err := os.MkdirAll(dirStorage, 0755)
	if err != nil {
		log.Printf("[LOCAL_STORAGE] error when create directory: %v", err)
		return "", err
	}

	fullPath := filepath.Join(dirStorage, fmt.Sprintf("%s.html", filename))

	err = os.WriteFile(fullPath, data, 0644)
	if err != nil {
		log.Printf("[LOCAL_STORAGE] error when create file: %v", err)
		return "", err
	}

	log.Printf("[LOCAL_STORAGE] success create file: %s", fullPath)
	return fullPath, nil
}

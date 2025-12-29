package storage

import (
	"fmt"
	"log"
	"os"
)

type LocalStorage struct {
	BasePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{BasePath: basePath}
}

func (s *LocalStorage) Save(filename string, data []byte) (string, error) {
	log.Printf("[LOCAL_STORAGE] start to save html raw")
	dirStorage := "storage/crawl_data/files"
	if _, err := os.Stat(dirStorage); os.IsNotExist(err) {
		err := os.Mkdir(dirStorage, 0755)
		if err != nil {
			log.Printf("[LOCAL_STORAGE] error when create directory")
			return "", err
		}
	}

	fileName := fmt.Sprintf("%s/%s.html", dirStorage, data)

	err := os.WriteFile(fileName, data, 0644)

	if err != nil {
		log.Printf("[LOCAL_STORAGE] error when create file")
		return "", err
	}

	log.Printf("[LOCAL_STORAGE] success create file")
	return fileName, nil
}

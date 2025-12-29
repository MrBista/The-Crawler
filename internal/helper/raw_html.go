package helper

import (
	"fmt"
	"log"
	"os"
)

func saveRawHTml(id string, htmlContent []byte) (string, error) {
	dirName := "crawled_data"

	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.Mkdir(dirName, 0755)

		if err != nil {
			log.Printf("[RAW_HTML] failed to create crawled_data dir")
			return "", err
		}

	}

	filename := fmt.Sprintf("crawled_data/%s.html", id)

	err := os.WriteFile(filename, htmlContent, 0644)

	if err != nil {
		log.Printf("[RAW_HTML] failed to create file %s", filename)
		return "", err
	}
	log.Printf("[RAW_HTML] success to save file %s", filename)

	return filename, err
}

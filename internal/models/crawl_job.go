package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type CrawlJob struct {
	ID        string   `json:"id"`
	ParentId  string   `json:"parent_id"`
	Url       string   `json:"url"`
	Depth     int      `json:"depth"`
	Selectors []string `json:"selectors"`
}

type CrawlPage struct {
	ID         string  `gorm:"primaryKey;type:uuid"`
	ParentID   *string `gorm:"type:uuid;index"` // Pointer agar bisa null (root)
	URL        string  `gorm:"not null"`
	Title      string  `gorm:"type:text"`
	FilePath   string  `gorm:"type:text"`  // Lokasi file .html
	ParsedData JSONB   `gorm:"type:jsonb"` // Hasil ekstraksi selector
	Status     string  `gorm:"type:varchar(20)"`
	DepthLevel int     `gorm:"type:int"`
	CreatedAt  time.Time
}

type JSONB map[string]string

func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONB) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), j)
}

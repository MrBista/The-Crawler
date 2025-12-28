package dto

import "time"

type CrawlRequest struct {
	URl   string `json:"url"`
	Depth int    `json:"int"`
}

type CrawlJob struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Depth     int       `json:"depth"`
	CreatedAt time.Time `json:"created_at"`
}

package handler

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MrBista/The-Crawler/internal/models"
	"github.com/MrBista/The-Crawler/internal/queue"
	"github.com/MrBista/The-Crawler/internal/repository"
	"github.com/MrBista/The-Crawler/internal/storage"
	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
)

type CrawlHandler struct {
	repo       repository.CrawlRepository
	storage    storage.Storage
	producer   *queue.Producer
	kafkaTopic string
}

func NewCrawlHandler(repo repository.CrawlRepository, storage storage.Storage, producer *queue.Producer, topic string) *CrawlHandler {
	return &CrawlHandler{
		repo:       repo,
		storage:    storage,
		producer:   producer,
		kafkaTopic: topic,
	}
}

func (h *CrawlHandler) ProcessCrawl(job models.CrawlJob) {
	log.Printf("[Worker] starting to crawl for: %s", job.Url)

	if !strings.HasPrefix(job.Url, "http") {
		log.Printf("[Worker] Invalid url schema: %s", job.Url)
		return
	}

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, job.Url, nil)

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")
	req.Header.Set("Connection", "keep-alive")

	if err != nil {
		log.Printf("[Worker] Failed to create request %s", job.Url)
		return
	}
	res, err := client.Do(req)

	if err != nil {
		log.Printf("[Worker] Failed to get request %v", err)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("[Error] Non-200 Status Code: %d", res.StatusCode)
		return
	}

	var bodyBytes bytes.Buffer
	_, err = bodyBytes.ReadFrom(res.Body)
	if err != nil {
		log.Printf("[Error] Failed to read boy: %v", err)
		return
	}

	rawHtml := bodyBytes.Bytes()

	savePath, err := h.storage.Save(job.ID, rawHtml)
	if err != nil {
		log.Printf("[PROCESS_CRAWL_ERROR] failed to save file")
		return
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(rawHtml))

	if err != nil {
		log.Printf("[PROCESS_CRAWL_ERROR] failed to get doc")
		return
	}

	pageTitle := strings.TrimSpace(doc.Find("title").Text())
	extractedData := make(models.JSONB)

	if len(job.Selectors) > 0 {
		for _, selector := range job.Selectors {
			text := strings.TrimSpace(doc.Find(selector).Text())
			if text != "" {
				extractedData[selector] = text
			}
		}
	} else {
		desc, exists := doc.Find("meta[name='description']").Attr("content")
		if exists {
			extractedData["meta_description"] = desc
		}
	}

	var parentIdPtr *string

	if job.ParentId != "" {
		parentIdPtr = &job.ParentId
	}

	pageRecord := models.CrawlPage{
		ID:         job.ID,
		ParentID:   parentIdPtr,
		URL:        job.Url,
		Title:      pageTitle,
		FilePath:   savePath,
		ParsedData: extractedData,
		Status:     "completed",
		DepthLevel: job.Depth,
		CreatedAt:  time.Now(),
	}

	if err := h.repo.SavePage(&pageRecord); err != nil {
		log.Printf("[PROCESS_CRAWL_ERROR] failed to save page crawled")
	} else {
		log.Printf("[PROCESS_CRAWL] success to save page crawl")
	}
	log.Printf("[NEXT] depth value %s", job.Depth)
	if job.Depth > 0 {
		h.handleRecursiveLinks(doc, job)
	}
}

func (h *CrawlHandler) handleRecursiveLinks(doc *goquery.Document, parentJob models.CrawlJob) {
	log.Printf("[CRAWL_RECURSIVE] START TO RECURSIVE TASK LINK")
	visitedLinks := make(map[string]bool)

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		absoluteURl := resolveURL(parentJob.Url, href)
		if absoluteURl == "" || visitedLinks[absoluteURl] || !strings.HasPrefix(absoluteURl, "http") {
			log.Printf("[ERROR_RECURSIVE] failed to recursive links")
			return
		}
		visitedLinks[absoluteURl] = true
		uidChild := uuid.New().String()
		childJob := models.CrawlJob{
			ID:        uidChild,
			ParentId:  parentJob.ID,
			Url:       absoluteURl,
			Depth:     parentJob.Depth - 1,
			Selectors: parentJob.Selectors,
		}

		payload, _ := json.Marshal(childJob)

		partion, offset, err := h.producer.PublishMessage(h.kafkaTopic, uidChild, string(payload))

		if err != nil {
			log.Printf("[ERROR] failed to publish message")
		}

		log.Printf("[CHILD_RECURSIVE] success to send to partion (%s) and offset (%s)", partion, offset)
	})
}

func resolveURL(baseUrl, href string) string {
	base, err := url.Parse(baseUrl)
	if err != nil {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	return base.ResolveReference(ref).String()
}

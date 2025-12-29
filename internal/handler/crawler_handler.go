package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/MrBista/The-Crawler/internal/dto"
	"github.com/MrBista/The-Crawler/internal/helper"
	"github.com/MrBista/The-Crawler/internal/models"
	"github.com/MrBista/The-Crawler/internal/repository"
	"github.com/MrBista/The-Crawler/internal/storage"
)

type CrawlHandler struct {
	repo       repository.CrawlRepository
	storage    storage.Storage
	producer   sarama.SyncProducer
	kafkaTopic string
}

func NewCrawlHandler(repo repository.CrawlRepository, storage storage.Storage, producer sarama.SyncProducer, topic string) *CrawlHandler {
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
}

type ConsumerCrawlerHandler struct{}

func NewConsumerCrawlerHandler() (*ConsumerCrawlerHandler, error) {
	return &ConsumerCrawlerHandler{}, nil
}

func (c *ConsumerCrawlerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *ConsumerCrawlerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
func (c *ConsumerCrawlerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("Message claim: value = %s, timestamp = %v, topic = %s", msg.Value, msg.Timestamp, msg.Topic)

		var job dto.CrawlJob

		err := json.Unmarshal(msg.Value, &job)
		if err != nil {
			log.Printf("Error when parshing JSON : %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		helper.ProcessCrawl(job)
		session.MarkMessage(msg, "")
	}
	return nil
}

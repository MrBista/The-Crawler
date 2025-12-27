package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	log.Printf("Connecting producer to broker %v", brokers)
	producer, err := sarama.NewSyncProducer(brokers, config)

	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
	}, nil
}

func (p *Producer) SendMessage(topic, key, message string) (int32, int64, error) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
		Key:   sarama.StringEncoder(key),
	}

	partion, offset, err := p.producer.SendMessage(msg)

	if err != nil {
		return 0, 0, err
	}

	return partion, offset, nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

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

func main() {

	fmt.Println("Hello Crawler")

	brokers := []string{"localhost:9092"}
	topic := "crawler-get"

	producer, err := NewProducer(brokers)

	if err != nil {
		log.Fatalf("Faild to create producer :%v", err)
	}

	defer producer.Close()

	app := fiber.New()

	app.Get("/api/v1/crawl", func(c *fiber.Ctx) error {

		var req CrawlRequest

		dataFailedBody := make(map[string]interface{})

		dataFailedBody["data"] = nil
		dataFailedBody["message"] = "Something went wrong"

		if err := c.BodyParser(req); err != nil {
			dataFailedBody["message"] = "Failed to parse body"
			return c.Status(fiber.StatusBadRequest).JSON(dataFailedBody)
		}

		if req.URl == "" {
			dataFailedBody["message"] = "Url is required"
			return c.Status(fiber.StatusBadRequest).JSON(dataFailedBody)
		}

		jobId := uuid.New().String()
		job := CrawlJob{
			ID:        jobId,
			URL:       req.URl,
			Depth:     req.Depth,
			CreatedAt: time.Now(),
		}

		jobBytes, err := json.Marshal(job)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{
					"data":    nil,
					"message": "Failed to parse marshl",
				},
			)
		}

		partision, offset, err := producer.SendMessage(topic, jobId, string(jobBytes))

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{
					"data":    nil,
					"message": "Failed to send message to broker",
				},
			)
		}
		log.Printf("message stored in topic (%s)/partision(%d)/offset(%d)", topic, partision, offset)

		return c.Status(fiber.StatusOK).JSON(
			fiber.Map{
				"data": fiber.Map{
					"message": "crawl job in queued",
					"job_id":  jobId,
					"status":  "pending",
				},
			},
		)
	})

	log.Printf("Successfully listen to port 3000")
	app.Listen(":3000")
}

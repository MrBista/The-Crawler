package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/MrBista/The-Crawler/internal/dto"
	"github.com/MrBista/The-Crawler/internal/queue"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func main() {
	fmt.Println("Hello World")

	app := fiber.New()

	brokers := []string{"localhost:9092"}
	topic := "crawler-get"

	producer, err := queue.NewProducer(brokers)

	if err != nil {
		log.Panicf("Error to start producer brokers %v with message %v", brokers, err)
	}

	defer producer.Close()

	app.Post("/api/v1/crawl", func(c *fiber.Ctx) error {
		var reqBody dto.CrawlRequest

		if err := c.BodyParser(&reqBody); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"data":    nil,
				"message": "failed to parse body",
			})
		}

		if reqBody.URl == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"data":    nil,
				"message": "url body is required",
			})
		}

		jobId := uuid.New().String()
		job := dto.CrawlJob{
			ID:        jobId,
			URL:       reqBody.URl,
			Depth:     reqBody.Depth,
			CreatedAt: time.Now(),
		}

		jobMarshal, err := json.Marshal(job)

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"data":    nil,
				"message": "url body is required",
			})
		}
		partion, offset, err := producer.PublishMessage(topic, jobId, string(jobMarshal))

		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"data":    nil,
				"message": "url body is required",
			})
		}

		log.Printf("Message stored in topic (%s) with partion = %v and offset = %v", topic, partion, offset)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"data": fiber.Map{
				"job_id":  jobId,
				"status":  "pending",
				"Message": "message stored in brokers",
			},
		})
	})

	log.Printf("Successfully listen to port 3000")

	err = app.Listen(":3001")
	if err != nil {
		log.Printf("Error when startup application %v", err)
	}
}

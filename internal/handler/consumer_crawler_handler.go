package handler

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/MrBista/The-Crawler/internal/models"
)

type ConsumerCrawlerHandler struct {
	crawlHandler *CrawlHandler
}

func NewConsumerCrawlerHandler(crawlHandler *CrawlHandler) (*ConsumerCrawlerHandler, error) {
	return &ConsumerCrawlerHandler{
		crawlHandler: crawlHandler,
	}, nil
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

		var job models.CrawlJob

		err := json.Unmarshal(msg.Value, &job)
		if err != nil {
			log.Printf("Error when parshing JSON : %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		c.crawlHandler.ProcessCrawl(job)
		session.MarkMessage(msg, "")
	}
	return nil
}

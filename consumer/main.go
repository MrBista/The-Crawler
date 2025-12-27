package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/PuerkitoBio/goquery"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

type CrawlJob struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Depth     int       `json:"depth"`
	CreatedAt time.Time `json:"created_at"`
}

type Consumer struct {
	consumer         sarama.Consumer
	partionConsumers map[int32]sarama.PartitionConsumer
}

func NewConsumer(brokers []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // dibaca dari yg terlama masuk broker
	config.Version = sarama.V2_8_0_0

	log.Printf("Connecting consumer to kafka brokers %v", brokers)
	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{consumer: consumer, partionConsumers: make(map[int32]sarama.PartitionConsumer)}, nil
}

func (c *Consumer) Consume(topic string, partion int32) (<-chan *sarama.ConsumerMessage, <-chan *sarama.ConsumerError) {

	log.Printf("Consuming partion %d of topic %s", partion, topic)

	pc, err := c.consumer.ConsumePartition(topic, partion, sarama.OffsetOldest)

	if err != nil {
		log.Printf("Failed to start consume for partion %d of topic %s", partion, topic)
		return nil, nil
	}
	c.partionConsumers[partion] = pc

	return pc.Messages(), pc.Errors()

}

func (c *Consumer) ClosePartion(partion int32) error {
	if pc, ok := c.partionConsumers[partion]; ok {
		err := pc.Close()
		delete(c.partionConsumers, partion)
		return err
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

type ConsumerGroup struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
}

func NewConsumerGroup(brokers []string, groupId string, topic string) (*ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.V2_8_0_0
	group, err := sarama.NewConsumerGroup(brokers, groupId, config)

	if err != nil {
		return nil, err

	}

	return &ConsumerGroup{
		consumerGroup: group,
		topic:         topic,
	}, nil
}

type ConsumerHandler struct{}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Printf("Message claim: value = %v, timestamp = %v, topic= %s", string(msg.Value), msg.Timestamp, msg.Topic)

		var job CrawlJob
		err := json.Unmarshal(msg.Value, &job)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			session.MarkMessage(msg, "")
			continue
		}
		processCrawl(job)
		session.MarkMessage(msg, "")
	}

	return nil
}

func processCrawl(job CrawlJob) {
	log.Printf("[Worker] starting to crawl for: %s", job.URL)

	if !strings.HasPrefix(job.URL, "http") {
		log.Printf("[Worker] Invalid url schema: %s", job.URL)
		return
	}

	resp, err := http.Get(job.URL)
	if err != nil {
		log.Printf("[Worker] Failed to fetch url %v", job.URL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[Worker] Non-200 status code: %d", resp.StatusCode)
		return
	}

	// Parsing HTML dengan GoQuery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("[Worker] Failed to parse HTML: %v", err)
		return
	}

	// Ekstrak Title Halaman
	title := doc.Find("title").Text()

	// Simulasi hasil scraping
	fmt.Println("---------------------------------------------------")
	fmt.Printf("✅ SUCCESS CRAWL\n")
	fmt.Printf("ID    : %s\n", job.ID)
	fmt.Printf("URL   : %s\n", job.URL)
	fmt.Printf("TITLE : %s\n", strings.TrimSpace(title))
	fmt.Println("---------------------------------------------------")

	html, err := doc.Html()
	if err != nil {
		log.Printf("[Worker] Failed to render HTML: %v", err)
		return
	}

	// log.Printf("[Worker] HTML RESULT:\n%s", html)
	generatePDF(html, job.ID)
}

func generatePDF(html, nameFile string) error {
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return err
	}

	page := wkhtmltopdf.NewPageReader(strings.NewReader(html))
	page.EnableLocalFileAccess.Set(true)

	pdfg.AddPage(page)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Dpi.Set(300)

	err = pdfg.Create()
	if err != nil {
		return err
	}

	return pdfg.WriteFile(nameFile + ".pdf")
}

func main() {
	fmt.Println("Hello Crawler")

	brokers := []string{"localhost:9092"}
	topic := "crawler-get"

	group, err := NewConsumerGroup(brokers, "crawler-worker-group", topic)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}
	defer func() {
		_ = group.consumerGroup.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			handler := ConsumerHandler{}

			if err := group.consumerGroup.Consume(ctx, []string{topic}, &handler); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	log.Println("Worker is running... Waiting for jobs.")

	<-sigterm
	log.Println("Terminating: context cancelled")
	cancel()
	wg.Wait()
}

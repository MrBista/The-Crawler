package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MrBista/The-Crawler/conf"
	"github.com/MrBista/The-Crawler/internal/handler"
	"github.com/MrBista/The-Crawler/internal/queue"
	"github.com/MrBista/The-Crawler/internal/repository"
	"github.com/MrBista/The-Crawler/internal/storage"
)

func main() {
	fmt.Println("Hello Crawler")

	envConv, err := conf.LoadConfig()

	if err != nil {
		log.Panicf("failed to load config %v", err)
	}

	dbConnect, _ := conf.Connect(envConv.DBConfig)

	crawlRepository := repository.NewCrawlRepositoryImpl(dbConnect)

	fileStore := storage.NewLocalStorage("./storage/crawl_data/files")

	brokers := []string{"localhost:9092"}
	topic := "crawler-get"
	groupConsumerId := "crawler-worker-group"
	producer, err := queue.NewProducer(brokers)

	if err != nil {
		log.Panicf("Error to start producer brokers %v with message %v", brokers, err)
	}

	defer producer.Close()

	crawlHandler := handler.NewCrawlHandler(crawlRepository, fileStore, producer, topic)

	consumerCrawlHandler, err := handler.NewConsumerCrawlerHandler(crawlHandler)

	group, err := queue.NewConsumerGroup(brokers, groupConsumerId, topic)
	if err != nil {
		log.Panicf("Error creating consumer group client: %v", err)
	}
	defer func() {
		_ = group.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {

			if err := group.ConsumerGroup.Consume(ctx, []string{topic}, consumerCrawlHandler); err != nil {
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

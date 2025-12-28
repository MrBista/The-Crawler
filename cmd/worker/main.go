package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
	"github.com/MrBista/The-Crawler/internal/queue"
)

type ConsumerHandler struct{}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	return nil
}

func main() {
	fmt.Println("Hello Crawler")

	brokers := []string{"localhost:9092"}
	topic := "crawler-get"

	group, err := queue.NewConsumerGroup(brokers, "crawler-worker-group", topic)
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
			handler := ConsumerHandler{}

			if err := group.ConsumerGroup.Consume(ctx, []string{topic}, &handler); err != nil {
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

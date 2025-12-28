package queue

import (
	"log"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	log.Printf("[PRODUCER] starting to setup producer kafka")
	config := sarama.NewConfig()

	config.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(brokers, config)

	if err != nil {
		log.Printf("[PRODUCER] failed to connect producer to broker %v", brokers)
		return nil, err
	}

	log.Printf("[PRODUCER] sucessfully connect to producer kafka")

	return &Producer{
		producer: producer,
	}, nil
}

func (k *Producer) PublishMessage(topic, key, message string) (int32, int64, error) {

	msg := &sarama.ProducerMessage{
		Value: sarama.ByteEncoder(message),
		Key:   sarama.StringEncoder(key),
		Topic: topic,
	}

	partion, offset, err := k.producer.SendMessage(msg)

	// partion,offset err := k.producer.Sendmes(msg)

	if err != nil {
		return 0, 0, err
	}

	return partion, offset, nil
}

func (k *Producer) Close() error {
	return k.producer.Close()
}

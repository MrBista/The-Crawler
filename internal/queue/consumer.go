package queue

import (
	"log"

	"github.com/IBM/sarama"
)

type ConsumerGroup struct {
	ConsumerGroup sarama.ConsumerGroup
	topic         string
}

func NewConsumerGroup(broker []string, groupId string, topic string) (*ConsumerGroup, error) {
	log.Printf("[CONSUMER] Start new consumer group")
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.V2_8_0_0

	group, err := sarama.NewConsumerGroup(broker, groupId, config)
	if err != nil {
		return nil, err
	}

	return &ConsumerGroup{
		ConsumerGroup: group,
		topic:         topic,
	}, nil
}

func (c *ConsumerGroup) Close() error {
	return c.ConsumerGroup.Close()
}

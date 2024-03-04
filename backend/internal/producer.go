package internal

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

type Producer struct {
	config   KafkaConfig
	producer sarama.SyncProducer
}

func NewProducer(config KafkaConfig) (*Producer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return &Producer{
		config:   config,
		producer: producer,
	}, nil
}

func (producer *Producer) SendTaskEvent(task *CaptionTaskEvent) error {
	jsonTask, err := json.Marshal(task)
	if err != nil {
		return err
	}

	_, _, err = producer.producer.SendMessage(&sarama.ProducerMessage{
		Topic: producer.config.TaskTopic,
		Value: sarama.StringEncoder(jsonTask),
	})
	return err
}

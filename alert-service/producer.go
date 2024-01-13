package main

import (
	"github.com/IBM/sarama"
)

type Producer interface {
	Send(id string, price string) error
}

type kafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaProducer(addr []string, topic string) (Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Retry.Max = 5
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(addr, config)
	if err != nil {
		return nil, err
	}

	return &kafkaProducer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (k *kafkaProducer) Send(id string, price string) error {
	msg := &sarama.ProducerMessage{
		Topic: k.topic,
		Key:   sarama.StringEncoder(id),
		Value: sarama.StringEncoder(price),
	}

	partition, offset, err := k.producer.SendMessage(msg)
	logger.Info().
		Int32("partition", partition).
		Int64("offset", offset).
		Str("id", id).
		Str("price", price).
		Send()

	return err
}

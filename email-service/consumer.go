package main

import (
	"context"
	"log"
	"strconv"

	database "email-service/database/sqlc"

	"github.com/IBM/sarama"
)

type state string

const (
	Created   state = "created"
	Triggered state = "triggered"
	Deleted   state = "deleted"
	Completed state = "completed"
)

// kafka consumer code here
type Consumer interface {
	Process(context.Context) error
}

type kafkaConsumer struct {
	db     database.Querier
	email  Emailer
	cg     sarama.ConsumerGroup
	topics []string
}

func NewKafkaConsumer(db database.Querier, email Emailer, addr []string, group string, topics []string) (Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(addr, group, config)
	if err != nil {
		return nil, err
	}

	return &kafkaConsumer{
		db:     db,
		email:  email,
		cg:     consumerGroup,
		topics: topics,
	}, nil
}

func (*kafkaConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (*kafkaConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (k *kafkaConsumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		alertID := string(msg.Key)
		price := string(msg.Value)

		alertIDInt64, err := strconv.ParseInt(alertID, 10, 64)
		if err != nil {
			log.Println("Error parsing alertID:", err)
			continue
		}

		email, err := k.db.GetUserEmailByAlertID(sess.Context(), alertIDInt64)
		if err != nil {
			log.Println("Error getting user email:", err)
			continue
		}

		k.email.send(
			"Crypto Alert",
			"Your alert has been triggered! The price is now "+price+".",
			[]string{email},
			nil,
			nil,
			nil,
		)

		// Mark message as processed
		params := database.UpdateAlertStatusParams{
			ID:     alertIDInt64,
			Status: string(Completed),
		}
		err = k.db.UpdateAlertStatus(sess.Context(), params)
		if err != nil {
			log.Println("Error updating alert status:", err)
			continue
		}
		
		sess.MarkMessage(msg, "")
	}

	return nil
}

func (k *kafkaConsumer) Process(ctx context.Context) error {
	for {
		err := k.cg.Consume(ctx, k.topics, k)
		if err != nil {
			log.Println("Error from consumer: ", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

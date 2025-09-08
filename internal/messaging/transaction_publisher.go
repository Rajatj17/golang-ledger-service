package messaging

import (
	"encoding/json"
	"fmt"
	"golang-exercise/config"
	dto "golang-exercise/internal/dto"
	"time"

	"github.com/streadway/amqp"
)

type TransactionPublisher struct {
	rabbitmq *RabbitMQ
}

const EXCHANGE_NAME = ""

func NewTransactionPublisher(rabbitmq *RabbitMQ) *TransactionPublisher {
	return &TransactionPublisher{
		rabbitmq: rabbitmq,
	}
}

func (TrxnPublisher *TransactionPublisher) PublishTransaction(txMsg *dto.TransactionMessage) error {
	body, err := json.Marshal(txMsg)
	if err != nil {
		return fmt.Errorf("failed to serialize the txMsg")
	}

	queueName := config.GetConfig().RabbitMQ.Queue
	fmt.Printf("Publishing to queue: %s\n", queueName)

	// Publish to the Queue
	err = TrxnPublisher.rabbitmq.channel.Publish(
		EXCHANGE_NAME,
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish transaction: %w", err)
	}

	fmt.Printf("Successfully published transaction to queue: %s\n", queueName)
	return nil
}

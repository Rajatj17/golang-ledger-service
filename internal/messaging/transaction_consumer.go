package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-exercise/config"
	dto "golang-exercise/internal/dto"
	"golang-exercise/internal/service"
	"log"

	"github.com/streadway/amqp"
)

type TransactionConsumer struct {
	rabbitmq       *RabbitMQ
	txService      *service.TransactionService
	accountService *service.AccountService
}

func NewTransactionConsumer(
	rabbitmq *RabbitMQ,
	accountService *service.AccountService,
	txService *service.TransactionService,
) *TransactionConsumer {

	return &TransactionConsumer{
		rabbitmq:       rabbitmq,
		txService:      txService,
		accountService: accountService,
	}
}

func (trxnConsumer *TransactionConsumer) StartConsuming() error {
	// Start consuming messages from the queue
	messages, err := trxnConsumer.rabbitmq.channel.Consume(
		config.GetConfig().RabbitMQ.Queue, // queue name
		"transaction-processor",           // consumer tag
		false,                             // auto-ack (we'll manually acknowledge)
		false,                             // exclusive
		false,                             // no-local
		false,                             // no-wait
		nil,                               // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("Started consuming transaction messages...")

	// Process messages in a goroutine
	go trxnConsumer.processMessages(messages)

	return nil
}

func (trxnConsumer *TransactionConsumer) processMessages(messages <-chan amqp.Delivery) {
	for delivery := range messages {
		log.Printf("Received transaction message: %s", delivery.MessageId)

		var txMsg dto.TransactionMessage

		if err := json.Unmarshal(delivery.Body, &txMsg); err != nil {
			log.Printf("failed to unmarshal transaction: %v", err)
			delivery.Nack(false, false)
			continue
		}

		err := trxnConsumer.processTransaction(&txMsg)
		if err != nil {
			log.Printf("Failed to process transaction %s: %v", txMsg.ID, err)
			delivery.Nack(false, true) // Reject and requeue for retry
		}

		log.Printf("Successfully processed transaction %s", txMsg.ID)
		delivery.Ack(false) // Acknowledge successful processing
	}
}

func (trxnConsumer *TransactionConsumer) processTransaction(txMsg *dto.TransactionMessage) error {
	// Process the transaction using the refactored method
	return trxnConsumer.txService.ProcessTransaction(
		context.Background(),
		txMsg.ID,
		txMsg.AccountNumber,
		txMsg.Amount,
		txMsg.Type,
	)
}

package messaging

import (
	"fmt"
	"golang-exercise/config"
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	connected  bool
}

func NewRabbitMQ() *RabbitMQ {
	return &RabbitMQ{}
}

func (r *RabbitMQ) createConnectionString(rabbitMqConfig config.RabbitMQ) string {
	// General format is amqp://<user>:<password>@<host>:<port>/<vhost>
	connectionString := fmt.Sprintf("amqp://%s:%s@%s:%d/",
		rabbitMqConfig.UserName,
		rabbitMqConfig.Password,
		rabbitMqConfig.Host,
		rabbitMqConfig.Port,
	)

	log.Println(connectionString, "SSS", rabbitMqConfig.UserName)

	return connectionString
}

func (r *RabbitMQ) Connect() error {
	dns := r.createConnectionString(config.GetConfig().RabbitMQ)

	// Connect to RabbitMQ
	connection, err := amqp.Dial(dns)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.connection = connection

	r.channel, err = r.connection.Channel()
	if err != nil {
		return fmt.Errorf("faile to opn up channel: %w", err)
	}

	r.connected = true
	log.Println("Connected to RabbitMQ successfully!")

	return nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}

	if r.connection != nil {
		return r.connection.Close()
	}

	return nil
}

func (r *RabbitMQ) IsConnected() bool {
	return r.connected
}

func (r *RabbitMQ) DeclareQueue(queueName string) error {
	// Setting values for following options
	// durable, autoDelete, exclusive, noWait bool, args
	_, err := r.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange queue: %w", err)
	}

	return nil
}

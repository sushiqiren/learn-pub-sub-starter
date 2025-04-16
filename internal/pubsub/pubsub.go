package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// SimpleQueueType enum values
const (
	QueueDurable = iota
	QueueTransient
)

// PublishJSON publishes a message as JSON to a RabbitMQ exchange with the specified routing key.
func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	// Marshal the value to JSON
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create a context
	ctx := context.Background()

	// Publish the message
	err = ch.PublishWithContext(
		ctx,
		exchange, // exchange
		key,      // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonBytes,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// DeclareAndBind creates a channel, declares a queue, and binds it to an exchange
func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
) (*amqp.Channel, amqp.Queue, error) {
	// Create a new channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Set queue parameters based on simpleQueueType
	durable := simpleQueueType == QueueDurable
	autoDelete := simpleQueueType == QueueTransient
	exclusive := simpleQueueType == QueueTransient

	// Declare a queue
	q, err := ch.QueueDeclare(
		queueName,  // name
		durable,    // durable
		autoDelete, // auto-delete
		exclusive,  // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		ch.Close() // Close channel on error
		return nil, amqp.Queue{}, fmt.Errorf("failed to declare a queue: %w", err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(
		q.Name,   // queue name
		key,      // routing key
		exchange, // exchange
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		ch.Close() // Close channel on error
		return nil, amqp.Queue{}, fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	return ch, q, nil
}

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

// SubscribeJSON subscribes to messages from a queue, unmarshals them into the specified type,
// and processes them with the provided handler function.
func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T),
) error {
	// Call DeclareAndBind to ensure the queue exists and is bound to the exchange
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	// Get a channel of delivery messages
	msgs, err := ch.Consume(
		q.Name, // queue name
		"",     // consumer name (auto-generated)
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		ch.Close() // Close channel on error
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	// Start a goroutine to process messages
	go func() {
		defer ch.Close() // Ensure channel is closed when goroutine exits

		for d := range msgs {
			// Unmarshal the message body into type T
			var msg T
			err := json.Unmarshal(d.Body, &msg)
			if err != nil {
				fmt.Printf("Error unmarshaling message: %v\n", err)
				d.Ack(false) // Acknowledge even failed messages to remove from queue
				continue
			}

			// Call the handler with the unmarshaled message
			handler(msg)

			// Acknowledge the message
			d.Ack(false)
		}
	}()

	return nil
}

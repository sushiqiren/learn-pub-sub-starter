package pubsub

import (
    "context"
    "encoding/json"
    "fmt"

    amqp "github.com/rabbitmq/amqp091-go"
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


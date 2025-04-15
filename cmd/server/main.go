package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sushiqiren/learn-pub-sub-starter/internal/pubsub"
	"github.com/sushiqiren/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril server...")

	// Declare the connection string
	connStr := "amqp://guest:guest@localhost:5672/"

	// Connect to RabbitMQ
	conn, err := amqp.Dial(connStr)
	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("Successfully connected to RabbitMQ")

	// Create a new channel
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Failed to open a channel: %v\n", err)
		return
	}
	defer ch.Close()
	fmt.Println("Successfully opened a channel")

	// Create a PlayingState message with paused set to true
	pauseMessage := routing.PlayingState{
		IsPaused: true,
	}

	// Publish the message to the exchange
	err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, pauseMessage)
	if err != nil {
		fmt.Printf("Failed to publish message: %v\n", err)
		return
	}
	fmt.Println("Successfully published pause message to exchange")

	// Wait for a signal to exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Waiting for shutdown signal...")
	<-signalChan

	fmt.Println("Shutdown signal received. Closing connection...")
}

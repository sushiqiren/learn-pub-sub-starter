package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sushiqiren/learn-pub-sub-starter/internal/gamelogic"
	"github.com/sushiqiren/learn-pub-sub-starter/internal/pubsub"
	"github.com/sushiqiren/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")

	// Get the username using the ClientWelcome function
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

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

	// Create a queue name with format: pause.username
	queueName := routing.PauseKey + "." + username

	// Declare and bind a queue to the exchange
	ch, q, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.QueueTransient,
	)
	if err != nil {
		fmt.Printf("Failed to declare and bind queue: %v\n", err)
		return
	}
	defer ch.Close()
	fmt.Printf("Successfully declared queue '%s' and bound to exchange for user: %s\n", q.Name, username)

	// Wait for a signal to exit
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Waiting for shutdown signal...")
	<-signalChan

	fmt.Println("Shutdown signal received. Closing connection...")
}

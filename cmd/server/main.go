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
	fmt.Println("Starting Peril server...")

	// Display the available commands
	gamelogic.PrintServerHelp()

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

	// Declare and bind a durable queue for game logs
	_, gameLogsQueue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug,
		pubsub.QueueDurable,
	)
	if err != nil {
		fmt.Printf("Failed to declare and bind queue for game logs: %v\n", err)
		return
	}
	fmt.Printf("Successfully declared durable queue '%s' for game logs\n", gameLogsQueue.Name)

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

	// Set up signal handling in a separate goroutine
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		fmt.Println("\nShutdown signal received. Closing connection...")
		os.Exit(0)
	}()

	// Start the input loop
	fmt.Println("Server is ready. Enter commands:")
	for {
		// Get input from the user
		words := gamelogic.GetInput()

		// If no input, continue the loop
		if len(words) == 0 {
			continue
		}

		// Process the command
		command := words[0]

		switch command {
		case "help":
			gamelogic.PrintServerHelp()
		case "quit":
			gamelogic.PrintQuit()
			fmt.Println("Shutting down server...")
			return
		case "pause":
			fmt.Println("Pausing game...")
			pauseMessage := routing.PlayingState{
				IsPaused: true,
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, pauseMessage)
			if err != nil {
				fmt.Printf("Failed to publish pause message: %v\n", err)
			} else {
				fmt.Println("Game paused successfully")
			}
		case "resume":
			fmt.Println("Resuming game...")
			resumeMessage := routing.PlayingState{
				IsPaused: false,
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, resumeMessage)
			if err != nil {
				fmt.Printf("Failed to publish resume message: %v\n", err)
			} else {
				fmt.Println("Game resumed successfully")
			}
		default:
			fmt.Printf("Unknown command: %s\n", command)
			gamelogic.PrintServerHelp()
		}
	}
}

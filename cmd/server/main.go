package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    amqp "github.com/rabbitmq/amqp091-go"
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

    // Wait for a signal to exit
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

    fmt.Println("Waiting for shutdown signal...")
    <-signalChan

    fmt.Println("Shutdown signal received. Closing connection...")
}

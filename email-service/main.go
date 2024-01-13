package main

import (
	"context"
	database "email-service/database/sqlc"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	gmail := NewGmailSender(
		os.Getenv("GMAIL_NAME"),
		os.Getenv("GMAIL_ADDRESS"),
		os.Getenv("GMAIL_PASSWORD"),
	)

	// initializing postgres database
	postgres, conn, err := database.NewPostresDB(context.TODO(), os.Getenv("POSTGRES_ADDRESS"))
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer conn.Close(context.TODO())

	consumer, err := NewKafkaConsumer(
		postgres,
		gmail,
		[]string{
			os.Getenv("KAFKA_ADDRESS"),
		},
		os.Getenv("KAFKA_GROUP"),
		[]string{
			os.Getenv("KAFKA_TOPIC"),
		},
	)
	if err != nil {
		log.Fatal("Error setting up kafka:", err)
	}

	log.Println("Starting consumer...")
	ctx := context.Background()
	log.Fatal(consumer.Process(ctx))
}

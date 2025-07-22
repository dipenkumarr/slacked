package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	fmt.Printf("Sending a message to Slack... %v\n", api)

	channelID, timestamp, err := api.PostMessage(
		"C09650BR7SR",
		slack.MsgOptionText("Hello, world!", false),
	)
	if err != nil {
		fmt.Printf("failed posting message: %v\n", err)
		return
	}

	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

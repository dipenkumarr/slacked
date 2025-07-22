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

	args := os.Args[1:]
	fmt.Println(args)

	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
	preText := "*Hello! Your Jenkins build has completed!*"
	jenkinsURL := "*Build URL:* " + args[0]
	buildResult := "*" + args[1] + "*"
	buildNumber := "*" + args[2] + "*"
	jobName := "*" + args[3] + "*"

	if buildResult == "*SUCCESS*" {
		buildResult = ":white_check_mark: " + buildResult
	} else {
		buildResult = ":x: " + buildResult
	}
}

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/slack-go/slack"
)

func sendSlackMessage(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "<h1>Sent Slack Message!</h1>")

	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	preText := "*Hello! Your Jenkins build has completed!*"
	dividerSection1 := slack.NewDividerBlock()
	preTextField := slack.NewTextBlockObject("mrkdwn", preText+"\n\n", false, false)
	preTextSection := slack.NewSectionBlock(preTextField, nil, nil)

	msg := slack.MsgOptionBlocks(preTextSection, dividerSection1)

	_, _, _, err := api.SendMessage(
		"C09650BR7SR",
		msg,
	)
	if err != nil {
		fmt.Printf("failed posting message: %v\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel\n")

}

func main() {
	http.HandleFunc("/sendSlackMessage", sendSlackMessage)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	http.ListenAndServe(":"+port, nil)
}

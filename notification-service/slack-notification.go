package main

import (
	"fmt"
	"os"

	"github.com/slack-go/slack"
)

func main() {

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

	dividerSection1 := slack.NewDividerBlock()
	jenkinsBuildDetails := jobName + " #" + buildNumber + " - " + buildResult + "\n" + jenkinsURL
	preTextField := slack.NewTextBlockObject("mrkdwn", preText+"\n\n", false, false)
	jenkinsBuildDetailsField := slack.NewTextBlockObject("mrkdwn", jenkinsBuildDetails, false, false)

	jenkinsBuildDetailsSection := slack.NewSectionBlock(jenkinsBuildDetailsField, nil, nil)
	preTextSection := slack.NewSectionBlock(preTextField, nil, nil)

	msg := slack.MsgOptionBlocks(preTextSection, dividerSection1, jenkinsBuildDetailsSection)

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

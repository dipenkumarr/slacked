package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/slack-go/slack"
)

type JenkinsBuild struct {
	BuildURL    string `json:"buildUrl"`
	BuildResult string `json:"buildResult"`
	BuildNumber int    `json:"buildNumber"`
	JobName     string `json:"jobName"`
}

func sendSlackMessage(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "<h1>Sent Slack Message!</h1>")

	build := JenkinsBuild{}

	err := json.NewDecoder(r.Body).Decode(&build)
	if err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}

	api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))

	preText := "*Hello! Your Jenkins build has completed!*"
	jenkinsURL := "*Build URL:* " + build.BuildURL
	buildResult := "*" + build.BuildResult + "*"
	buildNumber := "*" + fmt.Sprint(build.BuildNumber) + "*"
	jobName := "*" + build.JobName + "*"

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

	_, _, _, err = api.SendMessage(
		os.Getenv("SLACK_CHANNEL_ID"),
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

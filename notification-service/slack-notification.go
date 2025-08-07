package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/slack-go/slack"
)

type JenkinsBuild struct {
	BuildURL    string `json:"buildUrl"`
	BuildResult string `json:"buildResult"`
	BuildNumber int    `json:"buildNumber"`
	JobName     string `json:"jobName"`
}

type SlackCredentials struct {
	SlackBotToken  string `json:"SLACK_BOT_TOKEN"`
	SlackChannelID string `json:"SLACK_CHANNEL_ID"`
}

func getSecret(ctx context.Context, secretName string) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	input := secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := client.GetSecretValue(ctx, &input)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret %s, %v", secretName, err)
	}

	return *result.SecretString, nil
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
		fmt.Sprint(os.Getenv("SLACK_CHANNEL_ID")),
		msg,
	)
	if err != nil {
		fmt.Printf("failed posting message: %v\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel\n")

}

func main() {
	// 1. Get the secrets from Secrets Manager.
	secretJSON, err := getSecret(context.Background(), "slacked/slack-credentials")
	if err != nil {
		fmt.Printf("failed to get secret: %v\n", err)
		return
	}
	fmt.Printf("Successfully retrieved raw secret JSON: %s\n", secretJSON)

	var credentials SlackCredentials

	err = json.Unmarshal([]byte(secretJSON), &credentials)
	if err != nil {
		fmt.Printf("failed to unmarshal secret JSON: %v\n", err)
		return
	}
}

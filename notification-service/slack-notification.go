package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {

	// 1. Get the secrets from Secrets Manager.
	secretName := os.Getenv("SECRET_NAME")
	secretJSON, err := getSecret(ctx, secretName)
	if err != nil {
		fmt.Printf("FATAL: failed to get secret: %v\n", err)
		// Return an error to send the entire batch back to the queue for a retry
		return err
	}
	fmt.Println("successfully retrieved secret from Secrets Manager")

	var credentials SlackCredentials

	err = json.Unmarshal([]byte(secretJSON), &credentials)
	if err != nil {
		fmt.Printf("FATAL: failed to unmarshal secret JSON: %v\n", err)
		return err
	}

	// 2. Create a Slack client using the credentials
	api := slack.New(credentials.SlackBotToken)

	// 3. Process each message in the batch
	for _, message := range sqsEvent.Records {
		fmt.Printf("Processing message ID: %s\n", message.MessageId)

		// 3a. Decode the JSON payload from the SQS message body.
		build := JenkinsBuild{}
		err := json.Unmarshal([]byte(message.Body), &build)
		if err != nil {
			fmt.Printf("ERROR: could not unmarshal message body for messageId %s: %v\n", message.MessageId, err)
			// Log the error and continue with the next message
			continue
		}

		// 3b. Prepare the message to be sent to Slack
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

		// 3c. Send the message
		_, _, _, err = api.SendMessage(
			credentials.SlackChannelID,
			msg,
		)
		if err != nil {
			// If sending to Slack fails, this is a transient error, so retry.
			fmt.Printf("ERROR: failed to send Slack message for messageId %s: %v\n", message.MessageId, err)
			return err
		}

		fmt.Printf("Successfully processed and sent notification for message ID: %s\n", message.MessageId)
	}

	// 6. Return a nil on success, indicating all messages were processed.
	return nil

}

func main() {
	lambda.Start(HandleRequest)
}

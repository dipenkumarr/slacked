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

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// 1. Get the secrets from Secrets Manager.
	secretName := os.Getenv("SECRET_NAME")
	secretJSON, err := getSecret(context.Background(), secretName)
	if err != nil {
		fmt.Printf("failed to get secret: %v\n", err)
		return events.APIGatewayProxyResponse{Body: "Internal Server Error", StatusCode: 500}, err
	}
	fmt.Printf("successfully retrieved raw secret JSON: %s\n", secretJSON)

	var credentials SlackCredentials

	err = json.Unmarshal([]byte(secretJSON), &credentials)
	if err != nil {
		fmt.Printf("failed to unmarshal secret JSON: %v\n", err)
		return events.APIGatewayProxyResponse{Body: "Internal Server Error", StatusCode: 500}, err
	}

	// 2. Decode the JSON payload from the request body
	build := JenkinsBuild{}

	err = json.Unmarshal([]byte(request.Body), &build)
	if err != nil {
		fmt.Printf("could not unmarshal request body: %s\n", err)
		return events.APIGatewayProxyResponse{Body: "Bad Request", StatusCode: 400}, err
	}

	// 3. Create a Slack client using the credentials
	api := slack.New(credentials.SlackBotToken)

	// 4. Prepare the message to be sent to Slack
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

	// 5. Send the message
	_, _, _, err = api.SendMessage(
		fmt.Sprint(credentials.SlackChannelID),
		msg,
	)
	if err != nil {
		fmt.Printf("Error sending message: %s\n", err)
		return events.APIGatewayProxyResponse{Body: "Error sending Slack message", StatusCode: 500}, err
	}

	// 6. Return a success response to API Gateway
	return events.APIGatewayProxyResponse{Body: "Message sent successfully!", StatusCode: 200}, nil

}

func main() {
	lambda.Start(HandleRequest)
}

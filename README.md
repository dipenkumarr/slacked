# Slacked

This project provisions a serverless notification service using AWS CDK (Go). It processes Jenkins build events, queues them through AWS SQS, and delivers formatted notifications to Slack via an AWS Lambda function. The system ensures reliable delivery with a Dead-Letter Queue (DLQ) and integrates securely with AWS Secrets Manager for credential management.

## Project Structure

```
dipenkumarr-slacked/
├── infra/                 # AWS CDK infrastructure
│   ├── cdk.json
│   ├── infra.go
│   └── infra_test.go
└── notification-service/  # Go Lambda code and Jenkins pipeline
    ├── Jenkinsfile
    └── slack-notification.go
```

## Features

-   Infrastructure as Code with **AWS CDK**
-   **AWS API Gateway** endpoint for Jenkins webhooks
-   **AWS SQS** queue with **DLQ** for reliable message delivery
-   **Go-based AWS Lambda** for Slack notifications
-   Secure credential retrieval from **AWS Secrets Manager**
-   **Jenkinsfile** for automated build, test, deploy, and notification

## Setup AWS Secrets Manager

To store your Slack credentials securely, create a secret in AWS Secrets Manager. This will be referenced by your Lambda function in the CDK stack.

Run the following AWS CLI command:

```
aws secretsmanager create-secret --name slacked/slack-credentials \
  --secret-string '{"SLACK_BOT_TOKEN":"YOUR_xoxb_TOKEN_HERE","SLACK_CHANNEL_ID":"YOUR_CHANNEL_ID_HERE"}'
```

**Notes:**

-   Replace `YOUR_xoxb_TOKEN_HERE` with your actual Slack Bot Token.
-   Replace `YOUR_CHANNEL_ID_HERE` with your Slack channel's ID.
-   Make sure your AWS CLI is configured for the correct region and account.

## Deployment

1. **Install dependencies:**

    ```bash
    go mod tidy
    ```

2. **Bootstrap CDK (first time only):**

    ```bash
    cdk bootstrap
    ```

3. **Deploy stack:**

    ```bash
    cdk deploy
    ```

## Environment Variables

-   `SECRET_NAME` – Name of the secret in AWS Secrets Manager containing Slack credentials.

## API Endpoint

After deployment, CDK will output the API Gateway URL (e.g., `https://<api-id>.execute-api.<region>.amazonaws.com/prod/send-message`). Use this in your Jenkinsfile `curl` command.

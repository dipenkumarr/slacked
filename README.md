# Slacked

This project uses AWS CDK (Go) to provision infrastructure for an AWS Lambda service that processes Jenkins build events and sends formatted notifications to Slack. It integrates API Gateway for webhook handling, AWS Secrets Manager for secure credential storage, and provides an automated Jenkins pipeline for deployment and updates.

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

-   AWS CDK (Go) infrastructure as code
-   Go-based AWS Lambda function
-   Secure Slack credentials retrieval via AWS Secrets Manager
-   API Gateway endpoint for Jenkins webhook
-   Jenkinsfile to automate build, test, deploy, and Slack notification

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

package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"

	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type InfraStackProps struct {
	awscdk.StackProps
}

func NewInfraStack(scope constructs.Construct, id string, props *InfraStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Dead-Letter Queue
	deadLetterQueue := awssqs.NewQueue(stack, jsii.String("SlackedDeadLetterQueue"), &awssqs.QueueProps{
		QueueName:       jsii.String("slacked-dead-letter-queue"),
		RetentionPeriod: awscdk.Duration_Days(jsii.Number(14)),
	})

	// Main Queue for messages
	mainQueue := awssqs.NewQueue(stack, jsii.String("SlackedMainQueue"), &awssqs.QueueProps{
		QueueName: jsii.String("slacked-main-queue"),
		DeadLetterQueue: &awssqs.DeadLetterQueue{
			MaxReceiveCount: jsii.Number(3),
			Queue:           deadLetterQueue,
		},
		VisibilityTimeout: awscdk.Duration_Seconds(jsii.Number(30)),
	})

	// Get the Slack credentials secret reference
	slackSecret := awssecretsmanager.Secret_FromSecretNameV2(stack, jsii.String("SlackCredentialsSecret"), jsii.String("slacked/slack-credentials"))

	// Create the notification Lambda function
	notificationLambda := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("NotificationLambda"), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("slacked-notification-handler"),
		Entry:        jsii.String("../notification-service"),
		Timeout:      awscdk.Duration_Seconds(jsii.Number(25)),
		Environment: &map[string]*string{
			"SECRET_NAME": slackSecret.SecretName(),
		},
	})

	// Grant the Lambda function permissions to read the Slack credentials secret
	slackSecret.GrantRead(notificationLambda, nil)

	// Connect Lambda to Main SQS
	sqsEventSource := awslambdaeventsources.NewSqsEventSource(mainQueue, nil)
	notificationLambda.AddEventSource(sqsEventSource)

	// Create the API Gateway with SQS integration
	api := awsapigateway.NewRestApi(stack, jsii.String("SlackEndpointAPI"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("Slack Notification Service"),
		Description: jsii.String("Receives notifications from Jenkins and puts them onto the SQS queue"),
	})

	// IAM role to grant the API Gateway permission to send messages to the SQS queue
	apiGatewayRole := awsiam.NewRole(stack, jsii.String("ApiGatewaySqsRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("apigateway.amazonaws.com"), nil),
	})
	mainQueue.GrantSendMessages(apiGatewayRole)

	// SQS Integration to API Gateway
	sqsIntegration := awsapigateway.NewAwsIntegration(&awsapigateway.AwsIntegrationProps{
		Service:               jsii.String("sqs"),
		Path:                  mainQueue.QueueName(),
		IntegrationHttpMethod: jsii.String("POST"),
		Options: &awsapigateway.IntegrationOptions{
			CredentialsRole: apiGatewayRole,
			// Maps incoming req to SQS message API call
			RequestParameters: &map[string]*string{
				"integration.request.header.Content-Type": jsii.String("'application/x-www-form-urlencoded'"),
			},
			// Takes the body of the POST request from jenkins and makes it the body of the SQS message
			RequestTemplates: &map[string]*string{
				"application/json": jsii.String("Action=SendMessage&MessageBody=$util.urlEncoded($input.body)"),
			},
			IntegrationResponses: &[]*awsapigateway.IntegrationResponse{
				{
					StatusCode: jsii.String("200"),
					ResponseTemplates: &map[string]*string{
						"application/json": jsii.String("{\"status\": \"message queued to SQS\"}"),
					},
				},
			},
		},
	})

	// Create the resource for the send-message endpoint
	resource := api.Root().AddResource(jsii.String("send-message"), nil)

	// Add the POST method to the resource
	resource.AddMethod(jsii.String("POST"), sqsIntegration, &awsapigateway.MethodOptions{
		// Response to Jenkins client
		MethodResponses: &[]*awsapigateway.MethodResponse{
			{
				StatusCode: jsii.String("200"),
			},
		},
	})

	// Output the API URL
	awscdk.NewCfnOutput(stack, jsii.String("APIUrl"), &awscdk.CfnOutputProps{
		Value:       api.UrlForPath(resource.Path()),
		Description: jsii.String("The URL for the API Gateway endpoint to be used in Jenkins"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewInfraStack(app, "InfraStack", &InfraStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}

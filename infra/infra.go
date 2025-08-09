package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"

	// "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
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

	// Get the Slack credentials secret reference
	slackSecret := awssecretsmanager.Secret_FromSecretNameV2(stack, jsii.String("SlackCredentialsSecret"), jsii.String("slacked/slack-credentials"))

	// Create the notification Lambda function
	notificationLambda := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("NotificationLambda"), &awscdklambdagoalpha.GoFunctionProps{
		FunctionName: jsii.String("slacked-notification-handler"),
		Entry:        jsii.String("../notification-service"),
		Environment: &map[string]*string{
			"SECRET_NAME": slackSecret.SecretName(),
		},
	})

	// Grant the Lambda function permissions to read the Slack credentials secret
	slackSecret.GrantRead(notificationLambda, nil)

	// Create the API Gateway
	api := awsapigateway.NewRestApi(stack, jsii.String("SlackEndpointAPI"), &awsapigateway.RestApiProps{
		RestApiName: jsii.String("Slack Notification Service"),
		Description: jsii.String("Receives notifications from Jenkins"),
	})

	// Create the resource for the send-message endpoint
	resource := api.Root().AddResource(jsii.String("send-message"), nil)

	// Create the Lambda integration for the send-message endpoint
	lambdaIntegration := awsapigateway.NewLambdaIntegration(notificationLambda, nil)

	// Add the POST method to the resource
	resource.AddMethod(jsii.String("POST"), lambdaIntegration, nil)

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

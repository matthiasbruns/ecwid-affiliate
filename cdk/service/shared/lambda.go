package shared

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/jsii-runtime-go"
)

type SpawnLambdaOptions struct {
	Stage         string
	Name          string
	Path          string
	Timeout       awscdk.Duration
	Memory        float64
	DynamoARNs    []*string
	DynamoActions []*string
	CustomEnv     map[string]*string
	LogRetention  *awslogs.RetentionDays
}

func SpawnLambda(stack awscdk.Stack, opts *SpawnLambdaOptions) awslambda.IFunction {
	stage := opts.Stage
	name := opts.Name
	path := opts.Path
	timeout := opts.Timeout
	if timeout == nil {
		timeout = awscdk.Duration_Seconds(jsii.Number(30))
	}
	memory := opts.Memory
	if memory == 0 {
		memory = 128
	}
	dynamoARNs := opts.DynamoARNs
	customEnv := opts.CustomEnv

	// lambda execution role
	lambdaRole := awsiam.NewRole(stack, jsii.String(fmt.Sprintf("%s-role", name)), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), &awsiam.ServicePrincipalOpts{}),
	})
	lambdaRole.AddManagedPolicy(
		awsiam.ManagedPolicy_FromManagedPolicyArn(stack,
			jsii.String(fmt.Sprintf("%s-AWSLambdaBasicExecutionRole", name)),
			jsii.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
		),
	)
	lambdaRole.AddManagedPolicy(
		awsiam.ManagedPolicy_FromManagedPolicyArn(stack,
			jsii.String(fmt.Sprintf("%s-CloudWatchLambdaInsightsExecutionRolePolicy", name)),
			jsii.String("arn:aws:iam::aws:policy/CloudWatchLambdaInsightsExecutionRolePolicy"),
		),
	)
	lambdaRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions: &[]*string{
			jsii.String("ssm:Get*"),
		},
		Resources: &[]*string{
			jsii.String(fmt.Sprintf("arn:aws:ssm:%s:%s:*", *stack.Region(), *stack.Account())),
		},
	}))

	if len(dynamoARNs) > 0 {
		actions := opts.DynamoActions

		if len(actions) == 0 {
			actions = []*string{
				jsii.String("dynamodb:List*"),
				jsii.String("dynamodb:Get*"),
				jsii.String("dynamodb:Query*"),
				jsii.String("dynamodb:Put*"),
				jsii.String("dynamodb:Scan"),
				jsii.String("dynamodb:Update*"),
			}
		}

		lambdaRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Actions:   &actions,
			Resources: &dynamoARNs,
		}))
	}

	// lambda insights layer
	layer := awslambda.LayerVersion_FromLayerVersionArn(stack,
		jsii.String(fmt.Sprintf("%s-insights", name)),
		jsii.String("arn:aws:lambda:eu-central-1:580247275435:layer:LambdaInsightsExtension-Arm64:5"),
	)

	env := map[string]*string{
		"STAGE":           jsii.String(stage),
		"SSM_SENTRY_DNS":  jsii.String("/ecwid-lexoffice/sentry/dns"),
		"SSM_ECWID_TOKEN": jsii.String("/ecwid-lexoffice/ecwid/token"),
	}
	for k, v := range customEnv {
		env[k] = v
	}

	if opts.LogRetention == nil {
		days := awslogs.RetentionDays_THREE_DAYS
		opts.LogRetention = &days
	}
	logGroup := awslogs.NewLogGroup(stack, jsii.String(fmt.Sprintf("%s-log-group", name)), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String(fmt.Sprintf("/aws/lambda/%s", name)),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Retention:     *opts.LogRetention,
	})

	lambda := awslambda.NewFunction(stack, jsii.String(fmt.Sprintf("%s-lambda", name)), &awslambda.FunctionProps{
		Role:         lambdaRole,
		FunctionName: jsii.String(name),
		Layers:       &[]awslambda.ILayerVersion{layer},
		Architecture: awslambda.Architecture_ARM_64(),
		Runtime:      awslambda.Runtime_PROVIDED_AL2(),
		MemorySize:   jsii.Number(256),
		Timeout:      timeout,
		LogGroup:     logGroup,
		Code:         awslambda.AssetCode_FromAsset(jsii.String(fmt.Sprintf("../%s", path)), &awss3assets.AssetOptions{}),
		Handler:      jsii.String("bootstrap"),
		Environment:  &env,
	})

	awscdk.NewCfnOutput(stack, jsii.String(name), &awscdk.CfnOutputProps{
		Value: lambda.FunctionArn(),
	})

	return lambda
}

func RetentionDaysPtr(v awslogs.RetentionDays) *awslogs.RetentionDays {
	return &v
}

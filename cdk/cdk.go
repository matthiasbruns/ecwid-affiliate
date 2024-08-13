package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"

	"github.com/matthiasbruns/ecwid-affiliate/cdk/global"
	"github.com/matthiasbruns/ecwid-affiliate/cdk/service"
	"github.com/matthiasbruns/ecwid-affiliate/cdk/ses"
	"github.com/matthiasbruns/ecwid-affiliate/cdk/web"
)

const serviceName = "ecwid-affiliate"
const domain = "ecom-affiliate-link.appetizers.io"

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	stage := os.Getenv("STAGE")
	if stage == "" {
		stage = "dev"
	}
	fmt.Printf("Deploying stage: %s\n", stage)

	baseDomain := ""
	if stage == "prod" {
		baseDomain = domain
	} else {
		baseDomain = fmt.Sprintf("%s.%s", stage, domain)
	}

	globalStack, globalResources := global.Stack(app, fmt.Sprintf("%s-global", serviceName), &global.StackProps{
		Stage:      stage,
		BaseDomain: baseDomain,
		StackProps: awscdk.StackProps{
			CrossRegionReferences: jsii.Bool(true),
			Env: &awscdk.Environment{
				Region: jsii.String("us-east-1"),
			},
		},
	})

	ses.Stack(app, fmt.Sprintf("%s-ses", serviceName), &ses.StackProps{
		Stage:      stage,
		BaseDomain: baseDomain,
		HostedZone: globalResources.HostedZone,
		StackProps: awscdk.StackProps{
			CrossRegionReferences: jsii.Bool(true),
			Env:                   env(),
		},
	}).AddDependency(globalStack, jsii.String("Requires Route53 hosted zone"))

	service.Stack(app, fmt.Sprintf("%s-service", serviceName), &service.StackProps{
		Stage:      stage,
		BaseDomain: baseDomain,
		HostedZone: globalResources.HostedZone,
		StackProps: awscdk.StackProps{
			CrossRegionReferences: jsii.Bool(true),
			Env:                   env(),
		},
	}).AddDependency(globalStack, jsii.String("Requires ACME certificate and Route53 hosted zone"))

	web.Stack(app, fmt.Sprintf("%s-web", serviceName), &web.StackProps{
		Stage:       stage,
		ServiceName: serviceName,
		BaseDomain:  baseDomain,
		HostedZone:  globalResources.HostedZone,
		Certificate: globalResources.Certificate,
		StackProps: awscdk.StackProps{
			CrossRegionReferences: jsii.Bool(true),
			Env:                   env(),
		},
	}).AddDependency(globalStack, jsii.String("Requires ACME certificate and Route53 hosted zone"))

	awscdk.Tags_Of(app).Add(jsii.String("Environment"), jsii.String(stage), nil)
	awscdk.Tags_Of(app).Add(jsii.String("Project"), jsii.String("ecwid-affiliate"), nil)
	awscdk.Tags_Of(app).Add(jsii.String("Repository"), jsii.String("https://github.com/matthiasbruns/ecwid-affiliate"), nil)

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// try to get env from OS first
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("CDK_DEFAULT_REGION")
	}
	if region == "" {
		region = "eu-central-1"
	}
	return &awscdk.Environment{
		Region: jsii.String(region),
	}
}

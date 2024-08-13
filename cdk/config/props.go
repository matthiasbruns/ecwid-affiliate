package config

import "github.com/aws/aws-cdk-go/awscdk/v2"

type CdkStackProps struct {
	awscdk.StackProps
	BaseDomain string
}

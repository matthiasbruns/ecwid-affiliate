package web

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsssm"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type StackProps struct {
	awscdk.StackProps
	Stage       string
	ServiceName string
	BaseDomain  string
	HostedZone  awsroute53.IHostedZone
	Certificate awscertificatemanager.ICertificate
}

func Stack(scope constructs.Construct, id string, props *StackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	serviceName := props.ServiceName

	webDomain := fmt.Sprintf("iframe.%s", props.BaseDomain)

	awsssm.NewStringParameter(stack, jsii.String(fmt.Sprintf("%s-iframe-domain", serviceName)), &awsssm.StringParameterProps{
		ParameterName: jsii.String("/ecwid-lexoffice/lexoffice/iframe-url"),
		StringValue:   jsii.String(webDomain),
	})

	webBucket := awss3.NewBucket(stack, jsii.String(fmt.Sprintf("%s-web-bucket", serviceName)), &awss3.BucketProps{
		BucketName:           jsii.String(webDomain),
		WebsiteIndexDocument: jsii.String("index.html"),
		WebsiteErrorDocument: jsii.String("error.html"),
		BlockPublicAccess:    awss3.BlockPublicAccess_BLOCK_ACLS(),
	})

	webBucket.AddToResourcePolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("s3:GetObject"),
		},
		Principals: &[]awsiam.IPrincipal{
			awsiam.NewAnyPrincipal(),
		},
		Resources: &[]*string{
			jsii.String(fmt.Sprintf("arn:aws:s3:::%s/*", *webBucket.BucketName())),
		},
	}))

	_ = awss3deployment.NewBucketDeployment(stack, jsii.String(fmt.Sprintf("%s-web-bucket-deployment", serviceName)), &awss3deployment.BucketDeploymentProps{
		DestinationBucket: webBucket,
		Sources: &[]awss3deployment.ISource{
			awss3deployment.Source_Asset(jsii.String("../web/build"), &awss3assets.AssetOptions{
				AssetHashType: awscdk.AssetHashType_SOURCE,
			}),
		},
	})

	cloudfront := awscloudfront.NewDistribution(stack, jsii.String(fmt.Sprintf("%s-distribution", webDomain)), &awscloudfront.DistributionProps{
		DomainNames: &[]*string{jsii.String(webDomain)},
		DefaultBehavior: &awscloudfront.BehaviorOptions{
			Origin: awscloudfrontorigins.NewS3Origin(webBucket, &awscloudfrontorigins.S3OriginProps{}),
		},
		Certificate: props.Certificate,
	})

	awsroute53.NewARecord(stack, jsii.String(fmt.Sprintf("%s-web-record", serviceName)), &awsroute53.ARecordProps{
		Zone:       props.HostedZone,
		RecordName: jsii.String("iframe"),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewCloudFrontTarget(cloudfront)),
	})

	return stack
}

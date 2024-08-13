package service

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type StackProps struct {
	awscdk.StackProps
	Stage      string
	BaseDomain string
	HostedZone awsroute53.IHostedZone
}

func Stack(scope constructs.Construct, id string, props *StackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	apiDomain := fmt.Sprintf("api.%s", props.BaseDomain)

	certificate := awscertificatemanager.NewCertificate(stack, jsii.String(fmt.Sprintf("%s-cert", apiDomain)), &awscertificatemanager.CertificateProps{
		DomainName:              jsii.String(apiDomain),
		SubjectAlternativeNames: &[]*string{jsii.String("*." + apiDomain)},
		Validation:              awscertificatemanager.CertificateValidation_FromDns(props.HostedZone),
	})

	gatewayApi := awsapigateway.NewRestApi(stack, jsii.String(apiDomain), &awsapigateway.RestApiProps{
		DefaultCorsPreflightOptions: &awsapigateway.CorsOptions{
			AllowOrigins:     awsapigateway.Cors_ALL_ORIGINS(),
			AllowMethods:     awsapigateway.Cors_ALL_METHODS(),
			AllowCredentials: jsii.Bool(true),
		},
		Deploy: jsii.Bool(true),
		DeployOptions: &awsapigateway.StageOptions{
			LoggingLevel:   awsapigateway.MethodLoggingLevel_INFO,
			MetricsEnabled: jsii.Bool(true),
			StageName:      jsii.String("prod"),
			TracingEnabled: jsii.Bool(true),
		},
		DomainName: &awsapigateway.DomainNameOptions{
			Certificate:    certificate,
			DomainName:     jsii.String(apiDomain),
			EndpointType:   awsapigateway.EndpointType_REGIONAL,
			SecurityPolicy: awsapigateway.SecurityPolicy_TLS_1_2,
		},
		CloudWatchRole: jsii.Bool(true),
	})

	awsroute53.NewARecord(stack, jsii.String(fmt.Sprintf("%s-api-record", props.BaseDomain)), &awsroute53.ARecordProps{
		Zone:       props.HostedZone,
		RecordName: jsii.String("api"),
		Target:     awsroute53.RecordTarget_FromAlias(awsroute53targets.NewApiGateway(gatewayApi)),
	})

	return stack
}

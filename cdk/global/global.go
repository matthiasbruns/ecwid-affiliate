package global

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type StackProps struct {
	awscdk.StackProps
	Stage      string
	BaseDomain string
}

type StackResources struct {
	HostedZone  awsroute53.IHostedZone
	Certificate awscertificatemanager.ICertificate
	Cloudfront  awscloudfront.Distribution
}

func Stack(scope constructs.Construct, id string, props *StackProps) (awscdk.Stack, *StackResources) {
	stackResources := new(StackResources)

	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	baseDomain := props.BaseDomain

	hostedZone := awsroute53.NewHostedZone(stack, jsii.String(fmt.Sprintf("%s-zone", baseDomain)), &awsroute53.HostedZoneProps{
		ZoneName: jsii.String(baseDomain),
	})

	certificate := awscertificatemanager.NewCertificate(stack, jsii.String(fmt.Sprintf("%s-cert", baseDomain)), &awscertificatemanager.CertificateProps{
		DomainName:              jsii.String(baseDomain),
		SubjectAlternativeNames: &[]*string{jsii.String("*." + baseDomain)},
		Validation:              awscertificatemanager.CertificateValidation_FromDns(hostedZone),
	})

	if props.Stage == "prod" {
		awsroute53.NewNsRecord(stack, jsii.String(fmt.Sprintf("dev.%s-ns", baseDomain)), &awsroute53.NsRecordProps{
			Zone:       hostedZone,
			RecordName: jsii.String("dev"),
			Values: &[]*string{
				jsii.String("ns-1635.awsdns-12.co.uk."),
				jsii.String("ns-674.awsdns-20.net."),
				jsii.String("ns-400.awsdns-50.com."),
				jsii.String("ns-1116.awsdns-11.org."),
			},
		})

		awsroute53.NewARecord(stack, jsii.String(fmt.Sprintf("%s-a-records", baseDomain)), &awsroute53.ARecordProps{
			Zone:    hostedZone,
			Comment: jsii.String("GitHub routing for matthiasbruns.github.io"),
			Target: awsroute53.RecordTarget_FromIpAddresses(
				jsii.String("185.199.108.153"),
				jsii.String("185.199.109.153"),
				jsii.String("185.199.110.153"),
				jsii.String("185.199.111.153"),
			),
		})

		awsroute53.NewAaaaRecord(stack, jsii.String(fmt.Sprintf("%s-aaaa-records", baseDomain)), &awsroute53.AaaaRecordProps{
			Zone:    hostedZone,
			Comment: jsii.String("GitHub routing for matthiasbruns.github.io"),
			Target: awsroute53.RecordTarget_FromIpAddresses(
				jsii.String("2606:50c0:8000::153"),
				jsii.String("2606:50c0:8001::153"),
				jsii.String("2606:50c0:8002::153"),
				jsii.String("2606:50c0:8003::153"),
			),
		})

		awsroute53.NewCnameRecord(stack, jsii.String(fmt.Sprintf("%s-www-cname", baseDomain)), &awsroute53.CnameRecordProps{
			Zone:       hostedZone,
			RecordName: jsii.String("www"),
			DomainName: jsii.String("matthiasbruns.github.io"),
		})
	}

	stackResources.HostedZone = hostedZone
	stackResources.Certificate = certificate

	awscdk.NewCfnOutput(stack, jsii.String("HostedZoneId"), &awscdk.CfnOutputProps{
		Value: hostedZone.HostedZoneId(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("HostedZoneArn"), &awscdk.CfnOutputProps{
		Value: hostedZone.HostedZoneArn(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("CertificateArn"), &awscdk.CfnOutputProps{
		Value: certificate.CertificateArn(),
	})

	return stack, stackResources
}

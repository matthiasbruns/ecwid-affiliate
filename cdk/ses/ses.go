package ses

import (
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsses"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"log"
)

type records struct {
	txtRecordValues []*string
	dkimKey         []*string
}

var recordStageMap = map[string]records{
	"dev": {
		txtRecordValues: []*string{jsii.String("google-site-verification=6H1yDN4ZKS-cKAgyOs5v9Osf2FQqnBBFfTOZwPsrEKs")},
		dkimKey:         []*string{jsii.String("v=DKIM1; k=rsa; MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAgKb+EODlxidd2XQ+4os1bTZyM+w2PCj2VPwGgn/Hnk5Gjn0X7E69H0s8lwffjBGtVAH14W1nv2qXccLXmZLf+Oh9IxN+dwROYW7nF72Hj1GGP4+c9AAfvtcCU0ijVv6YlusKkO42BlY1+rvp7fqFIWxHSMLlyNeSOZ8dBnu1kG6OWsvoTK4LkHC/Wr9V6PxOJjBakIGqVwToxNrVcHILSnQPsGABZX1Xjk7m05DAo3/5K5gZ233OPjtKEphIH8i9LNwGcNJLuQkJLI57wxvjkkXXQcSLy2D5OaOmBxoS8h6Nc4ROGYptvUMacpKTnYBlXe6R7CD3XSgOFYmIxJfp8QIDAQAB")},
	},
	"prod": {
		txtRecordValues: []*string{jsii.String("google-site-verification=lhlPN3EZZZqD9v_yHvrzb20VzS0JVkdUE-QfJAjH0MY")},
		dkimKey:         []*string{jsii.String("v=DKIM1; k=rsa; MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAgfXb1hy6MipwlDMs1RyYpFlk/ulBMm5SX67f01HtR9kU0pGsk6HEsqIXMhqiDd3kEXCvJXByNqs6FO9Rtb+lwYZXq0nzu0KYCzYsdDtyRaHD5s5a6QTn+LoDACDpKpFGZk2DrUsXqJFvr4N1Bufh6CfxxICOq+K9ooc5StPEA7shzLJ7ibCkxjbJk4RNG9z4brljPrf0zKXBNYDfRR6a6+Soyimflgr8/DGhlggbCymHF2UBggpp+zvAowG4uI6fY6hLjQhV2KpB/73+aVI65ETrK9i8yvITNbHjhdtz/HQ5Y2vy4Y3Uqx122CT2uH2ciitnHgEIiHfiH6b7J8VS1wIDAQAB")},
	},
}

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

	noReplyMail := fmt.Sprintf("no-reply@%s", props.BaseDomain)
	dmarcMail := fmt.Sprintf("dmarc-reports@%s", props.BaseDomain)

	r, ok := recordStageMap[props.Stage]
	if !ok {
		log.Fatalln(fmt.Sprintf("no records for stage %s", props.Stage))
	}

	awsses.NewEmailIdentity(stack, jsii.String(fmt.Sprintf("%s-email-identity", props.BaseDomain)), &awsses.EmailIdentityProps{
		Identity: awsses.Identity_Email(jsii.String(noReplyMail)),
	})

	awsroute53.NewTxtRecord(stack, jsii.String(fmt.Sprintf("%s-mail-txt-record", props.BaseDomain)), &awsroute53.TxtRecordProps{
		Zone:    props.HostedZone,
		Comment: jsii.String("Google TXT verification"),
		Ttl:     awscdk.Duration_Seconds(jsii.Number(300)),
		Values:  &r.txtRecordValues,
	})

	awsroute53.NewTxtRecord(stack, jsii.String(fmt.Sprintf("%s-mail-dkim-record", props.BaseDomain)), &awsroute53.TxtRecordProps{
		Zone:       props.HostedZone,
		RecordName: jsii.String("google._domainkey"),
		Comment:    jsii.String("Google DKIM settings"),
		Ttl:        awscdk.Duration_Seconds(jsii.Number(300)),
		Values:     &r.dkimKey,
	})

	awsroute53.NewTxtRecord(stack, jsii.String(fmt.Sprintf("%s-mail-spf-record", props.BaseDomain)), &awsroute53.TxtRecordProps{
		Zone:       props.HostedZone,
		RecordName: jsii.String("_spf"),
		Comment:    jsii.String("Google SPF settings"),
		Ttl:        awscdk.Duration_Seconds(jsii.Number(300)),
		Values:     &[]*string{jsii.String("v=spf1 include:_spf.google.com ~all")},
	})

	awsroute53.NewTxtRecord(stack, jsii.String(fmt.Sprintf("%s-mail-dmark-record", props.BaseDomain)), &awsroute53.TxtRecordProps{
		Zone:       props.HostedZone,
		RecordName: jsii.String("_dmark"),
		Comment:    jsii.String("Google DMARK settings"),
		Ttl:        awscdk.Duration_Seconds(jsii.Number(300)),
		Values:     &[]*string{jsii.String(fmt.Sprintf("v=DMARC1; p=none; rua=mailto:%s", dmarcMail))},
	})

	awsroute53.NewMxRecord(stack, jsii.String(fmt.Sprintf("%s-mail-mx-record", props.BaseDomain)), &awsroute53.MxRecordProps{
		Zone: props.HostedZone,
		Values: &[]*awsroute53.MxRecordValue{
			{
				HostName: jsii.String("ASPMX.L.GOOGLE.COM"),
				Priority: jsii.Number(1),
			},
			{
				HostName: jsii.String("ALT1.ASPMX.L.GOOGLE.COM"),
				Priority: jsii.Number(5),
			},
			{
				HostName: jsii.String("ALT2.ASPMX.L.GOOGLE.COM"),
				Priority: jsii.Number(5),
			},
			{
				HostName: jsii.String("ALT3.ASPMX.L.GOOGLE.COM"),
				Priority: jsii.Number(10),
			},
			{
				HostName: jsii.String("ALT4.ASPMX.L.GOOGLE.COM"),
				Priority: jsii.Number(10),
			},
		},
	})

	return stack
}

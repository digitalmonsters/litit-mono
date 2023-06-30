package sqs_listener

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/digitalmonsters/go-common/boilerplate"
)

func InitSQS(conf boilerplate.SQSConfiguration) *sqs.SQS {
	// Create a new AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(conf.Region),
		},
	}))

	return sqs.New(sess)
}

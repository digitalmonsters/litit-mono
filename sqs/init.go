package sqs

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/digitalmonsters/go-common/boilerplate"
)

func InitSQS(conf boilerplate.SQSConfiguration) *sqs.SQS {
	// Create a new AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(conf.Region), // Replace with your desired region
	})
	if err != nil {
		log.Fatal(err)
	}

	return sqs.New(sess)
}

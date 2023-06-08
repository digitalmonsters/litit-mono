package sqs

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/digitalmonsters/go-common/boilerplate"
)

type Command struct {
	Conf    boilerplate.SQSConfiguration
	Message map[string]interface{}
	Svc     *sqs.SQS
}

func (c *Command) SendMessage() (*sqs.SendMessageOutput, error) {
	message, err := marshalMessage(c.Message)
	if err != nil {
		return nil, err
	}

	result, err := c.Svc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    aws.String(c.Conf.Url),
		MessageBody: aws.String(message),
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func marshalMessage(o map[string]interface{}) (string, error) {
	s, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

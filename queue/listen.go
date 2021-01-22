package queue

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// listen - A worker loop that reads and processes queue messages.
func (q *queueSQS) listen(fn MessageHandler) error {
	params := sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(q.URL),
		MaxNumberOfMessages: aws.Int64(maxNumberOfMessages),
		MessageAttributeNames: []*string{
			aws.String("All"), // Required
		},
		WaitTimeSeconds: aws.Int64(waitTimeSeconds),
	}

	log.Info("Starting the listen process")
	for {
		resp, err := q.SQS.ReceiveMessage(&params)

		if err != nil {
			return errors.Wrap(err, "SQS.ReceiveMessage error")
		}

		if len(resp.Messages) > 0 {
			for _, msg := range resp.Messages {
				if err := q.handleMessage(fn, msg); err != nil {
					log.Errorf("handling queue message: %v", err)
				}
			}
		}

	}
}

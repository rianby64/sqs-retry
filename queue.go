package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"

	// nolint: depguard
	log "github.com/sirupsen/logrus"
)

// PutString sends an string to the queue
func (q *queueSQS) PutString(method, msg string, delaySeconds int64) *sqsResponseThenable {
	thenable := &sqsResponseThenable{
		queue: q,
	}

	if q.NextDelayIncreaseSeconds == 0 {
		q.NextDelayIncreaseSeconds = nextDelayIncreaseSecondsDefault
	}

	nextDelay := delaySeconds + q.NextDelayIncreaseSeconds
	messageAttributes := map[string]*sqs.MessageAttributeValue{
		"NextDelayRetry": {
			DataType:    aws.String("Number"),
			StringValue: aws.String(fmt.Sprintf("%d", nextDelay)),
		},
	}

	if method != "" {
		messageAttributes["Method"] = &sqs.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(method),
		}
	}

	params := sqs.SendMessageInput{
		QueueUrl:          aws.String(q.URL),
		MessageBody:       aws.String(msg),
		DelaySeconds:      aws.Int64(delaySeconds),
		MessageAttributes: messageAttributes,
	}

	response, err := q.SQS.SendMessage(&params)
	if err != nil {
		thenable.Error = err
		return thenable
	}

	thenable.messageID = aws.StringValue(response.MessageId)
	q.thens[thenable.messageID] = []MessageHandler{}

	return thenable
}

// PutString sends a JSON to the queue
func (q *queueSQS) PutJSON(method string, msg interface{}, delaySeconds int64) *sqsResponseThenable {
	thenable := &sqsResponseThenable{}
	msgBytes, err := json.Marshal(msgJSON{
		Msg: msg,
	})

	if err != nil {
		thenable.Error = errors.Wrap(err, "PutJSON error")
		return thenable
	}

	return q.PutString(method, string(msgBytes), delaySeconds)
}

// Register method
func (q *queueSQS) Register(name string, method MessageHandler) {
	if q.handlerMap == nil {
		q.handlerMap = map[string]MessageHandler{}
	}

	q.handlerMap[name] = method
}

// Listen method
func (q *queueSQS) Listen() {
	for {
		if err := q.listen(); err != nil {
			log.Error(err, "terminated, retry to listen... wait")
		}

		time.Sleep(time.Duration(retrySecondsToListen) * time.Second)
	}
}

// NewSQSQueue jajaja
func NewSQSQueue(sqssession iSQSSession, url string) SQSQueue {
	queue := queueSQS{
		SQS:                      sqssession,
		URL:                      url,
		TimeoutSeconds:           timeoutSecondsDefault,
		NextDelayIncreaseSeconds: nextDelayIncreaseSecondsDefault,
		handlerMap:               map[string]MessageHandler{},
		msgIDerrs:                map[string]int{},
		thens:                    map[string][]MessageHandler{},
	}

	return &queue
}

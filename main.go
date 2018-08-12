package sqsworker

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// PartialSQSClient is an interface that describes a partial interface for an SQS client
// that can be used to delete messages.
type PartialSQSClient interface {
	DeleteMessage(input *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error)
}

// MessageProcessor is a function that will handle a single SQS message from a batch.
type MessageProcessor func(msg events.SQSMessage) error

// Handler is used for creating Lambdas that can process batches of SQS events.
type Handler struct {
	sqsClient PartialSQSClient
	process   MessageProcessor
}

// NewHandler creates an Handler instance using an SQS client instance and the
// processing function that handles the each message.
func NewHandler(sqsClient PartialSQSClient, processor MessageProcessor) *Handler {
	return &Handler{
		sqsClient: sqsClient,
		process:   processor,
	}
}

// handleMessage will handle a single SQS message from the batch provided.  If the message
// is able to be completed, then it will attempt to delete the message from SQS.
func (s *Handler) handleMessage(ch chan error, msg events.SQSMessage) {
	// process the message using the provided processor
	err := s.process(msg)

	// if we've reached this point with no error, then let's try and remove the message from SQS
	if err == nil {
		queueURL := convertARN2URL(msg.EventSourceARN)

		_, err = s.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
			ReceiptHandle: &msg.ReceiptHandle,
			QueueUrl:      &queueURL,
		})
	}

	ch <- err
}

// ProcessMessages handles a batch of SQS messages and returns the total number of
// successfully processed messages and any error that has occurred.
func (s *Handler) ProcessMessages(messages []events.SQSMessage) (completed int, err error) {
	count := len(messages)

	// check to see if there are any messages and report if there are none
	if count == 0 {
		return
	}

	// create a buffered channel for handling processed messages
	results := make(chan error, count)

	// process the messages in parallel
	for _, message := range messages {
		go s.handleMessage(results, message)
	}

	// wait on the processed messages and tally the results
	for i := 0; i < count; i++ {
		err := <-results
		if err == nil {
			completed++
		}
	}

	if completed != count {
		err = errors.New("failed to complete all given messages")
	}

	return
}

// Handle is the method responsible for processing each batch of messages for
// an SQS worker Lambda.
func (s *Handler) Handle(ctx context.Context, ev events.SQSEvent) error {
	completed, err := s.ProcessMessages(ev.Records)

	// print a status message to our logs
	fmt.Printf("%d message(s) received, %d closed\n", len(ev.Records), completed)

	return err
}

// convertARN2URL converts the ARN of an SQS queue to the URL version.
func convertARN2URL(arn string) string {
	parts := strings.Split(arn, ":")
	return "https://" + parts[2] + "." + parts[3] + ".amazonaws.com/" + parts[4] + "/" + parts[5]
}

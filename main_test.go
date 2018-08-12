package sqsworker

import "testing"

func TestNewHandler(t *testing.T) {

}

func TestHandleMessage(t *testing.T) {

}

func TestProcessMessages(t *testing.T) {}

func TestHandle(t *testing.T) {}

func TestConvertARN2URL(t *testing.T) {
	arn := "arn:aws:sqs:us-west-2:123456:my_queue_name"
	expected := "https://sqs.us-west-2.amazonaws.com/123456/my_queue_name"

	url := convertARN2URL(arn)

	if url != expected {
		t.Errorf("expected %v to equal %v", url, expected)
	}
}

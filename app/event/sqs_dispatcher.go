package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

const (
	// sqsTimeout is the timeout for SQS operations.
	sqsTimeout = 1 * time.Second
)

// sqsClient defines the interface for SQS operations used by SQSDispatcher.
// This allows for mocking in tests.
type sqsClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

// awsConfigLoader is a function type for loading AWS config.
// This allows for mocking in tests.
type awsConfigLoader func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error)

// sqsClientFactory creates an SQS client from an AWS config.
// This allows for mocking in tests.
type sqsClientFactory func(cfg *aws.Config) sqsClient

// defaultSQSClientFactory creates a real SQS client from an AWS config.
func defaultSQSClientFactory(cfg *aws.Config) sqsClient {
	return sqs.NewFromConfig(*cfg)
}

// SQSDispatcher sends events to an AWS SQS queue.
type SQSDispatcher struct {
	client   sqsClient
	queueURL string
}

// NewSQSDispatcher creates a new SQS dispatcher.
// queueURL is the URL of the SQS queue to send events to.
// region is the AWS region (e.g., "eu-west-1"). If empty, the default region from AWS config is used.
func NewSQSDispatcher(ctx context.Context, queueURL, region string) (*SQSDispatcher, error) {
	return newSQSDispatcher(ctx, queueURL, region, config.LoadDefaultConfig, defaultSQSClientFactory)
}

// newSQSDispatcher is the internal implementation that accepts dependencies for testing.
func newSQSDispatcher(
	ctx context.Context,
	queueURL, region string,
	configLoader awsConfigLoader,
	clientFactory sqsClientFactory,
) (*SQSDispatcher, error) {
	var awsConfig aws.Config
	var err error

	if region != "" {
		awsConfig, err = configLoader(ctx, config.WithRegion(region))
	} else {
		awsConfig, err = configLoader(ctx)
	}
	if err != nil {
		return nil, err
	}

	client := clientFactory(&awsConfig)

	return &SQSDispatcher{
		client:   client,
		queueURL: queueURL,
	}, nil
}

// newSQSDispatcherWithClient creates an SQS dispatcher with a custom client (for testing).
func newSQSDispatcherWithClient(client sqsClient, queueURL string) *SQSDispatcher {
	return &SQSDispatcher{
		client:   client,
		queueURL: queueURL,
	}
}

// Dispatch sends an event to the SQS queue.
// Uses a 1 second timeout to avoid blocking requests.
func (d *SQSDispatcher) Dispatch(ctx context.Context, event *Event) error {
	// Apply timeout to avoid blocking the request too long
	ctx, cancel := context.WithTimeout(ctx, sqsTimeout)
	defer cancel()

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = d.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &d.queueURL,
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"event_type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.Type),
			},
		},
	})

	return err
}

// Ensure SQSDispatcher implements Dispatcher.
var _ Dispatcher = (*SQSDispatcher)(nil)

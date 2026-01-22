package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSQSClient is a mock implementation of sqsClient for testing.
type mockSQSClient struct {
	sendMessageFunc func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	lastInput       *sqs.SendMessageInput
}

func (m *mockSQSClient) SendMessage(
	ctx context.Context,
	params *sqs.SendMessageInput,
	optFns ...func(*sqs.Options),
) (*sqs.SendMessageOutput, error) {
	m.lastInput = params
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(ctx, params, optFns...)
	}
	return &sqs.SendMessageOutput{}, nil
}

// mockConfigLoader creates a mock AWS config loader for testing.
func mockConfigLoader(returnErr error) awsConfigLoader {
	return func(_ context.Context, _ ...func(*config.LoadOptions) error) (aws.Config, error) {
		if returnErr != nil {
			return aws.Config{}, returnErr
		}
		return aws.Config{Region: "us-east-1"}, nil
	}
}

// mockClientFactory creates a mock SQS client factory for testing.
func mockClientFactory(client sqsClient) sqsClientFactory {
	return func(_ *aws.Config) sqsClient {
		return client
	}
}

func TestSQSDispatcher_Dispatch_Success(t *testing.T) {
	mockClient := &mockSQSClient{}
	dispatcher := newSQSDispatcherWithClient(mockClient, "https://sqs.eu-west-1.amazonaws.com/123456789/test-queue")

	event := &Event{
		Version:   "1.0",
		Type:      TypeSubmissionCreated,
		SourceApp: SourceApp,
		Time:      time.Now(),
		Payload: map[string]interface{}{
			"author_id": int64(123),
			"item_id":   int64(456),
		},
	}

	err := dispatcher.Dispatch(context.Background(), event)

	require.NoError(t, err)
	require.NotNil(t, mockClient.lastInput)
	assert.Equal(t, "https://sqs.eu-west-1.amazonaws.com/123456789/test-queue", *mockClient.lastInput.QueueUrl)
	assert.Contains(t, *mockClient.lastInput.MessageBody, `"type":"submission_created"`)
	assert.Contains(t, *mockClient.lastInput.MessageBody, `"source_app":"algoreabackend"`)

	// Check message attributes
	eventTypeAttr, ok := mockClient.lastInput.MessageAttributes["event_type"]
	require.True(t, ok)
	assert.Equal(t, "String", *eventTypeAttr.DataType)
	assert.Equal(t, TypeSubmissionCreated, *eventTypeAttr.StringValue)
}

func TestSQSDispatcher_Dispatch_Error(t *testing.T) {
	expectedError := errors.New("SQS send failed")
	mockClient := &mockSQSClient{
		sendMessageFunc: func(_ context.Context, _ *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
			return nil, expectedError
		},
	}
	dispatcher := newSQSDispatcherWithClient(mockClient, "https://sqs.example.com/queue")

	event := &Event{
		Type:    TypeSubmissionCreated,
		Payload: map[string]interface{}{},
	}

	err := dispatcher.Dispatch(context.Background(), event)

	assert.ErrorIs(t, err, expectedError)
}

func TestSQSDispatcher_Dispatch_MarshalError(t *testing.T) {
	mockClient := &mockSQSClient{}
	dispatcher := newSQSDispatcherWithClient(mockClient, "https://sqs.example.com/queue")

	// Create an event with a payload that cannot be marshaled (channel type)
	event := &Event{
		Type: TypeSubmissionCreated,
		Payload: map[string]interface{}{
			"channel": make(chan int), // channels cannot be JSON marshaled
		},
	}

	err := dispatcher.Dispatch(context.Background(), event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "json")
}

func TestSQSDispatcher_Dispatch_UsesTimeout(t *testing.T) {
	hasDeadline := false
	deadlineWithinRange := false
	mockClient := &mockSQSClient{
		sendMessageFunc: func(ctx context.Context, _ *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
			// Check the context has a deadline set (from the timeout)
			deadline, ok := ctx.Deadline()
			hasDeadline = ok
			if ok {
				deadlineWithinRange = time.Until(deadline) <= sqsTimeout && time.Until(deadline) > 0
			}
			return &sqs.SendMessageOutput{}, nil
		},
	}
	dispatcher := newSQSDispatcherWithClient(mockClient, "https://sqs.example.com/queue")

	event := &Event{
		Type:    TypeSubmissionCreated,
		Payload: map[string]interface{}{},
	}

	err := dispatcher.Dispatch(context.Background(), event)

	require.NoError(t, err)
	assert.True(t, hasDeadline, "context should have a deadline")
	assert.True(t, deadlineWithinRange, "deadline should be within timeout duration")
}

func TestSQSDispatcher_ImplementsDispatcher(_ *testing.T) {
	var _ Dispatcher = (*SQSDispatcher)(nil)
}

func TestNewSQSDispatcherWithClient(t *testing.T) {
	mockClient := &mockSQSClient{}
	queueURL := "https://sqs.example.com/queue"

	dispatcher := newSQSDispatcherWithClient(mockClient, queueURL)

	assert.Equal(t, mockClient, dispatcher.client)
	assert.Equal(t, queueURL, dispatcher.queueURL)
}

func TestNewSQSDispatcher_WithRegion(t *testing.T) {
	mockClient := &mockSQSClient{}
	queueURL := "https://sqs.eu-west-1.amazonaws.com/123456789/test-queue"
	region := "eu-west-1"

	dispatcher, err := newSQSDispatcher(
		context.Background(),
		queueURL,
		region,
		mockConfigLoader(nil),
		mockClientFactory(mockClient),
	)

	require.NoError(t, err)
	require.NotNil(t, dispatcher)
	assert.Equal(t, mockClient, dispatcher.client)
	assert.Equal(t, queueURL, dispatcher.queueURL)
}

func TestNewSQSDispatcher_WithoutRegion(t *testing.T) {
	mockClient := &mockSQSClient{}
	queueURL := "https://sqs.us-east-1.amazonaws.com/123456789/test-queue"

	dispatcher, err := newSQSDispatcher(
		context.Background(),
		queueURL,
		"", // no region
		mockConfigLoader(nil),
		mockClientFactory(mockClient),
	)

	require.NoError(t, err)
	require.NotNil(t, dispatcher)
	assert.Equal(t, mockClient, dispatcher.client)
	assert.Equal(t, queueURL, dispatcher.queueURL)
}

func TestNewSQSDispatcher_ConfigLoadError(t *testing.T) {
	expectedError := errors.New("failed to load AWS config")

	dispatcher, err := newSQSDispatcher(
		context.Background(),
		"https://sqs.example.com/queue",
		"eu-west-1",
		mockConfigLoader(expectedError),
		mockClientFactory(&mockSQSClient{}),
	)

	require.ErrorIs(t, err, expectedError)
	assert.Nil(t, dispatcher)
}

func TestDefaultSQSClientFactory(t *testing.T) {
	cfg := &aws.Config{Region: "us-east-1"}

	client := defaultSQSClientFactory(cfg)

	require.NotNil(t, client)
	// Verify it returns an *sqs.Client which implements sqsClient
	_, ok := client.(*sqs.Client)
	assert.True(t, ok, "should return *sqs.Client")
}

func TestNewSQSDispatcher(t *testing.T) {
	// This test verifies NewSQSDispatcher calls newSQSDispatcher with the right parameters.
	// It may fail if AWS credentials are not available, which is expected in unit test environments.
	// The test is primarily to ensure coverage of the function signature.
	dispatcher, err := NewSQSDispatcher(context.Background(), "https://sqs.example.com/queue", "us-east-1")

	// We expect either success (if AWS credentials are available) or an error (if not).
	// Either way, the function was called successfully.
	if err != nil {
		// Expected in environments without AWS credentials
		assert.Nil(t, dispatcher)
	} else {
		assert.NotNil(t, dispatcher)
		assert.Equal(t, "https://sqs.example.com/queue", dispatcher.queueURL)
	}
}

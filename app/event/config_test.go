package event

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDispatcherFromConfig_ReturnsNoopWhenEmpty(t *testing.T) {
	config := viper.New()
	config.Set("dispatcher", "")

	dispatcher, err := NewDispatcherFromConfig(context.Background(), config)

	require.NoError(t, err)
	assert.IsType(t, &NoopDispatcher{}, dispatcher)
}

func TestNewDispatcherFromConfig_ReturnsNoopWhenNotSet(t *testing.T) {
	config := viper.New()

	dispatcher, err := NewDispatcherFromConfig(context.Background(), config)

	require.NoError(t, err)
	assert.IsType(t, &NoopDispatcher{}, dispatcher)
}

func TestNewDispatcherFromConfig_ErrorOnUnknownType(t *testing.T) {
	config := viper.New()
	config.Set("dispatcher", "unknown")

	dispatcher, err := NewDispatcherFromConfig(context.Background(), config)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown event dispatcher type")
	assert.Nil(t, dispatcher)
}

func TestNewDispatcherFromConfig_ErrorOnMissingSQSQueueURL(t *testing.T) {
	config := viper.New()
	config.Set("dispatcher", "sqs")
	config.Set("sqs.queueURL", "")

	dispatcher, err := NewDispatcherFromConfig(context.Background(), config)

	require.ErrorIs(t, err, ErrSQSQueueURLRequired)
	assert.Nil(t, dispatcher)
}

func TestGetInstance(t *testing.T) {
	config := viper.New()
	config.Set("instance", "prod")

	instance := GetInstance(config)

	assert.Equal(t, "prod", instance)
}

func TestGetInstance_EmptyWhenNotSet(t *testing.T) {
	config := viper.New()

	instance := GetInstance(config)

	assert.Empty(t, instance)
}

func TestNewDispatcherFromConfig_ReturnsSQSDispatcher(t *testing.T) {
	viperConfig := viper.New()
	viperConfig.Set("dispatcher", "sqs")
	viperConfig.Set("sqs.queueURL", "https://sqs.example.com/queue")
	viperConfig.Set("sqs.region", "eu-west-1")

	mockSQSDispatcher := &SQSDispatcher{queueURL: "https://sqs.example.com/queue"}
	mockFactory := func(_ context.Context, queueURL, region string) (*SQSDispatcher, error) {
		assert.Equal(t, "https://sqs.example.com/queue", queueURL)
		assert.Equal(t, "eu-west-1", region)
		return mockSQSDispatcher, nil
	}

	dispatcher, err := newDispatcherFromConfig(context.Background(), viperConfig, mockFactory)

	require.NoError(t, err)
	assert.Equal(t, mockSQSDispatcher, dispatcher)
}

func TestNewDispatcherFromConfig_ReturnsSQSDispatcherWithoutRegion(t *testing.T) {
	viperConfig := viper.New()
	viperConfig.Set("dispatcher", "sqs")
	viperConfig.Set("sqs.queueURL", "https://sqs.example.com/queue")

	mockSQSDispatcher := &SQSDispatcher{queueURL: "https://sqs.example.com/queue"}
	mockFactory := func(_ context.Context, queueURL, region string) (*SQSDispatcher, error) {
		assert.Equal(t, "https://sqs.example.com/queue", queueURL)
		assert.Empty(t, region)
		return mockSQSDispatcher, nil
	}

	dispatcher, err := newDispatcherFromConfig(context.Background(), viperConfig, mockFactory)

	require.NoError(t, err)
	assert.Equal(t, mockSQSDispatcher, dispatcher)
}

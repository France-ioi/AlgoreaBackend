package event

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

const (
	dispatcherTypeSQS = "sqs"
)

// sqsDispatcherFactory is a function type for creating SQS dispatchers.
// This allows for dependency injection in tests.
type sqsDispatcherFactory func(ctx context.Context, queueURL, region string) (*SQSDispatcher, error)

// NewDispatcherFromConfig creates a dispatcher based on the configuration.
// If no dispatcher is configured (empty dispatcher type), returns a NoopDispatcher.
// Returns an error if the configuration is invalid.
func NewDispatcherFromConfig(ctx context.Context, viperConfig *viper.Viper) (Dispatcher, error) {
	return newDispatcherFromConfig(ctx, viperConfig, NewSQSDispatcher)
}

func newDispatcherFromConfig(
	ctx context.Context,
	viperConfig *viper.Viper,
	sqsFactory sqsDispatcherFactory,
) (Dispatcher, error) {
	dispatcherType := viperConfig.GetString("dispatcher")

	if dispatcherType == "" {
		return &NoopDispatcher{}, nil // No dispatcher configured
	}

	switch dispatcherType {
	case dispatcherTypeSQS:
		return newSQSDispatcherFromConfig(ctx, viperConfig, sqsFactory)
	default:
		return nil, fmt.Errorf("unknown event dispatcher type: %s", dispatcherType)
	}
}

// ErrSQSQueueURLRequired is returned when the SQS queue URL is not configured.
var ErrSQSQueueURLRequired = errors.New("event.sqs.queueURL is required when dispatcher is 'sqs'")

func newSQSDispatcherFromConfig(
	ctx context.Context,
	viperConfig *viper.Viper,
	sqsFactory sqsDispatcherFactory,
) (*SQSDispatcher, error) {
	queueURL := viperConfig.GetString("sqs.queueURL")
	if queueURL == "" {
		return nil, ErrSQSQueueURLRequired
	}

	region := viperConfig.GetString("sqs.region")

	return sqsFactory(ctx, queueURL, region)
}

// GetInstance returns the instance identifier from the configuration.
func GetInstance(config *viper.Viper) string {
	return config.GetString("instance")
}

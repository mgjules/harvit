package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/mgjules/minion/health"
	"github.com/mgjules/minion/logger"
)

// Publisher is a wrapper for amqp.Publisher.
type Publisher struct {
	*amqp.Publisher
	*gochannel.GoChannel

	logger *logger.Logger
	health *health.Checks
}

// NewPublisher returns a new publisher.
func NewPublisher(prod bool, amqpURI string, logger *logger.Logger, health *health.Checks) (*Publisher, error) {
	p := Publisher{
		logger: logger,
		health: health,
	}

	amqpConfig := amqp.NewDurablePubSubConfig(
		amqpURI,
		nil,
	)

	amqpConfig.Publish.ChannelPoolSize = 2

	publisher, err := amqp.NewPublisher(amqpConfig, watermill.NewStdLoggerWithOut(logger.Writer(), !prod, false))
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	p.Publisher = publisher
	p.health.RegisterChecks(p.Check())

	return &p, nil
}

// NewTestPublisher returns a new publisher for testing purposes.
func NewTestPublisher(logger *logger.Logger) *Publisher {
	return &Publisher{
		logger: logger,
		GoChannel: gochannel.NewGoChannel(
			gochannel.Config{},
			watermill.NewStdLoggerWithOut(logger.Writer(), true, false),
		),
	}
}

// Publish is a wrapper for the MessagePublishr.Publish.
func (p *Publisher) Publish(ctx context.Context, topic string, messages ...*message.Message) error {
	for _, m := range messages {
		m.SetContext(ctx)
	}

	return p.MessagePublisher().Publish(topic, messages...)
}

// MessagePublisher returns the message publisher.
func (p *Publisher) MessagePublisher() message.Publisher {
	if p.Publisher != nil {
		return p.Publisher
	}

	return p.GoChannel
}

// Check is used to perform healthcheck.
func (p *Publisher) Check() health.Check {
	//nolint:revive
	return health.Check{
		Name:          "messenger.publisher",
		RefreshPeriod: 10 * time.Second,
		InitialDelay:  10 * time.Second,
		Timeout:       5 * time.Second,
		Check: func(_ context.Context) error {
			if !p.IsConnected() {
				return errors.New("publisher is not connected")
			}

			return nil
		},
	}
}

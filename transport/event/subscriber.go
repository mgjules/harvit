package event

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/mgjules/minion/health"
	"github.com/mgjules/minion/logger"
)

// Subscriber is a wrapper for amqp.Subscriber.
type Subscriber struct {
	*amqp.Subscriber
	*gochannel.GoChannel

	logger *logger.Logger
	health *health.Checks
}

// NewSubscriber returns a new subscriber.
func NewSubscriber(
	prod bool,
	name,
	amqpURI string,
	logger *logger.Logger,
	health *health.Checks,
) (*Subscriber, error) {
	s := Subscriber{
		logger: logger,
		health: health,
	}

	amqpConfig := amqp.NewDurablePubSubConfig(
		amqpURI,
		func(topic string) string {
			return strings.ToUpper(name) + "_" + topic
		},
	)

	subscriber, err := amqp.NewSubscriber(amqpConfig, watermill.NewStdLoggerWithOut(logger.Writer(), !prod, false))
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	s.Subscriber = subscriber
	s.health.RegisterChecks(s.Check())

	return &s, nil
}

// NewTestSubscriber returns a new subscriber for testing purposes.
func NewTestSubscriber(logger *logger.Logger) *Subscriber {
	return &Subscriber{
		logger: logger,
		GoChannel: gochannel.NewGoChannel(
			gochannel.Config{},
			watermill.NewStdLoggerWithOut(logger.Writer(), true, false),
		),
	}
}

// Subscribe is a wrapper for the MessageSubscriber.Subscribe.
func (s *Subscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return s.MessageSubscriber().Subscribe(ctx, topic)
}

// MessageSubscriber returns the message subscriber.
func (s *Subscriber) MessageSubscriber() message.Subscriber {
	if s.Subscriber != nil {
		return s.Subscriber
	}

	return s.GoChannel
}

// Check is used to perform healthcheck.
func (s *Subscriber) Check() health.Check {
	//nolint:revive
	return health.Check{
		Name:          "messenger.subscriber",
		RefreshPeriod: 10 * time.Second,
		InitialDelay:  10 * time.Second,
		Timeout:       5 * time.Second,
		Check: func(_ context.Context) error {
			if !s.IsConnected() {
				return errors.New("subscriber is not connected")
			}

			return nil
		},
	}
}

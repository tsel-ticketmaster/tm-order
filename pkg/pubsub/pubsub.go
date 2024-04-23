package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ck "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

// MessageHeaders is type of message headers
type MessageHeaders map[string]string

// Add will add the key and value to headers.
func (mh MessageHeaders) Add(key, value string) {
	mh[key] = value
}

// DeadLetterQueueMessage is an entity.
type DeadLetterQueueMessage struct {
	Channel           string         `json:"channel"`
	Publisher         string         `json:"publisher"`
	Consumer          string         `json:"consumer"`
	Key               string         `json:"key"`
	Headers           MessageHeaders `json:"headers"`
	Message           string         `json:"message"`
	CausedBy          string         `json:"caused_by"`
	FailedConsumeDate string         `json:"failed_consume_date"`
}

// DLQHandlerAdapter is an dead letter queue adapter.
type DLQHandlerAdapter struct {
	topic     string
	publisher Publisher
}

// NewDLQHandlerAdapter is a constructor.
func NewDLQHandlerAdapter(topic string, publisher Publisher) *DLQHandlerAdapter {
	return &DLQHandlerAdapter{topic, publisher}
}

// Send will publish the dlq message to the assigned topic.
func (dlqHandlerAdapter *DLQHandlerAdapter) Send(ctx context.Context, dlqMessage *DeadLetterQueueMessage) (err error) {
	headers := MessageHeaders{}
	headers.Add("dlq", "true")

	key := fmt.Sprintf("%s:%s:%s:%d",
		dlqMessage.Consumer,
		dlqMessage.Channel,
		dlqMessage.Key,
		time.Now().UnixNano(),
	)

	messageByte, _ := json.Marshal(dlqMessage)
	topics := dlqHandlerAdapter.topic

	err = dlqHandlerAdapter.publisher.Publish(
		ctx,
		topics,
		key,
		headers,
		messageByte,
	)

	return
}

// DLQHandler is an handler to handler dead letter queue or an unprocessed message
type DLQHandler interface {
	Send(ctx context.Context, dlqMessage *DeadLetterQueueMessage) (err error)
}

// EventHandler is an event handler. It will be called after message is arrived to consumer
type EventHandler interface {
	Handle(ctx context.Context, message interface{}) (err error)
}

// Publisher is a collection of behavior of a publisher
type Publisher interface {
	// Will send the message to the assigned topic.
	Publish(ctx context.Context, topic string, key string, headers MessageHeaders, message []byte) (err error)
	Close() (err error)
}

// Subscriber is a collection of behavior of a subscriber
type Subscriber interface {
	Subscribe()
	Close() (err error)
}

type DefaultEventHandler struct {
	Logger *logrus.Logger
}

func (h DefaultEventHandler) Handle(ctx context.Context, message interface{}) (err error) {
	kafkaMessage, ok := message.(*ck.Message)
	if !ok {
		return fmt.Errorf("invalid message provider")
	}

	h.Logger.WithContext(ctx).WithFields(logrus.Fields{
		"kafka_message": string(kafkaMessage.Value),
	}).Info()

	return nil
}

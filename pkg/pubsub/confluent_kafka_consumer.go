package pubsub

import (
	"context"

	"go.opentelemetry.io/otel"

	ck "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	"github.com/sirupsen/logrus"
)

type ConfluentKafkaConsumer interface {
	Assign(partitions []ck.TopicPartition) (err error)
	Assignment() (partitions []ck.TopicPartition, err error)
	Unassign() (err error)
	SubscribeTopics(topics []string, rb ck.RebalanceCb) (err error)
	Poll(ms int) ck.Event
	Commit() (partitions []ck.TopicPartition, err error)
	Close() (err error)
}

type ConfluentKafkaConsumerProperty struct {
	Logger       *logrus.Logger
	Topic        string
	EventHandler EventHandler
	Consumer     ConfluentKafkaConsumer
}
type confluentKafkaConsumer struct {
	closeChan    chan struct{}
	logger       *logrus.Logger
	topic        string
	eventHandler EventHandler
	consumer     ConfluentKafkaConsumer
}

// Close implements Subscriber.
func (s confluentKafkaConsumer) Close() (err error) {
	close(s.closeChan)
	return s.consumer.Close()
}

// Subscribe implements Subscriber.
func (s confluentKafkaConsumer) Subscribe() {
	if err := s.consumer.SubscribeTopics([]string{s.topic}, nil); err != nil {
		s.logger.WithError(err).Error()
		return
	}

	go s.poll(100)
}

func (s confluentKafkaConsumer) poll(ms int) {
	for {
		select {
		case <-s.closeChan:
			return
		default:
			s.processEvent(s.consumer.Poll(ms))
		}
	}
}

func (s confluentKafkaConsumer) processEvent(event ck.Event) {
	if event == nil {
		return
	}

	switch e := event.(type) {
	case ck.RevokedPartitions:
		if err := s.consumer.Unassign(); err != nil {
			s.logger.WithError(err).Error()
		}
	case ck.AssignedPartitions:
		if err := s.consumer.Assign(e.Partitions); err != nil {
			s.logger.WithError(err).Error()
		}

	case *ck.Message:
		carrier := splunkkafka.NewMessageCarrier(e)
		propagator := otel.GetTextMapPropagator()

		ctx := propagator.Extract(context.Background(), carrier)

		s.eventHandler.Handle(ctx, e)
		s.consumer.Commit()
	case ck.Error:
		if e.Code() == ck.ErrAllBrokersDown {
			s.logger.WithError(e).WithFields(logrus.Fields{
				"code": e.Code().String(),
			}).Error()
		}
	default:

	}
}

func SubscriberFromConfluentKafkaConsumer(props ConfluentKafkaConsumerProperty) Subscriber {
	return confluentKafkaConsumer{
		closeChan:    make(chan struct{}, 1),
		logger:       props.Logger,
		topic:        props.Topic,
		eventHandler: props.EventHandler,
		consumer:     props.Consumer,
	}
}

package pubsub

import (
	"context"

	ck "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sirupsen/logrus"
)

type ConfluentKafkaProducer interface {
	Events() chan ck.Event
	Produce(msg *ck.Message, deliveryChan chan ck.Event) error
	Flush(timeoutMs int) int
	Close()
}

type confluentKafkaProducer struct {
	closeChan chan struct{}
	logger    *logrus.Logger
	producer  ConfluentKafkaProducer
}

func (p *confluentKafkaProducer) watchDeliveryReport() {
	for {

		select {
		case <-p.closeChan:
			return
		case event := <-p.producer.Events():
			p.processEventReport(event)
		}
	}
}

func (p *confluentKafkaProducer) processEventReport(event ck.Event) {
	switch e := event.(type) {
	case *ck.Message:
		m := e
		if m.TopicPartition.Error != nil {
			p.logger.WithError(m.TopicPartition.Error).Error()
		}

	default:
		p.logger.Infof("unexpected: %s", e.String())
	}

}

// Close implements Publisher.
func (p *confluentKafkaProducer) Close() (err error) {
	close(p.closeChan)
	p.producer.Flush(15 * 1000)
	p.producer.Close()
	return nil
}

// Publish implements Publisher.
func (p *confluentKafkaProducer) Publish(ctx context.Context, topic string, key string, headers MessageHeaders, message []byte) (err error) {
	kafkaMessageKey := []byte(key)
	kafkaMessageHeader := make([]ck.Header, 0)
	for k, v := range headers {
		kafkaMessageHeader = append(kafkaMessageHeader, ck.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	kafkaMessage := &ck.Message{
		TopicPartition: ck.TopicPartition{
			Topic:     &topic,
			Partition: ck.PartitionAny,
		},
		Value:   message,
		Key:     kafkaMessageKey,
		Headers: kafkaMessageHeader,
	}

	if err := p.producer.Produce(kafkaMessage, nil); err != nil {
		p.logger.WithContext(ctx).Error(err)
	}

	return nil
}

func PublisherFromConfluentKafkaProducer(logger *logrus.Logger, producer ConfluentKafkaProducer) Publisher {
	publisher := &confluentKafkaProducer{
		closeChan: make(chan struct{}, 1),
		logger:    logger,
		producer:  producer,
	}

	go publisher.watchDeliveryReport()

	return publisher
}

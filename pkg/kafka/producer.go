package kafka

import (
	"log"

	ck "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	"github.com/tsel-ticketmaster/tm-order/config"
)

func NewProducer() *splunkkafka.Producer {
	cfg := config.Get()

	cm := ck.ConfigMap{
		"bootstrap.servers": cfg.Kafka.Hosts,
		"security.protocol": cfg.Kafka.SecurityProtocol,
		"sasl.mechanisms":   cfg.Kafka.SASLMechanisms,
		"sasl.username":     cfg.Kafka.SASLUsername,
		"sasl.password":     cfg.Kafka.SASLPassword,
	}

	producer, err := splunkkafka.NewProducer(&cm)
	if err != nil {
		log.Println(err)
	}

	return producer
}

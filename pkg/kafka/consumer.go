package kafka

import (
	"log"

	ck "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/signalfx/splunk-otel-go/instrumentation/github.com/confluentinc/confluent-kafka-go/kafka/splunkkafka"
	"github.com/tsel-ticketmaster/tm-order/config"
)

func NewConsumer(groupID string, autoCommit bool) *splunkkafka.Consumer {
	cfg := config.Get()

	cm := ck.ConfigMap{
		"bootstrap.servers":               cfg.Kafka.Hosts,
		"group.id":                        groupID,
		"security.protocol":               cfg.Kafka.SecurityProtocol,
		"sasl.mechanisms":                 cfg.Kafka.SASLMechanisms,
		"sasl.username":                   cfg.Kafka.SASLUsername,
		"sasl.password":                   cfg.Kafka.SASLPassword,
		"session.timeout.ms":              cfg.Kafka.SessionTimeout,
		"enable.auto.commit":              autoCommit,
		"go.application.rebalance.enable": true,
		"partition.assignment.strategy":   "roundrobin",
		"auto.offset.reset":               "earliest",
	}

	consumer, err := splunkkafka.NewConsumer(&cm)
	if err != nil {
		log.Println(err)
	}

	return consumer
}

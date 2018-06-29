package cmd

import (
	"github.com/Shopify/sarama"
	"time"
)

const (
	maxProducerRetry = 5
)

//ProducerClient handles Kafka's producer operations
type ProducerClient interface {
	Publish(topic string, value string) error
}

//KafkaProducer implements production version of ProducerClient interface
type KafkaProducer struct {
	producer sarama.SyncProducer
}

//NewProducer creates a new SyncProducer using the given broker addresses
func NewProducer(hosts []string) (ProducerClient, error) {
	config := sarama.NewConfig()
	config.Producer.Retry.Max = maxProducerRetry
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_10_0_0

	producer, err := sarama.NewSyncProducer(hosts, config)

	return &KafkaProducer{producer: producer}, err
}

//Publish sends message to broker
func (p *KafkaProducer) Publish(topic string, value string) error {
	message := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(value),
		Timestamp: time.Now(),
	}

	_, _, err := p.producer.SendMessage(message)

	return err
}

package cmd

import (
	"fmt"
	"github.com/Shopify/sarama"
	"strings"
	"time"
)

const (
	maxProducerRetry = 5

	stagingSuffix    = "-staging"
	productionSuffix = "-production"
)

//ProducerProfile config type of producer's profile - staging/production
type ProducerProfile string

var (
	//StagingProfile is profile config value for staging
	StagingProfile = "staging"
	//ProductionProfile is profile config config value for production
	ProductionProfile = "production"
)

//ProducerClient handles Kafka's producer operations
type ProducerClient interface {
	Publish(topic string, value string) error
}

//KafkaProducer implements production version of ProducerClient interface
type KafkaProducer struct {
	producer sarama.SyncProducer
	suffix   string
}

//NewProducer creates a new SyncProducer using the given broker addresses
func NewProducer(hosts []string, profile string) (ProducerClient, error) {
	config := sarama.NewConfig()
	config.Producer.Retry.Max = maxProducerRetry
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Return.Successes = true
	config.Version = sarama.V0_10_0_0

	producer, err := sarama.NewSyncProducer(hosts, config)

	var suffix string
	if profile == ProductionProfile {
		suffix = productionSuffix
	} else if profile == StagingProfile {
		suffix = stagingSuffix
	} else {
		return nil, fmt.Errorf("Unrecognized producer profile: %s", profile)
	}

	return &KafkaProducer{producer: producer, suffix: suffix}, err
}

//Publish sends message to broker
func (p *KafkaProducer) Publish(topic string, value string) error {
	message := &sarama.ProducerMessage{
		Topic:     p.applySuffix(topic),
		Value:     sarama.StringEncoder(value),
		Timestamp: time.Now(),
	}

	_, _, err := p.producer.SendMessage(message)

	return err
}

//applySuffix is add string to topic for different approach in staging/production
func (p *KafkaProducer) applySuffix(topic string) string {
	var str strings.Builder
	str.WriteString(topic)
	str.WriteString(p.suffix)
	return str.String()
}

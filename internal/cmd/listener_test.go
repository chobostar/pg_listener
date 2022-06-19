package cmd

import (
	"errors"
	"github.com/chobostar/pg_listener/internal/model"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

//TestProducer implements test version of ProducerClient interface
type TestProducer struct {
	lastTopic string
	lastValue string
}

//Publish of TestProducer just stores last values
func (p *TestProducer) Publish(topic string, value string) error {
	p.lastTopic = topic
	p.lastValue = value
	return nil
}

//TestProducerRetries implements test version of ProducerClient interface
type TestProducerRetries struct {
	retryCount int
	lastTopic  string
	lastValue  string
}

//Publish of TestProducerRetries just stores last values
func (p *TestProducerRetries) Publish(topic string, value string) error {
	if p.retryCount < 1 {
		p.retryCount++
		return errors.New("mock error")
	}
	p.lastTopic = topic
	p.lastValue = value
	return nil
}

func TestProduceMessage(t *testing.T) {
	mockProducer := &TestProducer{}
	testMetrics := NewMetrics()

	testLogger, hook := test.NewNullLogger()
	testLogger.SetLevel(logrus.DebugLevel)

	mockListener := &PgListener{
		log:      testLogger,
		producer: mockProducer,
		metrics:  testMetrics,
	}

	var messages []model.Message
	messages = append(messages, model.Message{
		Topic: "test_topic",
		Value: "hello value",
	})

	mockListener.produceMessage(&messages)

	assert.Equal(t, "test_topic", mockProducer.lastTopic)
	assert.Equal(t, "hello value", mockProducer.lastValue)
	assert.Equal(t, "sending message {Topic:test_topic Value:hello value}", hook.LastEntry().Message)

	assert.Equal(t, uint64(1), testMetrics.totalMessages)
	assert.Equal(t, uint64(0), testMetrics.totalErrors)
}

func TestProduceMessageRetry(t *testing.T) {
	mockProducer := &TestProducerRetries{}
	testMetrics := NewMetrics()

	testLogger, hook := test.NewNullLogger()
	testLogger.SetLevel(logrus.DebugLevel)

	mockListener := &PgListener{
		log:      testLogger,
		producer: mockProducer,
		metrics:  testMetrics,
	}

	var messages []model.Message
	messages = append(messages, model.Message{
		Topic: "test_topic",
		Value: "hello value",
	})

	sleepSecondBetweenRetry = 0
	mockListener.produceMessage(&messages)

	assert.Equal(t, "test_topic", mockProducer.lastTopic)
	assert.Equal(t, "hello value", mockProducer.lastValue)
	assert.Equal(t, 1, mockProducer.retryCount)
	assert.Equal(t, "Error while producing message: mock error", hook.LastEntry().Message)

	assert.Equal(t, uint64(1), testMetrics.totalMessages)
	assert.Equal(t, uint64(1), testMetrics.totalErrors)
}

func TestGetAsMessages(t *testing.T) {
	testMetrics := NewMetrics()
	mockListener := &PgListener{metrics: testMetrics}

	var values []interface{}
	values = append(values, 3)
	values = append(values, "2022-03-23 05:44:45.360555")
	values = append(values, "demo_topic")
	values = append(values, "{\"id\": \"file\"}")

	wal2jsonMsg := model.Wal2JsonMessage{
		Change: []model.Wal2JsonChange{{
			Kind:         "insert",
			Schema:       "queue",
			Table:        "events",
			ColumnNames:  []string{"id", "added_at", "topic", "payload"},
			ColumnTypes:  []string{"bigint", "timestamp without time zone", "text", "json"},
			ColumnValues: values,
			OldKeys:      model.Wal2JsonOldKeys{},
		},
			{
				Kind:         "update", //wrong kind, suppose to be ignored
				Schema:       "queue",
				Table:        "events",
				ColumnNames:  []string{"id", "added_at", "topic", "payload"},
				ColumnTypes:  []string{"bigint", "timestamp without time zone", "text", "json"},
				ColumnValues: values,
				OldKeys:      model.Wal2JsonOldKeys{},
			},
		},
	}
	messages := mockListener.getAsMessages(wal2jsonMsg)

	assert.Equal(t, 1, len(*messages))
	assert.Equal(t, "demo_topic", (*messages)[0].Topic)
	assert.Equal(t, "{\"id\": \"file\"}", (*messages)[0].Value)
}

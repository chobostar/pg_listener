package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/chobostar/pg_listener/internal/model"
	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	defaultPostgresTimeline = -1
	columnNameForTopic      = "topic"
	columnNameForPayload    = "payload"
)

var (
	sleepSecondBetweenRetry = 3
)

//PgListener defines resources and config
type PgListener struct {
	log      logrus.FieldLogger
	producer ProducerClient

	dbHost   string
	dbPort   uint16
	dbName   string
	user     string
	password string

	slot       string
	lsn        string
	tableNames string
	chunks     string

	lastWal uint64

	buffer []byte
}

//InitListener initiates not started listener with given resources and configuration
func InitListener(log logrus.FieldLogger, producer ProducerClient, cfg *Config) *PgListener {
	port, _ := strconv.ParseUint(cfg.DbPort, 10, 16)

	pgListener := &PgListener{
		log:        log,
		producer:   producer,
		dbHost:     cfg.DbHost,
		dbName:     cfg.DbName,
		dbPort:     uint16(port),
		user:       cfg.DbUser,
		password:   cfg.DbPass,
		slot:       cfg.Slot,
		lsn:        cfg.Lsn,
		tableNames: cfg.TableNames,
		chunks:     cfg.Chunks,
		buffer:     []byte{},
	}

	if pgListener.slot == "" {
		log.Fatal("slot is not set")
	}
	if pgListener.lsn == "" {
		pgListener.lsn = "0/0"
	}

	return pgListener
}

//StartToListen connects to replication slot and launches WAL listener
func (l *PgListener) StartToListen() error {
	l.log.Info("pg_listener initiated, trying to start...")

	config := pgx.ConnConfig{
		Host:     l.dbHost,
		Port:     l.dbPort,
		Database: l.dbName,
		User:     l.user,
		Password: l.password,
	}
	conn, err := pgx.ReplicationConnect(config)
	if err != nil {
		return err
	}

	lsnAsInt, err := pgx.ParseLSN(l.lsn)
	if err != nil {
		return err
	}

	var outputParams []string
	if l.tableNames != "" {
		outputParams = append(outputParams, fmt.Sprintf("\"add-tables\" '%s'", l.tableNames))
	}
	if l.chunks != "" {
		outputParams = append(outputParams, fmt.Sprintf("\"write-in-chunks\" '%s'", l.chunks))
	}

	err = conn.StartReplication(l.slot, lsnAsInt, defaultPostgresTimeline, outputParams...)
	if err == nil {
		l.log.Info("now listening...")
		l.listen(conn)
	}

	return err
}

//listen is main infinite process for handle this replication slot
func (l *PgListener) listen(rc *pgx.ReplicationConn) {
	for {
		r, err := rc.WaitForReplicationMessage(context.Background())
		if err != nil {
			l.log.Fatal(err)
		}

		if r != nil {
			if r.ServerHeartbeat != nil {
				l.handleHeartbeat(rc, r.ServerHeartbeat.ServerWalEnd)
			} else if r.WalMessage != nil {
				l.handleMessage(rc, r.WalMessage.WalData, r.WalMessage.WalStart, r.WalMessage.ServerWalEnd)
			}
		}
	}
}

//handleHeartbeat sends standby status
func (l *PgListener) handleHeartbeat(rc *pgx.ReplicationConn, walEnd uint64) {
	l.log.Debug("server heartbeat received: ", walEnd)

	l.lastWal = walEnd

	if err := l.sendStandBy(rc); err != nil {
		l.log.Error(err)
	}
}

//handleMessage parses WAL data, send message to Kafka, and sends standby status
func (l *PgListener) handleMessage(rc *pgx.ReplicationConn, walData []byte, walStart uint64, walEnd uint64) {
	l.log.Debug("wal message received: ", walStart, " ", walEnd)

	l.lastWal = walStart

	//buffering all input bytes
	l.buffer = append(l.buffer, walData...)
	var change model.Wal2JsonMessage
	//trying to deserialize to JSON
	if err := json.Unmarshal(l.buffer, &change); err == nil {
		messages := l.getAsMessages(change)

		l.produceMessage(messages)

		l.buffer = []byte{}
	}

	if err := l.sendStandBy(rc); err != nil {
		l.log.Fatal(err)
	}
}

//produceMessage iterates over parsed messages and publish it to Kafka
func (l *PgListener) produceMessage(messages *[]model.Message) {
	if len(*messages) > 0 {
		for _, message := range *messages {
			l.log.Debug(fmt.Sprintf("sending message %+v", message))

			for {
				if err := l.producer.Publish(message.Topic, message.Value); err != nil {
					l.log.Error("Error while producing message: ", err)
					time.Sleep(time.Duration(sleepSecondBetweenRetry) * time.Second)
					continue
				}
				break
			}
		}
	}
}

func (l *PgListener) sendStandBy(rc *pgx.ReplicationConn) error {
	status, err := pgx.NewStandbyStatus(l.lastWal)
	if err != nil {
		return err
	}

	err = rc.SendStandbyStatus(status)
	return err
}

//getAsMessages parses topic and value from Wal2Json data
func (l *PgListener) getAsMessages(change model.Wal2JsonMessage) *[]model.Message {
	var messages []model.Message

	if len(change.Change) == 0 {
		return &messages
	}

	for _, item := range change.Change {
		//watch only inserts
		if item.Kind == "insert" {
			var topic, value string

			for i, name := range item.ColumnNames {
				if name == columnNameForTopic {
					topic = l.getParsedValue(item.ColumnValues[i])
				} else if name == columnNameForPayload {
					value = l.getParsedValue(item.ColumnValues[i])
				}
			}
			messages = append(messages, model.Message{Topic: topic, Value: value})
		}
	}

	return &messages
}

func (l *PgListener) getParsedValue(input interface{}) string {
	switch v := input.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case string:
		return v
	case nil:
		return "null"
	default:
		l.log.Fatal(fmt.Sprintf("Unknown type %T", v))
	}
	return ""
}

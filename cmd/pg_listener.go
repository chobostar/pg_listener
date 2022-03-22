package main

import (
	"flag"
	"github.com/chobostar/pg_listener/internal/cmd"
	"github.com/sirupsen/logrus"
	"strings"
)

//env PGLSN_DB_HOST=localhost
//PGLSN_DB_PORT=5432
//PGLSN_DB_USER=postgres
//PGLSN_DB_PASS=postgres
//PGLSN_DB_NAME=postgres
//PGLSN_TABLE_NAMES=queue.events
//PGLSN_SLOT=pg_listener
//PGLSN_KAFKA_HOSTS=localhost:9092
//PGLSN_CHUNKS=1
// go run cmd/pg_listener.go
func main() {
	isDebug := flag.Bool("debug", false, "debug mode for more log info")
	flag.Parse()

	log := logrus.New() //StdLogger
	if *isDebug {
		log.Info("Debug mode")
		log.SetLevel(logrus.DebugLevel)
	}

	cfg, err := cmd.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	producer, err := cmd.NewProducer(strings.Split(cfg.KafkaHosts, ","))
	if err != nil {
		log.Fatal(err)
	}

	listener := cmd.InitListener(log, producer, cfg)
	if err = listener.StartToListen(); err != nil {
		log.Fatal(err)
	}
}

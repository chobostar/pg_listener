# Go application to capture inserts to PostgreSQL table and produce events to Apache Kafka using replication slot and WAL

It connects to the replication slot and listens to INSERTs into the specified tables, from where it copies the value of the topic and payload columns, then sends this message to Kafka.
Listening table example:
```sql
CREATE TABLE queue.events (
    id bigserial primary key,
    added_at timestamp NOT NULL default clock_timestamp(),
    topic text NOT NULL,
    payload json
);
```

#### Installing
* it requires https://github.com/eulerto/wal2json 
* create slot in PostgreSQL ```SELECT pg_create_logical_replication_slot('my_slot','wal2json')```
* Go 1.13 or newer

```bash
make build
```

#### Example run
```bash
make up
make init
make build
make demo
```

make insert to postgres:
```bash
$ echo "insert into queue.events(topic, payload) values('demo_topic', '{\"id\": \"file\"}'::json);" | docker-compose exec -T postgres psql -U postgres

INSERT 0 1
```
see topics list:
```bash
$ docker-compose exec -T broker /bin/kafka-topics --bootstrap-server=localhost:9092 --list

demo_topic
```
read content:
```bash
$ docker-compose exec -T broker /bin/kafka-console-consumer --bootstrap-server=localhost:9092 --topic demo_topic --from-beginning 

{"id": "file"}
```

clean infra after all:
```bash
make down
```

#### Configure
PostgreSQL connection:
- PGLSN_DB_HOST
- PGLSN_DB_PORT
- PGLSN_DB_NAME
- PGLSN_DB_USER
- PGLSN_DB_PASS

[wal2json](https://github.com/eulerto/wal2json) options (see also https://github.com/eulerto/wal2json#parameters):
- PGLSN_TABLE_NAMES is `add-tables` - comma "," separated table names which INSERTs are produced to Kafka
- PGLSN_CHUNKS is `write-in-chunks`, if "1", write after every change instead of every changeset

Postgres [replication options](https://www.postgresql.org/docs/10/static/protocol-replication.html):
- PGLSN_SLOT - replication slot name where to connect
- PGLSN_LSN - Instructs server to start streaming WAL, starting at WAL location XXX/XXX. 0/0 by default.

Apache Kafka connection:
- PGLSN_KAFKA_HOSTS - comma separated hostname:port  

#### Exported metrics
it exports metrics on port `9938` of http path `/metrics`
- `pg_listener_last_committed_wal_location` - Last successfully commited to producer wal
- `pg_listener_logged_errors_total` - Total amount logged errors
- `pg_listener_processed_messages_total` - Total amount processed logical messages
- `pg_listener_raw_buffer_bytes` - Current amount of raw buffer size of logical message
- `pg_listener_received_heartbeats_total` - Total amount received heartbeats

#### Сaution
- If you don't use ```write-in-chunks``` option, then `ERROR: invalid memory alloc request size` issue is possible - https://github.com/eulerto/wal2json/issues/46
- Go bytes [slice is limited to 2Gb](https://go.dev/blog/slices-intro), so buffer of raw changes can be > 2^32

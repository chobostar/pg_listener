# Приложение для доставки сообщений из PostgreSQL в Apache Kafka через логическое декодирование
# Go application to capture inserts to PostgreSQL table and produce events to Apache Kafka using replication slot and WAL

Соединяется к слоту репликации и слушает INSERT-ы в указанные таблицы, откуда копирует значение колонок topic и payload, потом отправляет это сообщение в Kafka.
Пример таблицы для прослушивания:
```
CREATE TABLE queue.events (
    id bigserial primary key,
    added_at timestamp NOT NULL default clock_timestamp(),
    topic text NOT NULL,
    payload json
);
```

#### Установка
* требует установленный https://github.com/eulerto/wal2json на сервере PostgreSQL
* создать слот репликации на PostgreSQL ```SELECT pg_create_logical_replication_slot('my_slot','wal2json')```
* установить Go 1.9 или выше
* установить менеджер пакетов https://github.com/golang/dep (например: ```$ GOPATH="/home/user/go"```; ```$ go get -u github.com/golang/dep/cmd/dep```)
* ```git clone %этот_репозиторий%```
* ```cd pg_listener```
* ```env GOPATH="%полный_путь_до_папки_репозитория%" go build pg_listener```
* ИЛИ ```GOPATH="%полный_путь_до_папки_репозитория%" CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo pg_listener``` для сборки статического бинарника

Получаем исполняемый файл pg_listener. Запускать используя в качестве параметров переменные среды:
```
env PGLSN_DB_HOST=localhost PGLSN_DB_PORT=5432 PGLSN_DB_USER=postgres PGLSN_DB_PASS=postgres =demo PGLSN_TABLE_NAMES=queue.events PGLSN_SLOT=my_slot
PGLSN_KAFKA_HOSTS=kafka1.local:9092,kafka2.local:9092,kafka3.local:9092 PGLSN_CHUNKS=1 pg_listener
```

#### Конфигурация
Реквизиты PostgreSQL: <br/>
PGLSN_DB_HOST, PGLSN_DB_PORT, PGLSN_DB_NAME, PGLSN_DB_USER, PGLSN_DB_PASS<br/>

Опции wal2json (https://github.com/eulerto/wal2json):<br/>
PGLSN_TABLE_NAMES - содержимое опции 'add-tables' - наименование таблиц через "," INSERT-ы на которых, передаются в Kafka<br/>
PGLSN_CHUNKS - содержимое опции 'write-in-chunks', если "1", то пакетные изменения получает, как построчные<br/>

Опции репликации postgres (https://www.postgresql.org/docs/10/static/protocol-replication.html):<br/>
PGLSN_SLOT - слот репликации откуда получать изменения<br/>
PGLSN_LSN - в виде строки XXX/XXX, начиная с этого WAL начать репликацию. По-умолчанию пустая значение, которое соответствует 0/0.<br/>

Реквизиты Apache Kafka:<br/>
PGLSN_KAFKA_HOSTS - строки hostname:port через запятую слитно<br/>  


#### Примечание
Если не использовать опцию ```write-in-chunks``` есть риск, что большие изменения не смогут отреплицироваться, даже если они не относятся к отслеживаемой таблице (https://github.com/eulerto/wal2json/issues/46)


CREATE SCHEMA IF NOT EXISTS queue;

CREATE TABLE IF NOT EXISTS queue.events (
    id bigserial primary key,
    added_at timestamp NOT NULL default clock_timestamp(),
    topic text NOT NULL,
    payload json
);
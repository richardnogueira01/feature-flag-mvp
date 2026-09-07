CREATE TABLE revision_counter (
    id integer PRIMARY KEY CHECK (id = 1),
    revision bigint NOT NULL
);

INSERT INTO revision_counter (id, revision) VALUES (1, 0);

CREATE TABLE feature_flags (
    key text PRIMARY KEY CHECK (length(key) BETWEEN 1 AND 128),
    enabled boolean NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE flag_history (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    revision bigint NOT NULL,
    key text NOT NULL,
    operation text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX flag_history_key_revision_idx ON flag_history (key, revision);

CREATE TABLE outbox_events (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    revision bigint NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    attempts integer NOT NULL DEFAULT 0,
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX outbox_events_pending_idx ON outbox_events (id) WHERE published_at IS NULL;

-- Application -> Environment -> Configuration hierarchy (AWS AppConfig-like scope model).
-- Additive: does not drop existing data. Existing feature_flags rows are backfilled into an
-- implicit "legacy/default" configuration so the flat /v1/flags API keeps working unchanged.

CREATE TABLE IF NOT EXISTS applications (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name text NOT NULL CHECK (length(name) BETWEEN 1 AND 128),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(name)
);

CREATE TABLE IF NOT EXISTS environments (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  application_id bigint NOT NULL REFERENCES applications(id),
  name text NOT NULL CHECK (length(name) BETWEEN 1 AND 128),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(application_id, name)
);

CREATE TABLE IF NOT EXISTS configurations (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  environment_id bigint NOT NULL REFERENCES environments(id),
  name text NOT NULL CHECK (length(name) BETWEEN 1 AND 128),
  revision bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE(environment_id, name)
);

ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS id bigint GENERATED ALWAYS AS IDENTITY;
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS configuration_id bigint;
ALTER TABLE flag_history ADD COLUMN IF NOT EXISTS configuration_id bigint;
ALTER TABLE outbox_events ADD COLUMN IF NOT EXISTS configuration_id bigint;

-- Backfill: create the legacy/default scope once and point every pre-existing flag at it.
DO $$
DECLARE
  legacy_app_id bigint;
  legacy_env_id bigint;
  legacy_cfg_id bigint;
BEGIN
  INSERT INTO applications(name) VALUES ('legacy')
    ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO legacy_app_id;

  INSERT INTO environments(application_id, name) VALUES (legacy_app_id, 'default')
    ON CONFLICT (application_id, name) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO legacy_env_id;

  INSERT INTO configurations(environment_id, name) VALUES (legacy_env_id, 'default')
    ON CONFLICT (environment_id, name) DO UPDATE SET name = EXCLUDED.name
    RETURNING id INTO legacy_cfg_id;

  UPDATE feature_flags SET configuration_id = legacy_cfg_id WHERE configuration_id IS NULL;
END $$;

-- Move the primary key from the globally-unique "key" column to the surrogate "id" column, so
-- the same key can exist in different configurations. Guarded so re-running the migration is a
-- no-op once the swap already happened.
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint c
    JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey)
    WHERE c.conrelid = 'feature_flags'::regclass AND c.contype = 'p' AND a.attname = 'id'
  ) THEN
    ALTER TABLE feature_flags DROP CONSTRAINT IF EXISTS feature_flags_pkey;
    ALTER TABLE feature_flags ADD PRIMARY KEY (id);
  END IF;
END $$;

ALTER TABLE feature_flags ALTER COLUMN configuration_id SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'feature_flags_configuration_fk'
  ) THEN
    ALTER TABLE feature_flags
      ADD CONSTRAINT feature_flags_configuration_fk
      FOREIGN KEY (configuration_id) REFERENCES configurations(id);
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS feature_flags_configuration_key_idx ON feature_flags(configuration_id, key);
CREATE INDEX IF NOT EXISTS flag_history_configuration_idx ON flag_history(configuration_id, revision);

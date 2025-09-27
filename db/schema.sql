BEGIN;

CREATE TABLE IF NOT EXISTS system_metrics (
    id          BIGSERIAL PRIMARY KEY,
    type        TEXT        NOT NULL,
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    values      JSONB       NOT NULL,
    CHECK (jsonb_typeof(values) = 'object')
);

CREATE INDEX IF NOT EXISTS system_metrics_type_timestamp_idx
    ON system_metrics (type, timestamp DESC);

CREATE OR REPLACE FUNCTION notify_system_metrics()
RETURNS TRIGGER AS
$$
DECLARE
    payload JSON;
BEGIN
    payload := json_build_object(
        'timestamp', NEW.timestamp,
        'type', NEW.type,
        'values', NEW.values
    );

    PERFORM pg_notify('metrics', payload::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS system_metrics_notify ON system_metrics;

CREATE TRIGGER system_metrics_notify
AFTER INSERT ON system_metrics
FOR EACH ROW
EXECUTE FUNCTION notify_system_metrics();

COMMIT;

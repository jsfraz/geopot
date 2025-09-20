-- Drop any existing primary key constraint (if needed during migrations)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'connections_pkey') THEN
        ALTER TABLE connections DROP CONSTRAINT connections_pkey;
    END IF;
END$$;

-- Create hypertable for TimescaleDB
SELECT create_hypertable('connections', 'timestamp', if_not_exists => TRUE, migrate_data => TRUE);

-- Create indexes for optimal querying
CREATE INDEX IF NOT EXISTS idx_connections_timestamp_ip ON connections (timestamp, ip_address);
CREATE INDEX IF NOT EXISTS idx_connections_timestamp_country ON connections (timestamp, country_code);
CREATE INDEX IF NOT EXISTS idx_connections_timestamp_user ON connections (timestamp, "user");
CREATE INDEX IF NOT EXISTS idx_connections_ip_timestamp ON connections (ip_address, timestamp DESC);

-- Add compression policy (compress chunks older than 7 days)
ALTER TABLE connections SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'ip_address',
    timescaledb.compress_orderby = 'timestamp'
);

SELECT add_compression_policy('connections', INTERVAL '7 days', if_not_exists => TRUE);
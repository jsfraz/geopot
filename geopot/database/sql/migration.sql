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
CREATE INDEX IF NOT EXISTS idx_connections_timestamp_country_name ON connections (timestamp, country_name);
CREATE INDEX IF NOT EXISTS idx_connections_timestamp_user ON connections (timestamp, "user");
CREATE INDEX IF NOT EXISTS idx_connections_ip_timestamp ON connections (ip_address, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_connections_lat_lng_round ON connections (ROUND(latitude::numeric, 1), ROUND(longitude::numeric, 1));

-- Add compression policy (compress chunks older than 7 days)
ALTER TABLE connections SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'ip_address',
    timescaledb.compress_orderby = 'timestamp'
);

SELECT add_compression_policy('connections', INTERVAL '7 days', if_not_exists => TRUE);

-- Continuous Aggregates for Heatmap (hourly)
CREATE MATERIALIZED VIEW IF NOT EXISTS heatmap_1h
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', "timestamp") AS bucket,
    ROUND(latitude::numeric, 1) AS latitude,
    ROUND(longitude::numeric, 1) AS longitude,
    COUNT(*) AS intensity
FROM connections
WHERE latitude != 0 AND longitude != 0
GROUP BY bucket, latitude, longitude
WITH NO DATA;

SELECT add_continuous_aggregate_policy('heatmap_1h',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

-- Continuous Aggregates for Stats (hourly)
CREATE MATERIALIZED VIEW IF NOT EXISTS stats_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', "timestamp") AS bucket,
    COUNT(*) AS connection_count
FROM connections
GROUP BY bucket
WITH NO DATA;

SELECT add_continuous_aggregate_policy('stats_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

-- Continuous Aggregates for Top Countries (hourly)
CREATE MATERIALIZED VIEW IF NOT EXISTS stats_country_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', "timestamp") AS bucket,
    country_name,
    COUNT(*) AS connection_count
FROM connections
WHERE country_name != ''
GROUP BY bucket, country_name
WITH NO DATA;

SELECT add_continuous_aggregate_policy('stats_country_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

-- Continuous Aggregates for Top Usernames (hourly)
CREATE MATERIALIZED VIEW IF NOT EXISTS stats_user_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', "timestamp") AS bucket,
    "user",
    COUNT(*) AS connection_count
FROM connections
WHERE "user" != ''
GROUP BY bucket, "user"
WITH NO DATA;

SELECT add_continuous_aggregate_policy('stats_user_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);

-- Continuous Aggregates for Top Passwords (hourly)
CREATE MATERIALIZED VIEW IF NOT EXISTS stats_password_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', "timestamp") AS bucket,
    password,
    COUNT(*) AS connection_count
FROM connections
WHERE password != ''
GROUP BY bucket, password
WITH NO DATA;

SELECT add_continuous_aggregate_policy('stats_password_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE);
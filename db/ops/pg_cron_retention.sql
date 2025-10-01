CREATE EXTENSION IF NOT EXISTS pg_cron;

SELECT cron.schedule('cron-retention', '0 1 * * *', $$DELETE FROM system_metrics WHERE timestamp < NOW() - INTERVAL '7 days'$$);
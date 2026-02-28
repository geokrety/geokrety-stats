-- No-op rollback by design.
-- Dropping potentially hundreds of historical/future partitions would be destructive.
SELECT 1;

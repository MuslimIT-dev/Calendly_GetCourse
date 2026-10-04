DROP TRIGGER IF EXISTS trg_payments_updated_at ON payments;

DROP INDEX IF EXISTS idx_payments_status_time;
DROP INDEX IF EXISTS idx_payments_paid_time;
DROP INDEX IF EXISTS idx_payments_pending_expires;
DROP INDEX IF EXISTS idx_payments_reference;
DROP INDEX IF EXISTS idx_payments_user_time;
DROP INDEX IF EXISTS uniq_payments_provider_external;

DROP TABLE IF EXISTS payments CASCADE;
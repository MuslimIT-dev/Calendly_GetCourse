DROP TRIGGER IF EXISTS trg_appointments_updated_at ON appointments;

DROP INDEX IF EXISTS idx_appointments_pending_payment;
DROP INDEX IF EXISTS idx_appointments_active;
DROP INDEX IF EXISTS idx_appointments_service;
DROP INDEX IF EXISTS idx_appointments_client_time;
DROP INDEX IF EXISTS idx_appointments_master_location_time;
DROP INDEX IF EXISTS idx_appointments_master_time;
DROP INDEX IF EXISTS uniq_appointments_slot;

DROP TABLE IF EXISTS appointments CASCADE;
DROP TRIGGER IF EXISTS trg_exceptions_updated_at ON exceptions;
DROP TRIGGER IF EXISTS trg_schedule_rules_updated_at ON schedule_rules;

DROP INDEX IF EXISTS idx_exceptions_date;
DROP INDEX IF EXISTS idx_exceptions_location_date;
DROP INDEX IF EXISTS idx_schedule_rules_location_working;
DROP INDEX IF EXISTS idx_schedule_rules_location;

DROP TABLE IF EXISTS exceptions CASCADE;
DROP TABLE IF EXISTS schedule_rules CASCADE;
DROP TRIGGER IF EXISTS trg_masters_updated_at ON masters;
DROP TRIGGER IF EXISTS trg_locations_updated_at ON locations;
DROP TRIGGER IF EXISTS trg_languages_updated_at ON languages;
DROP TRIGGER IF EXISTS trg_certificates_updated_at ON certificates;

DROP INDEX IF EXISTS idx_masters_user_id;
DROP INDEX IF EXISTS idx_masters_catalog_rating;
DROP INDEX IF EXISTS idx_masters_catalog_experience;
DROP INDEX IF EXISTS idx_masters_specialization_rating;
DROP INDEX IF EXISTS idx_masters_experience;
DROP INDEX IF EXISTS idx_users_timezone;
DROP INDEX IF EXISTS idx_locations_master_active;

DROP INDEX IF EXISTS idx_languages_language_master;
DROP INDEX IF EXISTS idx_certificates_master_id;
DROP INDEX IF EXISTS idx_masters_catalog_price;

DROP TABLE IF EXISTS certificates CASCADE;
DROP TABLE IF EXISTS languages CASCADE;
DROP TABLE IF EXISTS locations CASCADE;
DROP TABLE IF EXISTS masters CASCADE;

DROP TYPE IF EXISTS proficiency_level;
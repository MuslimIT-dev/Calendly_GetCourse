DROP TRIGGER IF EXISTS trg_service_locations_updated_at ON service_locations;
DROP TRIGGER IF EXISTS trg_services_updated_at ON services;
DROP TRIGGER IF EXISTS trg_default_services_updated_at ON default_services;

DROP INDEX IF EXISTS idx_services_master_active;
DROP INDEX IF EXISTS idx_services_id_active;
DROP INDEX IF EXISTS idx_services_price_currency;
DROP INDEX IF EXISTS idx_service_locations_service;
DROP INDEX IF EXISTS idx_service_locations_location;

DROP TABLE IF EXISTS service_locations CASCADE;
DROP TABLE IF EXISTS services CASCADE;
DROP TABLE IF EXISTS default_services CASCADE;
DROP TABLE IF EXISTS categories CASCADE;
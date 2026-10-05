CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE default_services (
    id SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price INT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    duration_mins INT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE services (
    id SERIAL PRIMARY KEY,
    master_id INT NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    default_service_id INT REFERENCES default_services(id) ON DELETE SET NULL,
    slug VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price INT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    duration_mins INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE service_locations (
    service_id INT NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    location_id INT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (service_id, location_id)
);

CREATE INDEX idx_services_master_active
ON services(master_id)
WHERE is_active = TRUE;

CREATE INDEX idx_services_id_active
ON services(id)
WHERE is_active = TRUE;

CREATE INDEX idx_services_price_currency
ON services(price, currency)
WHERE is_active = TRUE;

CREATE INDEX idx_service_locations_service
ON service_locations(service_id, location_id);

CREATE INDEX idx_service_locations_location
ON service_locations(location_id)
WHERE is_active = TRUE;

CREATE TRIGGER trg_default_services_updated_at
BEFORE UPDATE ON default_services
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_services_updated_at
BEFORE UPDATE ON services
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

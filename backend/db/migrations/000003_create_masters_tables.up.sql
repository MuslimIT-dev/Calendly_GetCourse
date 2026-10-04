CREATE TYPE proficiency_level AS ENUM ('beginner', 'intermediate', 'advanced');

CREATE TABLE masters (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slug VARCHAR(255) UNIQUE NOT NULL,
    bio TEXT,
    specialization VARCHAR(255) NOT NULL,
    years_of_experience INT NOT NULL,
    is_accepting_bookings BOOLEAN DEFAULT FALSE,
    default_location_id INT,
    min_price INT DEFAULT 0,
    avg_rating DECIMAL(3, 2) DEFAULT 0.0,
    reviews_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE locations (
    id SERIAL PRIMARY KEY,
    master_id INT NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    timezone VARCHAR(255) NOT NULL,
    is_online BOOLEAN DEFAULT TRUE,
    meeting_url VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE masters
ADD CONSTRAINT fk_masters_default_location
FOREIGN KEY (default_location_id) REFERENCES locations(id) ON DELETE SET NULL;

CREATE TABLE languages (
    id SERIAL PRIMARY KEY,
    master_id INT NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    language VARCHAR(255) NOT NULL,
    proficiency proficiency_level NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE certificates (
    id SERIAL PRIMARY KEY,
    master_id INT NOT NULL REFERENCES masters(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    organization VARCHAR(255) NOT NULL,
    year INT NOT NULL,
    file_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_masters_user_id ON masters(user_id);

CREATE INDEX idx_masters_catalog_rating
ON masters(avg_rating DESC, id)
WHERE is_accepting_bookings = TRUE;

CREATE INDEX idx_masters_catalog_experience
ON masters(years_of_experience DESC, id)
WHERE is_accepting_bookings = TRUE;

CREATE INDEX idx_masters_specialization_rating
ON masters(specialization, avg_rating DESC)
WHERE is_accepting_bookings = TRUE;

CREATE INDEX idx_masters_experience
ON masters(years_of_experience)
WHERE is_accepting_bookings = TRUE;

CREATE INDEX idx_masters_timezone
ON masters(timezone)
WHERE is_accepting_bookings = TRUE;

CREATE INDEX idx_locations_master_active
ON locations(master_id)
WHERE is_active = TRUE;

CREATE INDEX idx_languages_language_master
ON languages(language, master_id);

CREATE INDEX idx_certificates_master_id
ON certificates(master_id);

CREATE INDEX idx_masters_catalog_price
ON masters(min_price ASC, id)
WHERE is_accepting_bookings = TRUE;

CREATE TRIGGER trg_masters_updated_at
BEFORE UPDATE ON masters
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_locations_updated_at
BEFORE UPDATE ON locations
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_languages_updated_at
BEFORE UPDATE ON languages
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_certificates_updated_at
BEFORE UPDATE ON certificates
FOR EACH ROW EXECUTE FUNCTION update_updated_at();
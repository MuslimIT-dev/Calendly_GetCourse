CREATE TABLE appointments (
    id            BIGSERIAL PRIMARY KEY,
    client_id     INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_by    INT REFERENCES users(id) ON DELETE SET NULL,
    master_id     INT NOT NULL REFERENCES masters(id) ON DELETE RESTRICT,
    service_id    INT NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    location_id   INT NOT NULL REFERENCES locations(id) ON DELETE RESTRICT,

    start_time    TIMESTAMPTZ NOT NULL,
    end_time      TIMESTAMPTZ NOT NULL,
    timezone      VARCHAR(64) NOT NULL DEFAULT 'UTC',

    status        SMALLINT NOT NULL DEFAULT 1
                  CHECK (status IN (1, 2, 3, 4, 5, 6)),
    -- 1=PENDING, 2=CONFIRMED, 3=CANCELLED_BY_CLIENT,
    -- 4=CANCELLED_BY_MASTER, 5=COMPLETED, 6=NO_SHOW

    price          INT NOT NULL CHECK (price >= 0),
    currency       VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_status SMALLINT NOT NULL DEFAULT 1
                   CHECK (payment_status IN (1, 2, 3, 4)),
    -- 1=PENDING, 2=PAID, 3=REFUNDED, 4=FAILED

    is_online    BOOLEAN NOT NULL DEFAULT FALSE,
    meeting_url  VARCHAR(512),

    client_notes TEXT,
    master_notes TEXT,

    cancelled_at TIMESTAMPTZ,
    cancelled_by INT REFERENCES users(id) ON DELETE SET NULL,
    cancel_reason TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CHECK (end_time > start_time),
    CHECK (is_online = FALSE OR meeting_url IS NOT NULL)
);

CREATE UNIQUE INDEX uniq_appointments_slot
ON appointments(master_id, location_id, start_time)
WHERE status IN (1, 2);

CREATE INDEX idx_appointments_master_time
ON appointments(master_id, start_time DESC);

CREATE INDEX idx_appointments_master_location_time
ON appointments(master_id, location_id, start_time DESC);

CREATE INDEX idx_appointments_client_time
ON appointments(client_id, start_time DESC);

CREATE INDEX idx_appointments_service
ON appointments(service_id, start_time DESC);

CREATE INDEX idx_appointments_active
ON appointments(start_time)
WHERE status IN (1, 2);

CREATE INDEX idx_appointments_pending_payment
ON appointments(start_time)
WHERE payment_status = 1 AND status = 1;

CREATE TRIGGER trg_appointments_updated_at
BEFORE UPDATE ON appointments
FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TABLE payments (
    id              BIGSERIAL PRIMARY KEY,
    user_id         INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    amount          INT NOT NULL CHECK (amount > 0),
    currency        CHAR(3) NOT NULL DEFAULT 'RUB',

    status          SMALLINT NOT NULL DEFAULT 1
                    CHECK (status IN (1, 2, 3, 4, 5)),
    -- 1=PENDING, 2=PAID, 3=FAILED, 4=REFUNDED, 5=EXPIRED

    provider        SMALLINT NOT NULL
                    CHECK (provider IN (1, 2, 3, 4)),
    -- 1=YOOKASSA, 2=STRIPE, 3=TINKOFF, 4=MANUAL

    purpose         SMALLINT NOT NULL
                    CHECK (purpose IN (1, 2, 3)),
    -- 1=BOOKING, 2=COURSE, 3=SUBSCRIPTION

    reference_id    VARCHAR(64) NOT NULL,  -- booking_id / course_id
    external_id     VARCHAR(255),          -- ID in pay system
    payment_url     VARCHAR(1024),

    failure_reason  TEXT,
    receipt_url     VARCHAR(1024),

    paid_at         TIMESTAMPTZ,
    refunded_at     TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CHECK (status <> 2 OR paid_at IS NOT NULL),
    CHECK (status <> 4 OR refunded_at IS NOT NULL)
);

CREATE UNIQUE INDEX uniq_payments_provider_external
ON payments(provider, external_id)
WHERE external_id IS NOT NULL;

CREATE INDEX idx_payments_user_time
ON payments(user_id, created_at DESC);

CREATE INDEX idx_payments_reference
ON payments(purpose, reference_id);

CREATE INDEX idx_payments_pending_expires
ON payments(expires_at)
WHERE status = 1;

CREATE INDEX idx_payments_paid_time
ON payments(paid_at DESC)
WHERE status = 2;

CREATE INDEX idx_payments_status_time
ON payments(status, created_at DESC);

CREATE TRIGGER trg_payments_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TABLE reviews (
    id            BIGSERIAL PRIMARY KEY,
    author_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id     INT NOT NULL,
    target        SMALLINT NOT NULL CHECK (target IN (1, 2)),
    reference_id  VARCHAR(64) NOT NULL,
    rating        SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment       TEXT,
    reply         TEXT,
    replied_at    TIMESTAMPTZ,
    status        SMALLINT NOT NULL DEFAULT 1 CHECK (status IN (1, 2, 3)),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (author_id, reference_id)
);

CREATE INDEX idx_reviews_target
ON reviews(target, target_id, created_at DESC)
WHERE status = 1;

CREATE INDEX idx_reviews_author
ON reviews(author_id, created_at DESC);

CREATE INDEX idx_reviews_reference
ON reviews(reference_id);

CREATE TRIGGER trg_reviews_updated_at
BEFORE UPDATE ON reviews
FOR EACH ROW EXECUTE FUNCTION update_updated_at();
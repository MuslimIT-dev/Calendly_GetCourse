DROP TRIGGER IF EXISTS trg_reviews_updated_at ON reviews;
DROP INDEX IF EXISTS idx_reviews_reference;
DROP INDEX IF EXISTS idx_reviews_author;
DROP INDEX IF EXISTS idx_reviews_target;
DROP TABLE IF EXISTS reviews CASCADE;
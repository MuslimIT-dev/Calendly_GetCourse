CREATE TABLE schedule_rules (
    id                        SERIAL PRIMARY KEY,
    location_id               INT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    day_of_week               SMALLINT NOT NULL CHECK (day_of_week BETWEEN 1 AND 7),  -- 1=Mon, 7=Sun
    is_working_day            BOOLEAN NOT NULL DEFAULT TRUE,
    work_start                TIME,
    work_end                  TIME,
    rest_start                TIME,
    rest_end                  TIME,
    buffer_between_slots_mins INT NOT NULL DEFAULT 0 CHECK (buffer_between_slots_mins >= 0),
    slot_step_mins            INT NOT NULL DEFAULT 0 CHECK (slot_step_mins >= 0),
    created_at                TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (location_id, day_of_week),
    CHECK (is_working_day = FALSE OR (work_start IS NOT NULL AND work_end IS NOT NULL)),
    CHECK (work_start IS NULL OR work_end IS NULL OR work_start < work_end),
    CHECK (rest_start IS NULL OR rest_end IS NULL OR rest_start < rest_end)
);

CREATE INDEX idx_schedule_rules_location
ON schedule_rules(location_id);

CREATE INDEX idx_schedule_rules_location_working
ON schedule_rules(location_id, day_of_week)
WHERE is_working_day = TRUE;

CREATE TRIGGER trg_schedule_rules_updated_at
BEFORE UPDATE ON schedule_rules
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TABLE exceptions (
    id           SERIAL PRIMARY KEY,
    location_id  INT NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    date         DATE NOT NULL,
    type         SMALLINT NOT NULL CHECK (type IN (1, 2, 3)),  -- 1=DAY_OFF, 2=CUSTOM_HOURS, 3=VACATION
    is_full_day  BOOLEAN NOT NULL DEFAULT TRUE,
    custom_start TIME,
    custom_end   TIME,
    rest_start   TIME,
    rest_end     TIME,
    reason       TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (is_full_day = TRUE OR (custom_start IS NOT NULL AND custom_end IS NOT NULL)),
    CHECK (custom_start IS NULL OR custom_end IS NULL OR custom_start < custom_end)
);

CREATE INDEX idx_exceptions_location_date
ON exceptions(location_id, date);

CREATE INDEX idx_exceptions_date
ON exceptions(date);

CREATE TRIGGER trg_exceptions_updated_at
BEFORE UPDATE ON exceptions
FOR EACH ROW EXECUTE FUNCTION update_updated_at();
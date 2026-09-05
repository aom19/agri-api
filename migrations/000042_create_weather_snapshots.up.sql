CREATE TABLE IF NOT EXISTS weather_snapshots (
    id                  BIGSERIAL PRIMARY KEY,
    field_id            UUID REFERENCES fields(id) ON DELETE CASCADE,
    latitude            DOUBLE PRECISION NOT NULL,
    longitude           DOUBLE PRECISION NOT NULL,
    observed_at         TIMESTAMPTZ NOT NULL,
    temperature_c       DOUBLE PRECISION NOT NULL,
    condition           VARCHAR(120) NOT NULL DEFAULT '',
    weather_code        INT,
    humidity_percent    INT,
    wind_speed_kmh      DOUBLE PRECISION,
    wind_direction_deg  INT,
    precipitation_mm    DOUBLE PRECISION,
    cloud_cover_percent INT,
    source              VARCHAR(40) NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (field_id, observed_at)
);

CREATE INDEX IF NOT EXISTS idx_weather_snapshots_field_observed
    ON weather_snapshots (field_id, observed_at DESC);
CREATE INDEX IF NOT EXISTS idx_weather_snapshots_observed
    ON weather_snapshots (observed_at);

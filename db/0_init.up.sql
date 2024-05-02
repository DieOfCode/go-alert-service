CREATE TYPE metric_type AS ENUM ('counter', 'gauge');
CREATE TABLE metrics (
    id VARCHAR NOT NULL,
    type metric_type NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    UNIQUE (id)
);

CREATE TABLE metrics (
    id VARCHAR NOT NULL,
    type VARCHAR NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    UNIQUE (id, type)
);

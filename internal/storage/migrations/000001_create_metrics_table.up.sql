CREATE TABLE IF NOT EXISTS metrics (
    id    text PRIMARY KEY,
    mtype text NOT NULL,
    delta bigint,
    value double precision
);

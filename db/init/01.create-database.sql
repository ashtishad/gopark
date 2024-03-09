CREATE DATABASE gopark;
\c gopark
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS parking_lots
(
    id             SERIAL PRIMARY KEY,
    uuid UUID DEFAULT uuid_generate_v4(),
    name           VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS slots
(
    id             SERIAL PRIMARY KEY,
    uuid        UUID    DEFAULT uuid_generate_v4(),
    parking_lot_id INTEGER NOT NULL REFERENCES parking_lots (id),
    slot_number    INTEGER NOT NULL,
    is_available   BOOLEAN DEFAULT TRUE,
    is_maintenance BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS vehicles
(
    id                  SERIAL PRIMARY KEY,
    uuid          UUID DEFAULT uuid_generate_v4(),
    registration_number VARCHAR(255) NOT NULL,
    slot_id             INTEGER      NOT NULL REFERENCES slots (id),
    parked_at           TIMESTAMPTZ    NOT NULL,
    unparked_at         TIMESTAMPTZ
);

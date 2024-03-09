CREATE DATABASE gopark;
\c gopark
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users
(
    id      SERIAL PRIMARY KEY,
    user_id UUID DEFAULT uuid_generate_v4(),
    name    VARCHAR(255) NOT NULL,
    role    VARCHAR(255) NOT NULL
        CHECK (role IN ('general', 'manager'))
);

CREATE TABLE IF NOT EXISTS parking_lots
(
    id             SERIAL PRIMARY KEY,
    parking_lot_id UUID DEFAULT uuid_generate_v4(),
    name           VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS slots
(
    id             SERIAL PRIMARY KEY,
    slot_id        UUID    DEFAULT uuid_generate_v4(),
    parking_lot_id INTEGER NOT NULL REFERENCES parking_lots (id),
    slot_number    INTEGER NOT NULL,
    distance       INTEGER NOT NULL,
    is_available   BOOLEAN DEFAULT TRUE,
    is_maintenance BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS vehicles
(
    id                  SERIAL PRIMARY KEY,
    vehicle_id          UUID DEFAULT uuid_generate_v4(),
    registration_number VARCHAR(255) NOT NULL,
    slot_id             INTEGER      NOT NULL REFERENCES slots (id),
    parked_at           TIMESTAMP    NOT NULL,
    unparked_at         TIMESTAMP,
    user_id             INTEGER      NOT NULL REFERENCES users (id)
);

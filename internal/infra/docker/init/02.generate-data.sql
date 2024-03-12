\c gopark

-- Table: parking_lots, create 1
INSERT INTO parking_lots (name) VALUES ('Parking Lot 1');


-- Table: slots, slots per parking lot with slot_number 1-10
INSERT INTO slots (uuid, parking_lot_id, slot_number)
SELECT uuid_generate_v4(),
       id,
       generate_series(1, 10)
FROM parking_lots;

-- Table: vehicles, insert 10 vehicles to the each slots order by slot number to find the nearest one(slot_number 1 is nearer than slot_number 2)
INSERT INTO vehicles (uuid, registration_number, slot_id, parked_at)
VALUES
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 1,  now() - interval '6 hours'),
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 2,  now() - interval '6 hours' - interval '10 minutes'),
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 3,  now() - interval '6 hours' - interval '1 hour 20 minutes'),
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 4,  now() - interval '6 hours' - interval '2 hour 30 minutes'),
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 5,  now() - interval '6 hours' - interval '2 hour 40 minutes'),
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 6,  now() - interval '6 hours' - interval '2 hour 50 minutes'),
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 7,  now() - interval '6 hours' - interval '3 hour'),
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 8,  now() - interval '6 hours' - interval '4 hour 10 minutes'),
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 9,  now() - interval '6 hours' - interval '5 hour 20 minutes'),
    (uuid_generate_v4(), 'ABC-' || substring(md5(random()::text), 0, 6), 10, now() - interval '6 hours' - interval '5 hour 30 minutes')
;

UPDATE vehicles v
SET unparked_at = now()
WHERE v.slot_id IN (1, 3, 5, 7, 9);

UPDATE slots s
SET is_available = false
WHERE s.id IN (2, 4, 6, 8, 10);

-- Update sequences
SELECT setval('parking_lots_id_seq', 2, false);
SELECT setval('slots_id_seq', 11, false);
SELECT setval('vehicles_id_seq', 11, false);

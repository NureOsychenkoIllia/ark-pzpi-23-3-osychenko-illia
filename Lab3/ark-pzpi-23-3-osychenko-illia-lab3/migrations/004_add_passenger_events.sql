-- Додаємо тестові події пасажирів для рейсу ID=1
-- Рейс відправився 2025-12-15 06:05:00, прибув 2025-12-15 12:15:00
-- Додаємо 435 пасажирів (entry події) протягом поїздки

-- Генеруємо події входу пасажирів
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced)
SELECT 
    1 as trip_id,
    'entry' as event_type,
    '2025-12-15 06:05:00'::timestamptz + (generate_series * interval '1 minute') as timestamp,
    49.9935 + (random() - 0.5) * 0.01 as latitude,  -- Харків координати з варіацією
    36.2304 + (random() - 0.5) * 0.01 as longitude,
    generate_series as passenger_count_after,
    generate_series as device_local_id,
    true as is_synced
FROM generate_series(1, 100);

-- Додаємо кілька подій виходу в кінці поїздки (частина пасажирів виходить)
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced)
SELECT 
    1 as trip_id,
    'exit' as event_type,
    '2025-12-15 12:10:00'::timestamptz + (generate_series * interval '30 seconds') as timestamp,
    50.4501 + (random() - 0.5) * 0.01 as latitude,  -- Київ координати з варіацією
    30.5234 + (random() - 0.5) * 0.01 as longitude,
    435 - generate_series as passenger_count_after,
    435 - generate_series + 1 as device_local_id,
    true as is_synced
FROM generate_series(1, 35);

-- Додаємо події для рейсу ID=20 (completed)
-- Цей рейс мав 0 current_passengers, але додамо реалістичні дані
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced)
SELECT 
    20 as trip_id,
    'entry' as event_type,
    '2025-12-21 18:08:33'::timestamptz + (generate_series * interval '2 minutes') as timestamp,
    49.9935 + (random() - 0.5) * 0.01 as latitude,
    36.2304 + (random() - 0.5) * 0.01 as longitude,
    generate_series as passenger_count_after,
    1000 + generate_series as device_local_id,
    true as is_synced
FROM generate_series(1, 28);

-- Всі пасажири виходять в кінці рейсу ID=20
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced)
SELECT 
    20 as trip_id,
    'exit' as event_type,
    '2025-12-21 21:05:00'::timestamptz + (generate_series * interval '1 minute') as timestamp,
    50.4501 + (random() - 0.5) * 0.01 as latitude,
    30.5234 + (random() - 0.5) * 0.01 as longitude,
    28 - generate_series as passenger_count_after,
    1000 + generate_series as device_local_id,
    true as is_synced
FROM generate_series(1, 28);

-- Додаємо події для рейсу ID=17 (in_progress)
-- Цей рейс має 42 current_passengers
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced)
SELECT 
    17 as trip_id,
    'entry' as event_type,
    '2025-12-22 19:08:33'::timestamptz + (generate_series * interval '3 minutes') as timestamp,
    49.9935 + (random() - 0.5) * 0.01 as latitude,
    36.2304 + (random() - 0.5) * 0.01 as longitude,
    generate_series as passenger_count_after,
    2000 + generate_series as device_local_id,
    true as is_synced
FROM generate_series(1, 42);

-- Додаємо події для рейсу ID=2 (in_progress)
-- Цей рейс має 42 current_passengers
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced)
SELECT 
    2 as trip_id,
    'entry' as event_type,
    '2025-12-15 12:02:00'::timestamptz + (generate_series * interval '2 minutes') as timestamp,
    49.9935 + (random() - 0.5) * 0.01 as latitude,
    36.2304 + (random() - 0.5) * 0.01 as longitude,
    generate_series as passenger_count_after,
    3000 + generate_series as device_local_id,
    true as is_synced
FROM generate_series(1, 42);
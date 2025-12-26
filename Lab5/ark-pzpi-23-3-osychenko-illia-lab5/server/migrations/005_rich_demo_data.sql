-- Багаті демонстраційні дані для дешборду
-- Цей файл створює реалістичні дані для всіх компонентів системи

-- Очищуємо існуючі дані (опціонально)
-- TRUNCATE TABLE passenger_events, trip_analytics, demand_forecasts, audit_logs, trips RESTART IDENTITY CASCADE;

-- Додаємо більше рейсів з різними статусами та часовими періодами
INSERT INTO trips (route_id, bus_id, scheduled_departure, actual_departure, actual_arrival, status, current_passengers, driver_name) VALUES
-- Завершені рейси за останній тиждень
(1, 1, NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days', NOW() - INTERVAL '7 days' + INTERVAL '6 hours', 'completed', 0, 'Іван Петренко'),
(2, 2, NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days', NOW() - INTERVAL '6 days' + INTERVAL '4 hours', 'completed', 0, 'Олег Коваленко'),
(3, 3, NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days' + INTERVAL '8 hours', 'completed', 0, 'Микола Сидоренко'),
(1, 4, NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days' + INTERVAL '5 hours', 'completed', 0, 'Андрій Мельник'),
(2, 1, NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days' + INTERVAL '4 hours', 'completed', 0, 'Іван Петренко'),
(3, 2, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '7 hours', 'completed', 0, 'Олег Коваленко'),
(1, 3, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '6 hours', 'completed', 0, 'Микола Сидоренко'),

-- Поточні активні рейси
(2, 4, NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours', NULL, 'in_progress', 35, 'Андрій Мельник'),
(3, 1, NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour', NULL, 'in_progress', 28, 'Іван Петренко'),
(1, 2, NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes', NULL, 'boarding', 15, 'Олег Коваленко'),

-- Заплановані рейси
(2, 3, NOW() + INTERVAL '1 hour', NULL, NULL, 'scheduled', 0, 'Микола Сидоренко'),
(3, 4, NOW() + INTERVAL '2 hours', NULL, NULL, 'scheduled', 0, 'Андрій Мельник'),
(1, 1, NOW() + INTERVAL '3 hours', NULL, NULL, 'scheduled', 0, 'Іван Петренко');

-- Додаємо аналітику рентабельності для всіх завершених рейсів (тільки нові)
INSERT INTO trip_analytics (trip_id, total_passengers, max_passengers, avg_occupancy_rate, revenue, fuel_cost, driver_cost, other_costs, profit, profitability_percent, calculated_at)
SELECT 
    t.id,
    CASE 
        WHEN t.route_id = 1 THEN 40 + (random() * 15)::int
        WHEN t.route_id = 2 THEN 25 + (random() * 20)::int
        ELSE 30 + (random() * 25)::int
    END as total_passengers,
    CASE 
        WHEN t.route_id = 1 THEN 50
        WHEN t.route_id = 2 THEN 45
        ELSE 55
    END as max_passengers,
    CASE 
        WHEN t.route_id = 1 THEN 80.0 + (random() * 20)
        WHEN t.route_id = 2 THEN 55.0 + (random() * 35)
        ELSE 54.0 + (random() * 40)
    END as avg_occupancy_rate,
    CASE 
        WHEN t.route_id = 1 THEN 7200.00 + (random() * 2000)
        WHEN t.route_id = 2 THEN 3600.00 + (random() * 1500)
        ELSE 5400.00 + (random() * 2500)
    END as revenue,
    1000.00 + (random() * 500) as fuel_cost,
    700.00 + (random() * 200) as driver_cost,
    200.00 + (random() * 150) as other_costs,
    0 as profit, -- Буде обчислено нижче
    0 as profitability_percent, -- Буде обчислено нижче
    t.actual_arrival + INTERVAL '1 hour' as calculated_at
FROM trips t 
WHERE t.status = 'completed' 
AND NOT EXISTS (SELECT 1 FROM trip_analytics ta WHERE ta.trip_id = t.id);

-- Оновлюємо profit та profitability_percent
UPDATE trip_analytics 
SET profit = revenue - fuel_cost - driver_cost - other_costs,
    profitability_percent = ROUND(((revenue - fuel_cost - driver_cost - other_costs) / (fuel_cost + driver_cost + other_costs) * 100)::numeric, 1);

-- Додаємо прогнози попиту на наступні 14 днів
INSERT INTO demand_forecasts (route_id, forecast_date, day_of_week, predicted_passengers, confidence_lower, confidence_upper, created_at)
SELECT 
    r.id as route_id,
    (NOW() + (s.day_offset || ' days')::interval)::date as forecast_date,
    EXTRACT(dow FROM NOW() + (s.day_offset || ' days')::interval)::int as day_of_week,
    CASE 
        WHEN r.id = 1 THEN 35 + (random() * 20)::int
        WHEN r.id = 2 THEN 25 + (random() * 15)::int
        ELSE 30 + (random() * 18)::int
    END as predicted_passengers,
    CASE 
        WHEN r.id = 1 THEN 28 + (random() * 10)::int
        WHEN r.id = 2 THEN 18 + (random() * 8)::int
        ELSE 22 + (random() * 12)::int
    END as confidence_lower,
    CASE 
        WHEN r.id = 1 THEN 45 + (random() * 10)::int
        WHEN r.id = 2 THEN 35 + (random() * 8)::int
        ELSE 40 + (random() * 12)::int
    END as confidence_upper,
    NOW() - INTERVAL '1 hour' as created_at
FROM routes r
CROSS JOIN generate_series(1, 14) s(day_offset)
WHERE NOT EXISTS (
    SELECT 1 FROM demand_forecasts df 
    WHERE df.route_id = r.id 
    AND df.forecast_date = (NOW() + (s.day_offset || ' days')::interval)::date
);

-- Додаємо події пасажирів для активних рейсів
-- Для рейсу в процесі (35 пасажирів)
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced)
SELECT 
    (SELECT id FROM trips WHERE status = 'in_progress' AND current_passengers = 35 LIMIT 1) as trip_id,
    'entry' as event_type,
    NOW() - INTERVAL '2 hours' + (generate_series * interval '2 minutes') as timestamp,
    49.9935 + (random() - 0.5) * 0.01 as latitude,
    36.2304 + (random() - 0.5) * 0.01 as longitude,
    generate_series as passenger_count_after,
    4000 + generate_series as device_local_id,
    true as is_synced
FROM generate_series(1, 35);

-- Для другого рейсу в процесі (28 пасажирів)
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced)
SELECT 
    (SELECT id FROM trips WHERE status = 'in_progress' AND current_passengers = 28 LIMIT 1) as trip_id,
    'entry' as event_type,
    NOW() - INTERVAL '1 hour' + (generate_series * interval '1.5 minutes') as timestamp,
    49.9935 + (random() - 0.5) * 0.01 as latitude,
    36.2304 + (random() - 0.5) * 0.01 as longitude,
    generate_series as passenger_count_after,
    5000 + generate_series as device_local_id,
    true as is_synced
FROM generate_series(1, 28);

-- Для рейсу на посадці (15 пасажирів)
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced)
SELECT 
    (SELECT id FROM trips WHERE status = 'boarding' LIMIT 1) as trip_id,
    'entry' as event_type,
    NOW() - INTERVAL '30 minutes' + (generate_series * interval '1 minute') as timestamp,
    49.9935 + (random() - 0.5) * 0.01 as latitude,
    36.2304 + (random() - 0.5) * 0.01 as longitude,
    generate_series as passenger_count_after,
    6000 + generate_series as device_local_id,
    true as is_synced
FROM generate_series(1, 15);

-- Додаємо багато записів в журнал аудиту
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_values, new_values, ip_address, created_at) VALUES
-- Останні дії адміністратора
(1, 'UPDATE', 'routes', 1, '{"base_price": 180.00}', '{"base_price": 200.00}', '192.168.1.50', NOW() - INTERVAL '3 hours'),
(1, 'UPDATE', 'routes', 2, '{"base_price": 120.00}', '{"base_price": 140.00}', '192.168.1.50', NOW() - INTERVAL '2 hours 30 minutes'),
(1, 'CREATE', 'trips', (SELECT MAX(id) FROM trips), '{}', '{"route_id": 1, "bus_id": 2}', '192.168.1.50', NOW() - INTERVAL '2 hours'),
(1, 'UPDATE', 'system_settings', 1, '{"peak_hours_coefficient": 1.20}', '{"peak_hours_coefficient": 1.30}', '192.168.1.50', NOW() - INTERVAL '1 hour 45 minutes'),

-- Дії диспетчера
(2, 'UPDATE', 'trips', (SELECT id FROM trips WHERE status = 'in_progress' LIMIT 1), '{"status": "boarding"}', '{"status": "in_progress"}', '192.168.1.100', NOW() - INTERVAL '1 hour 30 minutes'),
(2, 'CREATE', 'trips', (SELECT MAX(id)-1 FROM trips), '{}', '{"route_id": 2, "bus_id": 3}', '192.168.1.100', NOW() - INTERVAL '1 hour'),
(2, 'UPDATE', 'buses', 1, '{"status": "maintenance"}', '{"status": "active"}', '192.168.1.100', NOW() - INTERVAL '45 minutes'),
(2, 'UPDATE', 'trips', (SELECT id FROM trips WHERE status = 'boarding' LIMIT 1), '{"current_passengers": 10}', '{"current_passengers": 15}', '192.168.1.100', NOW() - INTERVAL '30 minutes'),

-- Дії водіїв (через мобільний додаток)
(3, 'UPDATE', 'trips', (SELECT id FROM trips WHERE status = 'in_progress' AND current_passengers = 35 LIMIT 1), '{"current_passengers": 32}', '{"current_passengers": 35}', '10.0.2.15', NOW() - INTERVAL '25 minutes'),
(4, 'UPDATE', 'trips', (SELECT id FROM trips WHERE status = 'in_progress' AND current_passengers = 28 LIMIT 1), '{"current_passengers": 25}', '{"current_passengers": 28}', '10.0.2.16', NOW() - INTERVAL '20 minutes'),
(3, 'CREATE', 'passenger_events', (SELECT MAX(id) FROM passenger_events), '{}', '{"trip_id": 1, "event_type": "entry"}', '10.0.2.15', NOW() - INTERVAL '15 minutes'),
(4, 'CREATE', 'passenger_events', (SELECT MAX(id)-1 FROM passenger_events), '{}', '{"trip_id": 2, "event_type": "entry"}', '10.0.2.16', NOW() - INTERVAL '10 minutes');

-- Оновлюємо статистику для кращої продуктивності
ANALYZE trips;
ANALYZE trip_analytics;
ANALYZE demand_forecasts;
ANALYZE audit_logs;
ANALYZE passenger_events;

-- Додаємо коментар про успішне завантаження
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_values, new_values, ip_address, created_at) VALUES
(1, 'SYSTEM', 'database', 0, '{}', '{"message": "Rich demo data loaded successfully"}', '127.0.0.1', NOW());
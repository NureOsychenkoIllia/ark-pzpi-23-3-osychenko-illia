-- Початкові дані для системи

-- Дозволи
INSERT INTO permissions (name, description) VALUES
    ('routes:read', 'Перегляд маршрутів'),
    ('routes:write', 'Створення та редагування маршрутів'),
    ('buses:read', 'Перегляд автобусів'),
    ('buses:write', 'Управління автобусами'),
    ('users:read', 'Перегляд користувачів'),
    ('users:write', 'Управління користувачами'),
    ('analytics:read', 'Перегляд аналітики'),
    ('reports:export', 'Експорт звітів'),
    ('audit:read', 'Перегляд журналу дій'),
    ('system:backup', 'Резервне копіювання');

-- Ролі
INSERT INTO roles (name, description) VALUES
    ('dispatcher', 'Диспетчер - моніторинг та звіти'),
    ('admin', 'Бізнес-адміністратор - повний доступ'),
    ('tech_admin', 'Технічний адміністратор - системні функції');

-- Призначення дозволів ролям
-- Диспетчер
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'dispatcher' AND p.name IN 
    ('routes:read', 'buses:read', 'analytics:read', 'reports:export');

-- Бізнес-адміністратор
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN 
    ('routes:read', 'routes:write', 'buses:read', 'buses:write', 
     'users:read', 'users:write', 'analytics:read', 'reports:export', 'audit:read');

-- Технічний адміністратор
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'tech_admin' AND p.name IN 
    ('audit:read', 'system:backup');

-- Тестові користувачі (пароль: password123)
INSERT INTO users (email, password_hash, full_name, role_id) VALUES
    ('dispatcher@busoptima.ua', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Іван Петренко', 
     (SELECT id FROM roles WHERE name = 'dispatcher')),
    ('admin@busoptima.ua', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Марія Іваненко', 
     (SELECT id FROM roles WHERE name = 'admin')),
    ('tech@busoptima.ua', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Олексій Сидоренко', 
     (SELECT id FROM roles WHERE name = 'tech_admin'));

-- Тестові маршрути
INSERT INTO routes (origin_city, destination_city, distance_km, base_price, fuel_cost_per_km, driver_cost_per_trip, estimated_duration_minutes, is_active) VALUES
    ('Харків', 'Київ', 480.0, 150.00, 4.50, 800.00, 360, true),
    ('Київ', 'Одеса', 475.0, 180.00, 4.50, 900.00, 420, true),
    ('Харків', 'Дніпро', 215.0, 80.00, 4.50, 600.00, 180, false),
    ('Львів', 'Київ', 540.0, 200.00, 4.50, 1000.00, 480, true);

-- Тестові автобуси
INSERT INTO buses (registration_number, capacity, model, fuel_consumption_per_100km, is_active) VALUES
    ('AA1234BB', 50, 'Mercedes Sprinter', 12.5, true),
    ('BB5678CC', 35, 'Iveco Daily', 11.8, false),
    ('CC9012DD', 45, 'Volkswagen Crafter', 13.2,  true),
    ('DD3456EE', 55, 'MAN TGE', 14.1, true);

-- Тестові IoT-пристрої
INSERT INTO devices (serial_number, auth_token_hash, bus_id, firmware_version) VALUES
    ('IOT001', '$2a$10$UaADGyMmroMpTOL6/YG.XergCJWC8MgjfLJILsAT0xQXhDgKKcKNK', 1, '1.2.3'),
    ('IOT002', '$2a$10$UaADGyMmroMpTOL6/YG.XergCJWC8MgjfLJILsAT0xQXhDgKKcKNK', 2, '1.2.3'),
    ('IOT003', '$2a$10$UaADGyMmroMpTOL6/YG.XergCJWC8MgjfLJILsAT0xQXhDgKKcKNK', 3, '1.2.3'),
    ('IOT004', '$2a$10$UaADGyMmroMpTOL6/YG.XergCJWC8MgjfLJILsAT0xQXhDgKKcKNK', 4, '1.2.3');

-- Тестові рейси
INSERT INTO trips (route_id, bus_id, scheduled_departure, actual_departure, actual_arrival, status, current_passengers, driver_name) VALUES
    (1, 1, '2025-12-15 08:00:00+02', '2025-12-15 08:05:00+02', '2025-12-15 14:15:00+02', 'completed', 0, 'Петро Коваленко'),
    (1, 2, '2025-12-15 14:00:00+02', '2025-12-15 14:02:00+02', NULL, 'in_progress', 28, 'Андрій Мельник'),
    (2, 3, '2025-12-15 10:00:00+02', '2025-12-15 10:10:00+02', '2025-12-15 17:05:00+02', 'completed', 0, 'Василь Шевченко'),
    (3, 4, '2025-12-15 16:00:00+02', NULL, NULL, 'scheduled', 0, 'Олег Бондаренко'),
    (1, 1, '2025-12-16 08:00:00+02', NULL, NULL, 'scheduled', 0, 'Петро Коваленко'),
    (2, 2, '2025-12-16 09:30:00+02', NULL, NULL, 'scheduled', 0, 'Андрій Мельник');

-- Тестові події пасажирів
INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, longitude, passenger_count_after, device_local_id, is_synced) VALUES
    -- Рейс 1 (завершений)
    (1, 'entry', '2025-12-15 08:10:00+02', 49.9935, 36.2304, 15, 101, true),
    (1, 'entry', '2025-12-15 08:25:00+02', 49.9800, 36.2500, 28, 102, true),
    (1, 'exit', '2025-12-15 09:45:00+02', 50.1500, 36.8000, 22, 103, true),
    (1, 'entry', '2025-12-15 10:15:00+02', 50.2000, 37.0000, 35, 104, true),
    (1, 'exit', '2025-12-15 12:30:00+02', 50.4000, 30.5000, 18, 105, true),
    (1, 'exit', '2025-12-15 14:10:00+02', 50.4501, 30.5234, 0, 106, true),
    
    -- Рейс 2 (в процесі)
    (2, 'entry', '2025-12-15 14:05:00+02', 49.9935, 36.2304, 12, 201, true),
    (2, 'entry', '2025-12-15 14:20:00+02', 49.9800, 36.2500, 25, 202, true),
    (2, 'entry', '2025-12-15 14:35:00+02', 49.9600, 36.3000, 28, 203, true),
    
    -- Рейс 3 (завершений)
    (3, 'entry', '2025-12-15 10:15:00+02', 50.4501, 30.5234, 20, 301, true),
    (3, 'entry', '2025-12-15 10:30:00+02', 50.4000, 30.6000, 32, 302, true),
    (3, 'exit', '2025-12-15 13:45:00+02', 46.4825, 30.7233, 25, 303, true),
    (3, 'exit', '2025-12-15 16:50:00+02', 46.4775, 30.7326, 8, 304, true),
    (3, 'exit', '2025-12-15 17:00:00+02', 46.4693, 30.7408, 0, 305, true);

-- Рекомендації цін
INSERT INTO price_recommendations (trip_id, base_price, recommended_price, occupancy_rate, demand_coefficient, time_coefficient, day_coefficient) VALUES
    (1, 150.00, 165.00, 0.70, 1.1, 1.0, 1.0),
    (2, 150.00, 180.00, 0.80, 1.2, 1.0, 1.0),
    (3, 180.00, 195.00, 0.73, 1.08, 1.0, 1.0),
    (4, 80.00, 85.00, 0.00, 1.06, 1.0, 1.0),
    (5, 150.00, 160.00, 0.00, 1.07, 1.0, 1.0),
    (6, 180.00, 190.00, 0.00, 1.06, 1.0, 1.0);

-- Аналітика рейсів
INSERT INTO trip_analytics (trip_id, total_passengers, max_passengers, avg_occupancy_rate, revenue, fuel_cost, driver_cost, other_costs, profit, profitability_percent) VALUES
    (1, 142, 35, 0.70, 3550.00, 1200.25, 800.00, 150.00, 1399.75, 39.43),
    (3, 98, 32, 0.64, 2940.00, 1087.50, 900.00, 120.00, 832.50, 28.32);

-- Прогнози попиту
INSERT INTO demand_forecasts (route_id, forecast_date, day_of_week, predicted_passengers, confidence_lower, confidence_upper, actual_passengers) VALUES
    (1, '2025-12-16', 1, 145, 130, 160, NULL),
    (1, '2025-12-17', 2, 138, 125, 151, NULL),
    (1, '2025-12-18', 3, 142, 128, 156, NULL),
    (2, '2025-12-16', 1, 95, 85, 105, NULL),
    (2, '2025-12-17', 2, 88, 80, 96, NULL),
    (3, '2025-12-16', 1, 78, 70, 86, NULL),
    (4, '2025-12-16', 1, 125, 115, 135, NULL);

-- Сповіщення
INSERT INTO notifications (user_id, trip_id, type, severity, message, is_read) VALUES
    (1, 2, 'delay', 'warning', 'Рейс Харків-Київ затримується на 5 хвилин', false),
    (1, 4, 'capacity', 'info', 'Рейс Харків-Дніпро заповнений на 85%', true),
    (2, 1, 'completion', 'info', 'Рейс Харків-Київ успішно завершено', true),
    (2, NULL, 'system', 'warning', 'Необхідно оновити прошивку пристрою IOT003', false);

-- Журнал аудиту
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_values, new_values, ip_address) VALUES
    (2, 'CREATE', 'trip', 5, '{}', '{"route_id": 1, "bus_id": 1, "driver_name": "Петро Коваленко"}', '192.168.1.100'),
    (2, 'UPDATE', 'trip', 2, '{"status": "scheduled"}', '{"status": "in_progress"}', '192.168.1.100'),
    (1, 'READ', 'analytics', NULL, '{}', '{}', '192.168.1.101'),
    (2, 'CREATE', 'user', 4, '{}', '{"email": "test@busoptima.ua", "role": "dispatcher"}', '192.168.1.100'),
    (3, 'BACKUP', 'system', NULL, '{}', '{"backup_file": "backup_20251215.sql"}', '192.168.1.102');
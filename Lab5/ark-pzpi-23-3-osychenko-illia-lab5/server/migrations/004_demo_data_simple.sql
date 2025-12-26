-- Спрощені демонстраційні дані для Lab3

-- Додаємо активні рейси для демонстрації моніторингу
INSERT INTO trips (route_id, bus_id, scheduled_departure, status, current_passengers, driver_name) VALUES
(1, 1, NOW() + INTERVAL '1 hour', 'in_progress', 42, 'Іван Петренко'),
(2, 2, NOW() + INTERVAL '30 minutes', 'boarding', 22, 'Олег Коваленко'),
(3, 3, NOW() + INTERVAL '2 hours', 'scheduled', 0, 'Микола Сидоренко');

-- Додаємо завершені рейси для аналітики
INSERT INTO trips (route_id, bus_id, scheduled_departure, actual_departure, actual_arrival, status, current_passengers, driver_name) VALUES
(1, 1, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '3 hours', 'completed', 0, 'Іван Петренко'),
(2, 3, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '2 hours', 'completed', 0, 'Микола Сидоренко'),
(3, 4, NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '4 hours', 'completed', 0, 'Андрій Мельник');

-- Додаємо аналітику рентабельності для завершених рейсів
INSERT INTO trip_analytics (trip_id, total_passengers, max_passengers, avg_occupancy_rate, revenue, fuel_cost, driver_cost, other_costs, profit, profitability_percent, calculated_at) VALUES
-- Високорентабельний рейс
((SELECT id FROM trips WHERE route_id = 1 AND status = 'completed' LIMIT 1), 45, 50, 90.0, 8100.00, 1200.00, 800.00, 300.00, 5800.00, 252.2, NOW() - INTERVAL '12 hours'),
-- Нормальний рейс
((SELECT id FROM trips WHERE route_id = 2 AND status = 'completed' LIMIT 1), 28, 35, 80.0, 3360.00, 800.00, 600.00, 200.00, 1760.00, 110.0, NOW() - INTERVAL '12 hours'),
-- Низькорентабельний рейс
((SELECT id FROM trips WHERE route_id = 3 AND status = 'completed' LIMIT 1), 15, 45, 33.3, 3000.00, 1400.00, 900.00, 350.00, 350.00, 13.2, NOW() - INTERVAL '36 hours');

-- Додаємо прогнози попиту
INSERT INTO demand_forecasts (route_id, forecast_date, day_of_week, predicted_passengers, confidence_lower, confidence_upper, created_at) VALUES
(1, (NOW() + INTERVAL '1 day')::date, EXTRACT(dow FROM NOW() + INTERVAL '1 day')::int, 38, 32, 44, NOW()),
(1, (NOW() + INTERVAL '2 days')::date, EXTRACT(dow FROM NOW() + INTERVAL '2 days')::int, 42, 36, 48, NOW()),
(2, (NOW() + INTERVAL '1 day')::date, EXTRACT(dow FROM NOW() + INTERVAL '1 day')::int, 28, 22, 34, NOW()),
(3, (NOW() + INTERVAL '1 day')::date, EXTRACT(dow FROM NOW() + INTERVAL '1 day')::int, 35, 28, 42, NOW());

-- Додаємо записи в журнал аудиту
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, old_values, new_values, ip_address, created_at) VALUES
(2, 'UPDATE', 'routes', 1, '{"base_price": 180.00}', '{"base_price": 200.00}', '192.168.1.100', NOW() - INTERVAL '2 hours'),
(2, 'CREATE', 'trips', (SELECT MAX(id) FROM trips), '{}', '{"route_id": 1, "bus_id": 1}', '192.168.1.100', NOW() - INTERVAL '1 hour'),
(1, 'UPDATE', 'system_settings', 1, '{"peak_hours_coefficient": 1.20}', '{"peak_hours_coefficient": 1.30}', '192.168.1.50', NOW() - INTERVAL '30 minutes');

-- Оновлюємо статистику
ANALYZE trips;
ANALYZE trip_analytics;
ANALYZE demand_forecasts;
ANALYZE audit_logs;
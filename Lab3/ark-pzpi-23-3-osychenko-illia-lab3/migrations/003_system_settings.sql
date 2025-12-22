-- Міграція для створення таблиці системних налаштувань
CREATE TABLE system_settings (
    id SERIAL PRIMARY KEY,
    fuel_price_per_liter DECIMAL(10,2) NOT NULL DEFAULT 50.00,
    peak_hours_coefficient DECIMAL(5,2) NOT NULL DEFAULT 1.20,
    weekend_coefficient DECIMAL(5,2) NOT NULL DEFAULT 1.15,
    high_demand_threshold INTEGER NOT NULL DEFAULT 85,
    low_demand_threshold INTEGER NOT NULL DEFAULT 30,
    price_min_coefficient DECIMAL(5,2) NOT NULL DEFAULT 0.70,
    price_max_coefficient DECIMAL(5,2) NOT NULL DEFAULT 1.50,
    seasonal_coefficients JSONB NOT NULL DEFAULT '{"new_year": 1.30, "summer": 1.15, "regular": 1.00}',
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by INTEGER REFERENCES users(id),
    
    CONSTRAINT check_fuel_price CHECK (fuel_price_per_liter >= 10 AND fuel_price_per_liter <= 200),
    CONSTRAINT check_peak_hours_coeff CHECK (peak_hours_coefficient >= 0.5 AND peak_hours_coefficient <= 3.0),
    CONSTRAINT check_weekend_coeff CHECK (weekend_coefficient >= 0.5 AND weekend_coefficient <= 3.0),
    CONSTRAINT check_demand_thresholds CHECK (high_demand_threshold > low_demand_threshold),
    CONSTRAINT check_low_demand_threshold CHECK (low_demand_threshold >= 0 AND low_demand_threshold <= 100),
    CONSTRAINT check_high_demand_threshold CHECK (high_demand_threshold >= 0 AND high_demand_threshold <= 100),
    CONSTRAINT check_price_coefficients CHECK (price_min_coefficient < price_max_coefficient),
    CONSTRAINT check_price_min_coeff CHECK (price_min_coefficient >= 0.1 AND price_min_coefficient <= 1.0),
    CONSTRAINT check_price_max_coeff CHECK (price_max_coefficient >= 1.0 AND price_max_coefficient <= 5.0)
);

-- Створюємо індекс для швидкого пошуку останніх налаштувань
CREATE INDEX idx_system_settings_updated_at ON system_settings(updated_at DESC);

-- Вставляємо початкові налаштування
INSERT INTO system_settings (
    fuel_price_per_liter,
    peak_hours_coefficient,
    weekend_coefficient,
    high_demand_threshold,
    low_demand_threshold,
    price_min_coefficient,
    price_max_coefficient,
    seasonal_coefficients
) VALUES (
    50.00,
    1.20,
    1.15,
    85,
    30,
    0.70,
    1.50,
    '{"new_year": 1.30, "summer": 1.15, "regular": 1.00}'
);

-- Додаємо коментарі до таблиці та колонок
COMMENT ON TABLE system_settings IS 'Системні налаштування для алгоритмів ціноутворення та аналітики';
COMMENT ON COLUMN system_settings.fuel_price_per_liter IS 'Ціна палива за літр (грн)';
COMMENT ON COLUMN system_settings.peak_hours_coefficient IS 'Коефіцієнт для пікових годин (7-9, 17-19)';
COMMENT ON COLUMN system_settings.weekend_coefficient IS 'Коефіцієнт для вихідних та святкових днів';
COMMENT ON COLUMN system_settings.high_demand_threshold IS 'Поріг високого попиту (% завантаженості)';
COMMENT ON COLUMN system_settings.low_demand_threshold IS 'Поріг низького попиту (% завантаженості)';
COMMENT ON COLUMN system_settings.price_min_coefficient IS 'Мінімальний коефіцієнт ціни (від базової)';
COMMENT ON COLUMN system_settings.price_max_coefficient IS 'Максимальний коефіцієнт ціни (від базової)';
COMMENT ON COLUMN system_settings.seasonal_coefficients IS 'Сезонні коефіцієнти у форматі JSON';
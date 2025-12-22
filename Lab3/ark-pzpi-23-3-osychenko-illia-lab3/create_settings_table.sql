CREATE TABLE IF NOT EXISTS system_settings (
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
    updated_by INTEGER
);

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
) ON CONFLICT DO NOTHING;
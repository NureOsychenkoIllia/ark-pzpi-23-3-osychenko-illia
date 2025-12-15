-- Адміністративні таблиці
CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE role_permissions (
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role_id INTEGER NOT NULL REFERENCES roles(id),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Бізнесові таблиці
CREATE TABLE routes (
    id SERIAL PRIMARY KEY,
    origin_city VARCHAR(100) NOT NULL,
    destination_city VARCHAR(100) NOT NULL,
    distance_km DECIMAL(10,2) NOT NULL,
    base_price DECIMAL(10,2) NOT NULL,
    fuel_cost_per_km DECIMAL(6,2) NOT NULL DEFAULT 4.00,
    driver_cost_per_trip DECIMAL(10,2) NOT NULL DEFAULT 800.00,
    estimated_duration_minutes INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT positive_distance CHECK (distance_km > 0),
    CONSTRAINT positive_price CHECK (base_price > 0)
);

CREATE TABLE buses (
    id SERIAL PRIMARY KEY,
    registration_number VARCHAR(20) NOT NULL UNIQUE,
    capacity INTEGER NOT NULL,
    model VARCHAR(100),
    fuel_consumption_per_100km DECIMAL(5,2) NOT NULL DEFAULT 25.00,
    is_active BOOLEAN DEFAULT TRUE,
    CONSTRAINT positive_capacity CHECK (capacity > 0)
);

CREATE TABLE devices (
    id SERIAL PRIMARY KEY,
    serial_number VARCHAR(50) NOT NULL UNIQUE,
    auth_token_hash VARCHAR(255) NOT NULL,
    bus_id INTEGER UNIQUE REFERENCES buses(id),
    firmware_version VARCHAR(20),
    last_sync_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE trips (
    id SERIAL PRIMARY KEY,
    route_id INTEGER NOT NULL REFERENCES routes(id),
    bus_id INTEGER NOT NULL REFERENCES buses(id),
    scheduled_departure TIMESTAMPTZ NOT NULL,
    actual_departure TIMESTAMPTZ,
    actual_arrival TIMESTAMPTZ,
    status VARCHAR(20) DEFAULT 'scheduled' 
        CHECK (status IN ('scheduled', 'boarding', 'in_progress', 'completed', 'cancelled')),
    current_passengers INTEGER DEFAULT 0,
    driver_name VARCHAR(255)
);

CREATE TABLE passenger_events (
    id SERIAL,
    trip_id INTEGER NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    event_type VARCHAR(10) NOT NULL CHECK (event_type IN ('entry', 'exit')),
    timestamp TIMESTAMPTZ NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    passenger_count_after INTEGER NOT NULL,
    device_local_id INTEGER,
    is_synced BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- Створюємо партицію для поточного місяця
CREATE TABLE passenger_events_2025_12 PARTITION OF passenger_events
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

CREATE TABLE price_recommendations (
    id SERIAL PRIMARY KEY,
    trip_id INTEGER NOT NULL REFERENCES trips(id),
    base_price DECIMAL(10,2) NOT NULL,
    recommended_price DECIMAL(10,2) NOT NULL,
    occupancy_rate DECIMAL(5,2) NOT NULL,
    demand_coefficient DECIMAL(4,2) NOT NULL,
    time_coefficient DECIMAL(4,2) NOT NULL,
    day_coefficient DECIMAL(4,2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Аналітичні таблиці
CREATE TABLE trip_analytics (
    id SERIAL PRIMARY KEY,
    trip_id INTEGER NOT NULL REFERENCES trips(id) UNIQUE,
    total_passengers INTEGER NOT NULL,
    max_passengers INTEGER NOT NULL,
    avg_occupancy_rate DECIMAL(5,2) NOT NULL,
    revenue DECIMAL(12,2) NOT NULL,
    fuel_cost DECIMAL(10,2) NOT NULL,
    driver_cost DECIMAL(10,2) NOT NULL,
    other_costs DECIMAL(10,2) DEFAULT 0,
    profit DECIMAL(12,2) NOT NULL,
    profitability_percent DECIMAL(6,2) NOT NULL,
    calculated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE demand_forecasts (
    id SERIAL PRIMARY KEY,
    route_id INTEGER NOT NULL REFERENCES routes(id),
    forecast_date DATE NOT NULL,
    day_of_week INTEGER NOT NULL,
    predicted_passengers INTEGER NOT NULL,
    confidence_lower INTEGER NOT NULL,
    confidence_upper INTEGER NOT NULL,
    actual_passengers INTEGER,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (route_id, forecast_date)
);

CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    trip_id INTEGER REFERENCES trips(id),
    type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) DEFAULT 'info' 
        CHECK (severity IN ('info', 'warning', 'critical')),
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Індекси для оптимізації
CREATE INDEX idx_passenger_events_trip_ts ON passenger_events(trip_id, timestamp);
CREATE INDEX idx_trips_route_departure ON trips(route_id, scheduled_departure);
CREATE INDEX idx_audit_logs_user_time ON audit_logs(user_id, created_at);
CREATE INDEX idx_notifications_user_read ON notifications(user_id, is_read);
CREATE INDEX idx_demand_forecasts_route_date ON demand_forecasts(route_id, forecast_date);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_devices_serial ON devices(serial_number);
CREATE INDEX idx_buses_registration ON buses(registration_number);
package model

import (
	"time"
)

// User представляє користувача системи
type User struct {
	ID           int64     `json:"id" db:"id" example:"1"`
	Email        string    `json:"email" db:"email" example:"user@example.com"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FullName     string    `json:"full_name" db:"full_name" example:"Іван Іванов"`
	RoleID       int64     `json:"role_id" db:"role_id" example:"2"`
	Role         *Role     `json:"role,omitempty"`
	IsActive     bool      `json:"is_active" db:"is_active" example:"true"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// Role представляє роль користувача
type Role struct {
	ID          int64        `json:"id" db:"id"`
	Name        string       `json:"name" db:"name"`
	Description string       `json:"description" db:"description"`
	Permissions []Permission `json:"permissions,omitempty"`
}

// Permission представляє дозвіл
type Permission struct {
	ID          int64  `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
}

// Route представляє маршрут
type Route struct {
	ID                   int64     `json:"id" db:"id" example:"1"`
	OriginCity           string    `json:"origin_city" db:"origin_city" example:"Харків"`
	DestinationCity      string    `json:"destination_city" db:"destination_city" example:"Київ"`
	DistanceKm           float64   `json:"distance_km" db:"distance_km" example:"480.5"`
	BasePrice            float64   `json:"base_price" db:"base_price" example:"250.00"`
	FuelCostPerKm        float64   `json:"fuel_cost_per_km" db:"fuel_cost_per_km" example:"2.50"`
	DriverCostPerTrip    float64   `json:"driver_cost_per_trip" db:"driver_cost_per_trip" example:"800.00"`
	EstimatedDurationMin int       `json:"estimated_duration_minutes" db:"estimated_duration_minutes" example:"360"`
	IsActive             bool      `json:"is_active" db:"is_active" example:"true"`
	CreatedAt            time.Time `json:"created_at" db:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// Bus представляє автобус
type Bus struct {
	ID                      int64   `json:"id" db:"id" example:"1"`
	RegistrationNumber      string  `json:"registration_number" db:"registration_number" example:"AA1234BB"`
	Capacity                int     `json:"capacity" db:"capacity" example:"50"`
	Model                   string  `json:"model" db:"model" example:"Mercedes Sprinter"`
	FuelConsumptionPer100km float64 `json:"fuel_consumption_per_100km" db:"fuel_consumption_per_100km" example:"12.5"`
	IsActive                bool    `json:"is_active" db:"is_active" example:"true"`
}

// Device представляє IoT-пристрій
type Device struct {
	ID              int64      `json:"id" db:"id"`
	SerialNumber    string     `json:"serial_number" db:"serial_number"`
	AuthTokenHash   string     `json:"-" db:"auth_token_hash"`
	BusID           *int64     `json:"bus_id" db:"bus_id"`
	Bus             *Bus       `json:"bus,omitempty"`
	FirmwareVersion string     `json:"firmware_version" db:"firmware_version"`
	LastSyncAt      *time.Time `json:"last_sync_at" db:"last_sync_at"`
	IsActive        bool       `json:"is_active" db:"is_active"`
}

// Trip представляє рейс
type Trip struct {
	ID                 int64      `json:"id" db:"id" example:"1"`
	RouteID            int64      `json:"route_id" db:"route_id" example:"1"`
	Route              *Route     `json:"route,omitempty"`
	BusID              int64      `json:"bus_id" db:"bus_id" example:"1"`
	Bus                *Bus       `json:"bus,omitempty"`
	ScheduledDeparture time.Time  `json:"scheduled_departure" db:"scheduled_departure" example:"2023-12-15T08:00:00Z"`
	ActualDeparture    *time.Time `json:"actual_departure" db:"actual_departure" example:"2023-12-15T08:05:00Z"`
	ActualArrival      *time.Time `json:"actual_arrival" db:"actual_arrival" example:"2023-12-15T14:05:00Z"`
	Status             string     `json:"status" db:"status" example:"completed" enums:"scheduled,in_progress,completed,cancelled"`
	CurrentPassengers  int        `json:"current_passengers" db:"current_passengers" example:"35"`
	DriverName         string     `json:"driver_name" db:"driver_name" example:"Петро Петренко"`
}

// PassengerEvent представляє подію пасажира
type PassengerEvent struct {
	ID                  int64     `json:"id" db:"id" example:"1"`
	TripID              int64     `json:"trip_id" db:"trip_id" example:"1"`
	EventType           string    `json:"event_type" db:"event_type" example:"board" enums:"board,alight"`
	Timestamp           time.Time `json:"timestamp" db:"timestamp" example:"2023-12-15T08:15:00Z"`
	Latitude            *float64  `json:"latitude" db:"latitude" example:"49.9935"`
	Longitude           *float64  `json:"longitude" db:"longitude" example:"36.2304"`
	PassengerCountAfter int       `json:"passenger_count_after" db:"passenger_count_after" example:"25"`
	DeviceLocalID       *int      `json:"device_local_id" db:"device_local_id" example:"123"`
	IsSynced            bool      `json:"is_synced" db:"is_synced" example:"true"`
}

// PriceRecommendation представляє рекомендацію ціни
type PriceRecommendation struct {
	ID               int64     `json:"id" db:"id"`
	TripID           int64     `json:"trip_id" db:"trip_id"`
	BasePrice        float64   `json:"base_price" db:"base_price"`
	RecommendedPrice float64   `json:"recommended_price" db:"recommended_price"`
	OccupancyRate    float64   `json:"occupancy_rate" db:"occupancy_rate"`
	DemandCoeff      float64   `json:"demand_coefficient" db:"demand_coefficient"`
	TimeCoeff        float64   `json:"time_coefficient" db:"time_coefficient"`
	DayCoeff         float64   `json:"day_coefficient" db:"day_coefficient"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// TripAnalytics представляє аналітику рейсу
type TripAnalytics struct {
	ID                   int64     `json:"id" db:"id" example:"1"`
	TripID               int64     `json:"trip_id" db:"trip_id" example:"1"`
	TotalPassengers      int       `json:"total_passengers" db:"total_passengers" example:"142"`
	MaxPassengers        int       `json:"max_passengers" db:"max_passengers" example:"45"`
	AvgOccupancyRate     float64   `json:"avg_occupancy_rate" db:"avg_occupancy_rate" example:"0.75"`
	Revenue              float64   `json:"revenue" db:"revenue" example:"3550.00"`
	FuelCost             float64   `json:"fuel_cost" db:"fuel_cost" example:"1200.25"`
	DriverCost           float64   `json:"driver_cost" db:"driver_cost" example:"800.00"`
	OtherCosts           float64   `json:"other_costs" db:"other_costs" example:"150.00"`
	Profit               float64   `json:"profit" db:"profit" example:"1399.75"`
	ProfitabilityPercent float64   `json:"profitability_percent" db:"profitability_percent" example:"39.43"`
	CalculatedAt         time.Time `json:"calculated_at" db:"calculated_at" example:"2023-12-15T20:00:00Z"`
}

// DemandForecast представляє прогноз попиту
type DemandForecast struct {
	ID                  int64     `json:"id" db:"id"`
	RouteID             int64     `json:"route_id" db:"route_id"`
	ForecastDate        time.Time `json:"forecast_date" db:"forecast_date"`
	DayOfWeek           int       `json:"day_of_week" db:"day_of_week"`
	PredictedPassengers int       `json:"predicted_passengers" db:"predicted_passengers"`
	ConfidenceLower     int       `json:"confidence_lower" db:"confidence_lower"`
	ConfidenceUpper     int       `json:"confidence_upper" db:"confidence_upper"`
	ActualPassengers    *int      `json:"actual_passengers" db:"actual_passengers"`
	TrendCoefficient    float64   `json:"trend_coefficient,omitempty"`
	SeasonCoefficient   float64   `json:"season_coefficient,omitempty"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

// Notification представляє сповіщення
type Notification struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	TripID    *int64    `json:"trip_id" db:"trip_id"`
	Type      string    `json:"type" db:"type"`
	Severity  string    `json:"severity" db:"severity"`
	Message   string    `json:"message" db:"message"`
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuditLog представляє запис журналу аудиту
type AuditLog struct {
	ID         int64          `json:"id" db:"id"`
	UserID     *int64         `json:"user_id" db:"user_id"`
	User       *User          `json:"user,omitempty"`
	Action     string         `json:"action" db:"action"`
	EntityType string         `json:"entity_type" db:"entity_type"`
	EntityID   *int64         `json:"entity_id" db:"entity_id"`
	OldValues  map[string]any `json:"old_values" db:"old_values"`
	NewValues  map[string]any `json:"new_values" db:"new_values"`
	IPAddress  string         `json:"ip_address" db:"ip_address"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
}

// SystemSettings представляє системні налаштування
type SystemSettings struct {
	ID                   int64              `json:"id" db:"id"`
	FuelPricePerLiter    float64            `json:"fuel_price_per_liter" db:"fuel_price_per_liter" example:"50.00"`
	PeakHoursCoefficient float64            `json:"peak_hours_coefficient" db:"peak_hours_coefficient" example:"1.20"`
	WeekendCoefficient   float64            `json:"weekend_coefficient" db:"weekend_coefficient" example:"1.15"`
	HighDemandThreshold  int                `json:"high_demand_threshold" db:"high_demand_threshold" example:"85"`
	LowDemandThreshold   int                `json:"low_demand_threshold" db:"low_demand_threshold" example:"30"`
	PriceMinCoefficient  float64            `json:"price_min_coefficient" db:"price_min_coefficient" example:"0.70"`
	PriceMaxCoefficient  float64            `json:"price_max_coefficient" db:"price_max_coefficient" example:"1.50"`
	SeasonalCoefficients map[string]float64 `json:"seasonal_coefficients" db:"seasonal_coefficients"`
	UpdatedAt            time.Time          `json:"updated_at" db:"updated_at"`
	UpdatedBy            *int64             `json:"updated_by" db:"updated_by"`
	UpdatedByUser        *User              `json:"updated_by_user,omitempty"`
}

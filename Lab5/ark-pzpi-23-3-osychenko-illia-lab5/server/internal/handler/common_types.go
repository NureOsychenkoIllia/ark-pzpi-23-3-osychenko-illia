package handler

import (
	"busoptima/internal/model"
	"time"
)

// ErrorResponse представляє стандартну відповідь з помилкою
type ErrorResponse struct {
	Error string `json:"error" example:"Error message"`
}

// MessageResponse представляє стандартну відповідь з повідомленням про успіх
type MessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

// AuditLogsResponse представляє відповідь для ендпоінту журналу аудиту
type AuditLogsResponse struct {
	Logs  []model.AuditLog `json:"logs"`
	Total int64            `json:"total" example:"150"`
	Page  int              `json:"page" example:"1"`
	Limit int              `json:"limit" example:"20"`
}

// UpdateUserRoleResponse представляє відповідь для оновлення ролі користувача
type UpdateUserRoleResponse struct {
	Message string `json:"message" example:"User role updated successfully"`
}

// UpdateSystemSettingsResponse представляє відповідь для оновлення системних налаштувань
type UpdateSystemSettingsResponse struct {
	Message   string                  `json:"message" example:"System settings updated successfully"`
	Settings  *SystemSettingsResponse `json:"settings"`
	UpdatedAt *time.Time              `json:"updated_at" example:"2023-12-15T10:30:00Z"`
}

// SystemSettingsResponse представляє системні налаштування в API відповідях
type SystemSettingsResponse struct {
	ID                   int64                `json:"id"`
	FuelPricePerLiter    float64              `json:"fuel_price_per_liter" example:"50.00"`
	PeakHoursCoefficient float64              `json:"peak_hours_coefficient" example:"1.20"`
	WeekendCoefficient   float64              `json:"weekend_coefficient" example:"1.15"`
	HighDemandThreshold  int                  `json:"high_demand_threshold" example:"85"`
	LowDemandThreshold   int                  `json:"low_demand_threshold" example:"30"`
	PriceMinCoefficient  float64              `json:"price_min_coefficient" example:"0.70"`
	PriceMaxCoefficient  float64              `json:"price_max_coefficient" example:"1.50"`
	SeasonalCoefficients SeasonalCoefficients `json:"seasonal_coefficients"`
	UpdatedAt            time.Time            `json:"updated_at"`
	UpdatedBy            *int64               `json:"updated_by"`
	UpdatedByUser        *model.User          `json:"updated_by_user,omitempty"`
}

// ConfidenceInterval представляє довірчий інтервал для прогнозів
type ConfidenceInterval struct {
	Lower int `json:"lower" example:"20"`
	Upper int `json:"upper" example:"35"`
}

// ForecastItem представляє окремий елемент прогнозу з рекомендаціями
type ForecastItem struct {
	Date                 string             `json:"date" example:"2023-12-15"`
	DayOfWeek            string             `json:"day_of_week" example:"Friday"`
	PredictedPassengers  int                `json:"predicted_passengers" example:"28"`
	ConfidenceInterval   ConfidenceInterval `json:"confidence_interval"`
	TrendCoefficient     float64            `json:"trend_coefficient" example:"1.05"`
	SeasonCoefficient    float64            `json:"season_coefficient" example:"1.15"`
	Recommendation       string             `json:"recommendation" example:"Optimal capacity"`
	RecommendationDetail string             `json:"recommendation_detail" example:"Expected passenger load is within optimal range"`
}

// ForecastResponse представляє повну відповідь прогнозу
type ForecastResponse struct {
	Route     *model.Route   `json:"route"`
	Algorithm string         `json:"algorithm" example:"linear_regression"`
	Forecasts []ForecastItem `json:"forecasts"`
}

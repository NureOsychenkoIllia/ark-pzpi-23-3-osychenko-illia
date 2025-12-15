package handler

import (
	"busoptima/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticsService
	forecastService  service.ForecastService
}

func NewAnalyticsHandler(analyticsService service.AnalyticsService, forecastService service.ForecastService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		forecastService:  forecastService,
	}
}

// GetDashboard повертає дані для дашборду
//
//	@Summary		Отримати дані дашборду
//	@Description	Повертає основні метрики та статистику для дашборду
//	@Tags			Analytics
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/analytics/dashboard [get]
func (h *AnalyticsHandler) GetDashboard(c *fiber.Ctx) error {
	dashboard, err := h.analyticsService.GetDashboard(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(dashboard)
}

// GetForecast повертає прогноз попиту
//
//	@Summary		Отримати прогноз попиту
//	@Description	Повертає прогноз попиту на маршрути з аналітикою
//	@Tags			Analytics
//	@Accept			json
//	@Produce		json
//	@Param			route_id	query		int		false	"ID маршруту для прогнозу"
//	@Param			days		query		int		false	"Кількість днів для прогнозу"	default(7)
//	@Success		200			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/analytics/forecast [get]
func (h *AnalyticsHandler) GetForecast(c *fiber.Ctx) error {
	// TODO: Parse query parameters and call forecastService.ForecastDemand
	// Demo response with forecast data
	return c.JSON(fiber.Map{
		"route": fiber.Map{
			"id":               1,
			"origin_city":      "Харків",
			"destination_city": "Київ",
		},
		"forecasts": []fiber.Map{
			{
				"date":                    "2025-12-15",
				"day_of_week":             "Monday",
				"predicted_passengers":    42,
				"confidence_interval":     fiber.Map{"lower": 35, "upper": 49},
				"trend_coefficient":       1.05,
				"seasonality_coefficient": 1.00,
				"recommendation":          "normal",
				"recommendation_details":  "Очікується стандартна завантаженість",
			},
		},
		"algorithm": "moving_average_4w_trend_seasonality",
	})
}

// GetProfitability повертає аналіз прибутковості
//
//	@Summary		Отримати аналіз прибутковості
//	@Description	Повертає детальний аналіз прибутковості маршрутів та рейсів
//	@Tags			Analytics
//	@Accept			json
//	@Produce		json
//	@Param			date_from	query		string	false	"Дата початку періоду (YYYY-MM-DD)"
//	@Param			date_to		query		string	false	"Дата кінця періоду (YYYY-MM-DD)"
//	@Param			route_id	query		int		false	"ID маршруту для фільтрації"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/analytics/profitability [get]
func (h *AnalyticsHandler) GetProfitability(c *fiber.Ctx) error {
	// TODO: Parse query parameters and call analyticsService.GetProfitability
	// Demo response with profitability data
	return c.JSON(fiber.Map{
		"period": fiber.Map{
			"from": "2025-12-01",
			"to":   "2025-12-14",
		},
		"summary": fiber.Map{
			"total_trips":           156,
			"total_revenue":         234500.00,
			"total_costs":           156300.00,
			"total_profit":          78200.00,
			"average_profitability": 50.03,
		},
		"by_route": []fiber.Map{
			{
				"route_id":         1,
				"route_name":       "Харків - Київ",
				"trips_count":      42,
				"total_passengers": 1680,
				"avg_occupancy":    68.5,
				"revenue":          75600.00,
				"costs":            48200.00,
				"profit":           27400.00,
				"profitability":    56.85,
				"category":         "high_profit",
			},
		},
	})
}

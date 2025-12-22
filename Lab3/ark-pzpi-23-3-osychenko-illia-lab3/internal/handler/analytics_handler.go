package handler

import (
	"busoptima/internal/service"
	"strconv"
	"time"

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
//	@Description	Повертає основні метрики та статистику для інформаційної панелі
//	@Tags			Analytics
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	service.DashboardData
//	@Failure		500	{object}	ErrorResponse
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
//	@Param			route_id	query		int		true	"ID маршруту для прогнозу"
//	@Param			date		query		string	false	"Дата прогнозу (YYYY-MM-DD)"
//	@Success		200			{object}	service.ForecastResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/analytics/forecast [get]
func (h *AnalyticsHandler) GetForecast(c *fiber.Ctx) error {
	routeIDStr := c.Query("route_id")
	if routeIDStr == "" {
		return c.Status(400).JSON(fiber.Map{"error": "route_id is required"})
	}

	routeID, err := strconv.ParseInt(routeIDStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid route_id"})
	}

	// Парсимо дату або використовуємо завтра
	targetDate := time.Now().AddDate(0, 0, 1)
	if dateStr := c.Query("date"); dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid date format, use YYYY-MM-DD"})
		}
		targetDate = parsed
	}

	forecast, err := h.forecastService.ForecastDemand(c.Context(), routeID, targetDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Додаємо рекомендацію на основі прогнозу
	response := fiber.Map{
		"route":     forecast.Route,
		"algorithm": forecast.Algorithm,
		"forecasts": make([]fiber.Map, len(forecast.Forecasts)),
	}

	for i, f := range forecast.Forecasts {
		recommendation, details := generateRecommendation(f.PredictedPassengers, 50) // 50 - типова місткість
		response["forecasts"].([]fiber.Map)[i] = fiber.Map{
			"date":                   f.ForecastDate.Format("2006-01-02"),
			"day_of_week":            dayOfWeekName(f.DayOfWeek),
			"predicted_passengers":   f.PredictedPassengers,
			"confidence_interval":    fiber.Map{"lower": f.ConfidenceLower, "upper": f.ConfidenceUpper},
			"trend_coefficient":      f.TrendCoefficient,
			"season_coefficient":     f.SeasonCoefficient,
			"recommendation":         recommendation,
			"recommendation_details": details,
		}
	}

	return c.JSON(response)
}

// GetForecasts повертає збережені прогнози за період
//
//	@Summary		Отримати збережені прогнози
//	@Description	Повертає раніше розраховані прогнози попиту за період
//	@Tags			Analytics
//	@Accept			json
//	@Produce		json
//	@Param			route_id	query		int		true	"ID маршруту"
//	@Param			date_from	query		string	true	"Дата початку (YYYY-MM-DD)"
//	@Param			date_to		query		string	true	"Дата кінця (YYYY-MM-DD)"
//	@Success		200			{object}	service.ForecastsResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/analytics/forecasts [get]
func (h *AnalyticsHandler) GetForecasts(c *fiber.Ctx) error {
	routeID, err := strconv.ParseInt(c.Query("route_id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid route_id"})
	}

	from, err := time.Parse("2006-01-02", c.Query("date_from"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid date_from"})
	}

	to, err := time.Parse("2006-01-02", c.Query("date_to"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid date_to"})
	}

	forecasts, err := h.forecastService.GetForecasts(c.Context(), routeID, from, to)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(forecasts)
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
//	@Success		200			{object}	service.ProfitabilityData
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/analytics/profitability [get]
func (h *AnalyticsHandler) GetProfitability(c *fiber.Ctx) error {
	// Парсимо параметри
	var routeID int64
	if routeIDStr := c.Query("route_id"); routeIDStr != "" {
		var err error
		routeID, err = strconv.ParseInt(routeIDStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid route_id"})
		}
	}

	// За замовчуванням - останні 30 днів
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		parsed, err := time.Parse("2006-01-02", dateFrom)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid date_from format"})
		}
		from = parsed
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		parsed, err := time.Parse("2006-01-02", dateTo)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid date_to format"})
		}
		to = parsed
	}

	profitability, err := h.analyticsService.GetProfitability(c.Context(), routeID, from, to)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(profitability)
}

// GetTripAnalytics повертає аналітику конкретного рейсу
//
//	@Summary		Отримати аналітику рейсу
//	@Description	Розраховує та повертає аналітику для конкретного рейсу
//	@Tags			Analytics
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"ID рейсу"
//	@Success		200	{object}	model.TripAnalytics
//	@Failure		400	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/trips/{id}/analytics [get]
func (h *AnalyticsHandler) GetTripAnalytics(c *fiber.Ctx) error {
	tripID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid trip ID"})
	}

	// Спочатку пробуємо отримати існуючу аналітику
	analytics, err := h.analyticsService.GetTripAnalytics(c.Context(), tripID)
	if err != nil {
		// Якщо не знайдено - розраховуємо
		analytics, err = h.analyticsService.CalculateTripAnalytics(c.Context(), tripID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(analytics)
}

// CalculateTripAnalytics примусово перераховує аналітику рейсу
//
//	@Summary		Перерахувати аналітику рейсу
//	@Description	Примусово перераховує аналітику для конкретного рейсу
//	@Tags			Analytics
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"ID рейсу"
//	@Success		200	{object}	model.TripAnalytics
//	@Failure		400	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/trips/{id}/analytics/calculate [post]
func (h *AnalyticsHandler) CalculateTripAnalytics(c *fiber.Ctx) error {
	tripID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid trip ID"})
	}

	analytics, err := h.analyticsService.CalculateTripAnalytics(c.Context(), tripID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(analytics)
}

// generateRecommendation генерує рекомендацію на основі прогнозу
func generateRecommendation(predicted, capacity int) (string, string) {
	occupancy := float64(predicted) / float64(capacity)

	switch {
	case occupancy > 0.95:
		return "add_trip", "Рекомендується додати додатковий рейс через високий попит"
	case occupancy < 0.25:
		return "cancel_trip", "Рекомендується скасувати рейс через низький попит"
	case occupancy < 0.40:
		return "reduce_price", "Рекомендується знизити ціну для залучення пасажирів"
	case occupancy > 0.80:
		return "increase_price", "Можливе підвищення ціни через високий попит"
	default:
		return "normal", "Очікується стандартна завантаженість"
	}
}

// dayOfWeekName повертає назву дня тижня
func dayOfWeekName(dow int) string {
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if dow >= 0 && dow < 7 {
		return days[dow]
	}
	return "Unknown"
}

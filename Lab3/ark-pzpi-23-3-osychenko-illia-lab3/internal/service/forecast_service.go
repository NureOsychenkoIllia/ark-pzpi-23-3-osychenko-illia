package service

import (
	"busoptima/internal/model"
	"busoptima/internal/repository"
	"context"
	"math"
	"time"
)

// ForecastService інтерфейс для прогнозування попиту
type ForecastService interface {
	ForecastDemand(ctx context.Context, routeID int64, targetDate time.Time) (*ForecastResponse, error)
	GetForecasts(ctx context.Context, routeID int64, from, to time.Time) (*ForecastsResponse, error)
}

type ForecastResponse struct {
	Route     *model.Route           `json:"route"`
	Forecasts []model.DemandForecast `json:"forecasts"`
	Algorithm string                 `json:"algorithm"`
}

type ForecastsResponse struct {
	Route     *model.Route           `json:"route"`
	Forecasts []model.DemandForecast `json:"forecasts"`
}

type forecastService struct {
	analyticsRepo repository.AnalyticsRepository
	routeRepo     repository.RouteRepository
}

func NewForecastService(analyticsRepo repository.AnalyticsRepository, routeRepo repository.RouteRepository) ForecastService {
	return &forecastService{
		analyticsRepo: analyticsRepo,
		routeRepo:     routeRepo,
	}
}

// ForecastDemand прогнозує попит на конкретну дату методом ковзного середнього
func (s *forecastService) ForecastDemand(ctx context.Context, routeID int64, targetDate time.Time) (*ForecastResponse, error) {
	route, err := s.routeRepo.GetByID(ctx, routeID)
	if err != nil {
		return nil, err
	}

	dayOfWeek := int(targetDate.Weekday())

	// Отримуємо історичні дані за 12 тижнів
	historical, err := s.analyticsRepo.GetHistoricalPassengers(ctx, routeID, dayOfWeek, 12)
	if err != nil {
		return nil, err
	}

	forecast := s.calculateForecast(routeID, targetDate, dayOfWeek, historical)

	// Зберігаємо прогноз
	if err := s.analyticsRepo.SaveDemandForecast(ctx, &forecast); err != nil {
		return nil, err
	}

	return &ForecastResponse{
		Route:     route,
		Forecasts: []model.DemandForecast{forecast},
		Algorithm: "moving_average_4w_trend_seasonality",
	}, nil
}

// calculateForecast виконує розрахунок прогнозу
func (s *forecastService) calculateForecast(routeID int64, targetDate time.Time, dayOfWeek int, historical []int) model.DemandForecast {
	// Базовий прогноз: ковзне середнє за 4 тижні
	movingAvg := s.calculateMovingAverage(historical, 4)

	// Коефіцієнт тренду
	trendCoeff := s.calculateTrendCoefficient(historical)

	// Коефіцієнт сезонності
	seasonCoeff := s.calculateSeasonCoefficient(targetDate)

	// Фінальний прогноз
	predicted := movingAvg * trendCoeff * seasonCoeff

	// Стандартне відхилення для довірчого інтервалу
	stdDev := s.calculateStdDev(historical, movingAvg)

	// 95% довірчий інтервал
	confidenceLower := int(math.Max(0, predicted-1.96*stdDev))
	confidenceUpper := int(predicted + 1.96*stdDev)

	return model.DemandForecast{
		RouteID:             routeID,
		ForecastDate:        targetDate,
		DayOfWeek:           dayOfWeek,
		PredictedPassengers: int(math.Round(predicted)),
		ConfidenceLower:     confidenceLower,
		ConfidenceUpper:     confidenceUpper,
		TrendCoefficient:    trendCoeff,
		SeasonCoefficient:   seasonCoeff,
	}
}

// calculateMovingAverage обчислює ковзне середнє
func (s *forecastService) calculateMovingAverage(data []int, weeks int) float64 {
	if len(data) == 0 {
		return 30.0 // значення за замовчуванням
	}

	count := min(weeks, len(data))
	sum := 0
	for i := 0; i < count; i++ {
		sum += data[i]
	}
	return float64(sum) / float64(count)
}

// calculateTrendCoefficient обчислює коефіцієнт тренду
func (s *forecastService) calculateTrendCoefficient(data []int) float64 {
	if len(data) < 8 {
		return 1.0
	}

	// Сума за останні 4 тижні
	recent := 0
	for i := 0; i < 4; i++ {
		recent += data[i]
	}

	// Сума за попередні 4 тижні
	previous := 0
	for i := 4; i < 8; i++ {
		previous += data[i]
	}

	if previous == 0 {
		return 1.0
	}

	trend := float64(recent) / float64(previous)

	// Обмежуємо коефіцієнт тренду
	return math.Max(0.7, math.Min(1.5, trend))
}

// calculateSeasonCoefficient визначає коефіцієнт сезонності
func (s *forecastService) calculateSeasonCoefficient(date time.Time) float64 {
	month := date.Month()
	day := date.Day()

	// Новорічні свята (25 грудня - 10 січня)
	if (month == time.December && day >= 25) || (month == time.January && day <= 10) {
		return 1.30
	}

	// Великодній період (приблизно)
	if month == time.April && day >= 10 && day <= 25 {
		return 1.25
	}

	// Літні канікули
	if month >= time.June && month <= time.August {
		return 1.15
	}

	// Зимові канікули
	if month == time.January && day > 10 && day <= 31 {
		return 1.10
	}

	return 1.0
}

// calculateStdDev обчислює стандартне відхилення
func (s *forecastService) calculateStdDev(data []int, mean float64) float64 {
	if len(data) < 2 {
		return 5.0 // значення за замовчуванням
	}

	count := min(4, len(data))
	sumSquares := 0.0
	for i := 0; i < count; i++ {
		diff := float64(data[i]) - mean
		sumSquares += diff * diff
	}

	return math.Sqrt(sumSquares / float64(count))
}

func (s *forecastService) GetForecasts(ctx context.Context, routeID int64, from, to time.Time) (*ForecastsResponse, error) {
	route, err := s.routeRepo.GetByID(ctx, routeID)
	if err != nil {
		return nil, err
	}

	forecasts, err := s.analyticsRepo.GetDemandForecasts(ctx, routeID, from, to)
	if err != nil {
		return nil, err
	}

	return &ForecastsResponse{
		Route:     route,
		Forecasts: forecasts,
	}, nil
}

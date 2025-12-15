package service

import (
	"context"
	"time"
	"busoptima/internal/model"
	"busoptima/internal/repository"
)

// ForecastService інтерфейс для прогнозування попиту
type ForecastService interface {
	ForecastDemand(ctx context.Context, routeID int64, targetDate time.Time) (*ForecastResponse, error)
	GetForecasts(ctx context.Context, routeID int64, from, to time.Time) (*ForecastsResponse, error)
}

type ForecastResponse struct {
	Route                *model.Route         `json:"route"`
	Forecasts            []model.DemandForecast `json:"forecasts"`
	Algorithm            string               `json:"algorithm"`
}

type ForecastsResponse struct {
	Route     *model.Route         `json:"route"`
	Forecasts []model.DemandForecast `json:"forecasts"`
}

type forecastService struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewForecastService(analyticsRepo repository.AnalyticsRepository) ForecastService {
	return &forecastService{analyticsRepo: analyticsRepo}
}

func (s *forecastService) ForecastDemand(ctx context.Context, routeID int64, targetDate time.Time) (*ForecastResponse, error) {
	// TODO: Implement demand forecasting algorithm (Lab3)
	forecast := model.DemandForecast{
		RouteID:             routeID,
		ForecastDate:        targetDate,
		DayOfWeek:           int(targetDate.Weekday()),
		PredictedPassengers: 42,
		ConfidenceLower:     35,
		ConfidenceUpper:     49,
		TrendCoefficient:    1.05,
		SeasonCoefficient:   1.00,
	}

	return &ForecastResponse{
		Forecasts: []model.DemandForecast{forecast},
		Algorithm: "moving_average_4w_trend_seasonality",
	}, nil
}

func (s *forecastService) GetForecasts(ctx context.Context, routeID int64, from, to time.Time) (*ForecastsResponse, error) {
	forecasts, err := s.analyticsRepo.GetDemandForecasts(ctx, routeID, from, to)
	if err != nil {
		return nil, err
	}

	return &ForecastsResponse{
		Forecasts: forecasts,
	}, nil
}
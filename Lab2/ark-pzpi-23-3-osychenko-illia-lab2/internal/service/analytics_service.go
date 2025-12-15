package service

import (
	"context"
	"time"
	"busoptima/internal/model"
	"busoptima/internal/repository"
)

// AnalyticsService інтерфейс для роботи з аналітикою
type AnalyticsService interface {
	GetDashboard(ctx context.Context) (*DashboardData, error)
	GetProfitability(ctx context.Context, routeID int64, from, to time.Time) (*ProfitabilityData, error)
}

type DashboardData struct {
	ActiveTrips     int     `json:"active_trips"`
	TotalRevenue    float64 `json:"total_revenue"`
	AvgOccupancy    float64 `json:"avg_occupancy"`
	ProfitableTrips int     `json:"profitable_trips"`
}

type ProfitabilityData struct {
	Period  map[string]string         `json:"period"`
	Summary map[string]interface{}    `json:"summary"`
	ByRoute []model.TripAnalytics     `json:"by_route"`
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{analyticsRepo: analyticsRepo}
}

func (s *analyticsService) GetDashboard(ctx context.Context) (*DashboardData, error) {
	// TODO: Implement real dashboard data aggregation from database
	return &DashboardData{
		ActiveTrips:     12,
		TotalRevenue:    25600.00,
		AvgOccupancy:    68.5,
		ProfitableTrips: 10,
	}, nil
}

func (s *analyticsService) GetProfitability(ctx context.Context, routeID int64, from, to time.Time) (*ProfitabilityData, error) {
	// TODO: Implement real profitability calculation from trip_analytics table
	return &ProfitabilityData{
		Period: map[string]string{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		},
		Summary: map[string]interface{}{
			"total_trips":            42,
			"total_revenue":          75600.00,
			"total_costs":            48200.00,
			"total_profit":           27400.00,
			"average_profitability":  56.85,
		},
	}, nil
}
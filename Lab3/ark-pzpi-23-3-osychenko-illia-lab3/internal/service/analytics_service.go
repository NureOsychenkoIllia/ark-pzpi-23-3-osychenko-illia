package service

import (
	"busoptima/internal/model"
	"busoptima/internal/repository"
	"context"
	"time"
)

// AnalyticsService інтерфейс для роботи з аналітикою
type AnalyticsService interface {
	GetDashboard(ctx context.Context) (*DashboardData, error)
	GetProfitability(ctx context.Context, routeID int64, from, to time.Time) (*ProfitabilityData, error)
	CalculateTripAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error)
	GetTripAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error)
}

type DashboardData struct {
	ActiveTrips       int            `json:"active_trips"`
	TotalPassengers   int            `json:"total_passengers"`
	TotalRevenue      float64        `json:"total_revenue"`
	TotalProfit       float64        `json:"total_profit"`
	AvgOccupancy      float64        `json:"avg_occupancy"`
	AvgProfitability  float64        `json:"avg_profitability"`
	ProfitableTrips   int            `json:"profitable_trips"`
	UnprofitableTrips int            `json:"unprofitable_trips"`
	TripsByCategory   map[string]int `json:"trips_by_category"`
}

type ProfitabilityData struct {
	Period  PeriodInfo           `json:"period"`
	Summary ProfitabilitySummary `json:"summary"`
	ByRoute []RouteProfitability `json:"by_route"`
}

type PeriodInfo struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ProfitabilitySummary struct {
	TotalTrips           int     `json:"total_trips"`
	TotalPassengers      int     `json:"total_passengers"`
	TotalRevenue         float64 `json:"total_revenue"`
	TotalCosts           float64 `json:"total_costs"`
	TotalProfit          float64 `json:"total_profit"`
	AverageProfitability float64 `json:"average_profitability"`
	AvgOccupancy         float64 `json:"avg_occupancy"`
}

type RouteProfitability struct {
	RouteID         int64   `json:"route_id"`
	RouteName       string  `json:"route_name"`
	TripsCount      int     `json:"trips_count"`
	TotalPassengers int     `json:"total_passengers"`
	AvgOccupancy    float64 `json:"avg_occupancy"`
	Revenue         float64 `json:"revenue"`
	Costs           float64 `json:"costs"`
	Profit          float64 `json:"profit"`
	Profitability   float64 `json:"profitability"`
	Category        string  `json:"category"`
}

type analyticsService struct {
	analyticsRepo repository.AnalyticsRepository
	tripRepo      repository.TripRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository, tripRepo repository.TripRepository) AnalyticsService {
	return &analyticsService{
		analyticsRepo: analyticsRepo,
		tripRepo:      tripRepo,
	}
}

// GetDashboard повертає агреговані дані для панелі моніторингу
func (s *analyticsService) GetDashboard(ctx context.Context) (*DashboardData, error) {
	// Отримуємо активні рейси
	trips, err := s.tripRepo.GetAll(ctx, map[string]interface{}{
		"status": "in_progress",
	})
	if err != nil {
		return nil, err
	}

	// Отримуємо аналітику за останні 7 днів для демонстрації
	from := time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
	to := time.Now().Add(24 * time.Hour)

	analytics, err := s.analyticsRepo.GetProfitabilityByRoute(ctx, 0, from, to)
	if err != nil {
		// Якщо немає даних, повертаємо базові значення
		analytics = []model.TripAnalytics{}
	}

	dashboard := &DashboardData{
		ActiveTrips:     len(trips),
		TripsByCategory: make(map[string]int),
	}

	// Агрегуємо дані з аналітики
	for _, a := range analytics {
		dashboard.TotalPassengers += a.TotalPassengers
		dashboard.TotalRevenue += a.Revenue
		dashboard.TotalProfit += a.Profit
		dashboard.AvgOccupancy += a.AvgOccupancyRate
		dashboard.AvgProfitability += a.ProfitabilityPercent

		category := categorizeProfitability(a.ProfitabilityPercent)
		dashboard.TripsByCategory[category]++

		if a.ProfitabilityPercent >= 0 {
			dashboard.ProfitableTrips++
		} else {
			dashboard.UnprofitableTrips++
		}
	}

	// Обчислюємо середні значення
	if len(analytics) > 0 {
		dashboard.AvgOccupancy /= float64(len(analytics))
		dashboard.AvgProfitability /= float64(len(analytics))
	}

	return dashboard, nil
}

// GetProfitability повертає аналітику рентабельності за період
func (s *analyticsService) GetProfitability(ctx context.Context, routeID int64, from, to time.Time) (*ProfitabilityData, error) {
	analytics, err := s.analyticsRepo.GetProfitabilityByRoute(ctx, routeID, from, to)
	if err != nil {
		return nil, err
	}

	// Групуємо за маршрутами
	routeMap := make(map[int64]*RouteProfitability)

	summary := ProfitabilitySummary{}

	for _, a := range analytics {
		summary.TotalTrips++
		summary.TotalPassengers += a.TotalPassengers
		summary.TotalRevenue += a.Revenue
		summary.TotalCosts += a.FuelCost + a.DriverCost + a.OtherCosts
		summary.TotalProfit += a.Profit
		summary.AvgOccupancy += a.AvgOccupancyRate

		// Отримуємо route_id з trip
		trip, err := s.tripRepo.GetByID(ctx, a.TripID)
		if err != nil {
			continue
		}

		rp, exists := routeMap[trip.RouteID]
		if !exists {
			routeName := "Невідомий маршрут"
			if trip.Route != nil {
				routeName = trip.Route.OriginCity + " - " + trip.Route.DestinationCity
			}
			rp = &RouteProfitability{
				RouteID:   trip.RouteID,
				RouteName: routeName,
			}
			routeMap[trip.RouteID] = rp
		}

		rp.TripsCount++
		rp.TotalPassengers += a.TotalPassengers
		rp.AvgOccupancy += a.AvgOccupancyRate
		rp.Revenue += a.Revenue
		rp.Costs += a.FuelCost + a.DriverCost + a.OtherCosts
		rp.Profit += a.Profit
	}

	// Обчислюємо середні та категорії
	if summary.TotalTrips > 0 {
		summary.AvgOccupancy /= float64(summary.TotalTrips)
		if summary.TotalCosts > 0 {
			summary.AverageProfitability = (summary.TotalProfit / summary.TotalCosts) * 100
		}
	}

	byRoute := make([]RouteProfitability, 0, len(routeMap))
	for _, rp := range routeMap {
		if rp.TripsCount > 0 {
			rp.AvgOccupancy /= float64(rp.TripsCount)
			if rp.Costs > 0 {
				rp.Profitability = (rp.Profit / rp.Costs) * 100
			}
			rp.Category = categorizeProfitability(rp.Profitability)
		}
		byRoute = append(byRoute, *rp)
	}

	return &ProfitabilityData{
		Period: PeriodInfo{
			From: from.Format("2006-01-02"),
			To:   to.Format("2006-01-02"),
		},
		Summary: summary,
		ByRoute: byRoute,
	}, nil
}

// CalculateTripAnalytics розраховує аналітику для рейсу
func (s *analyticsService) CalculateTripAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error) {
	return s.analyticsRepo.CalculateTripAnalytics(ctx, tripID)
}

// GetTripAnalytics повертає аналітику рейсу
func (s *analyticsService) GetTripAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error) {
	return s.analyticsRepo.GetTripAnalytics(ctx, tripID)
}

// categorizeProfitability визначає категорію рентабельності
func categorizeProfitability(profitability float64) string {
	switch {
	case profitability < 0:
		return "unprofitable"
	case profitability < 20:
		return "low_profit"
	case profitability < 50:
		return "normal"
	default:
		return "high_profit"
	}
}

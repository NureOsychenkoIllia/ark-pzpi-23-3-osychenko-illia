package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"busoptima/internal/model"
)

// AnalyticsRepository інтерфейс для роботи з аналітикою
type AnalyticsRepository interface {
	CalculateTripAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error)
	GetTripAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error)
	GetProfitabilityByRoute(ctx context.Context, routeID int64, from, to time.Time) ([]model.TripAnalytics, error)
	GetHistoricalPassengers(ctx context.Context, routeID int64, dayOfWeek int, weeks int) ([]int, error)
	SaveDemandForecast(ctx context.Context, forecast *model.DemandForecast) error
	GetDemandForecasts(ctx context.Context, routeID int64, from, to time.Time) ([]model.DemandForecast, error)
}

// analyticsRepository реалізація AnalyticsRepository
type analyticsRepository struct {
	db *sqlx.DB
}

// NewAnalyticsRepository створює новий екземпляр репозиторію аналітики
func NewAnalyticsRepository(db *sqlx.DB) AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// CalculateTripAnalytics розраховує аналітику для рейсу
func (r *analyticsRepository) CalculateTripAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error) {
	// Отримуємо дані рейсу, маршруту та автобуса
	var tripData struct {
		TripID              int64   `db:"trip_id"`
		RouteID             int64   `db:"route_id"`
		BusCapacity         int     `db:"bus_capacity"`
		DistanceKm          float64 `db:"distance_km"`
		FuelCostPerKm       float64 `db:"fuel_cost_per_km"`
		DriverCostPerTrip   float64 `db:"driver_cost_per_trip"`
		FuelConsumption     float64 `db:"fuel_consumption_per_100km"`
		TotalPassengers     int     `db:"total_passengers"`
		MaxPassengers       int     `db:"max_passengers"`
		Revenue             float64 `db:"revenue"`
	}

	query := `
		SELECT 
			t.id as trip_id,
			t.route_id,
			b.capacity as bus_capacity,
			r.distance_km,
			r.fuel_cost_per_km,
			r.driver_cost_per_trip,
			b.fuel_consumption_per_100km,
			COALESCE(
				(SELECT COUNT(DISTINCT device_local_id) 
				 FROM passenger_events pe 
				 WHERE pe.trip_id = t.id AND pe.event_type = 'entry'), 0
			) as total_passengers,
			COALESCE(
				(SELECT MAX(passenger_count_after) 
				 FROM passenger_events pe 
				 WHERE pe.trip_id = t.id), 0
			) as max_passengers,
			COALESCE(
				(SELECT SUM(pr.recommended_price)
				 FROM price_recommendations pr
				 WHERE pr.trip_id = t.id), 
				r.base_price * COALESCE(
					(SELECT COUNT(DISTINCT device_local_id) 
					 FROM passenger_events pe 
					 WHERE pe.trip_id = t.id AND pe.event_type = 'entry'), 0
				)
			) as revenue
		FROM trips t
		JOIN routes r ON t.route_id = r.id
		JOIN buses b ON t.bus_id = b.id
		WHERE t.id = $1`

	err := r.db.GetContext(ctx, &tripData, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip data: %w", err)
	}

	// Розраховуємо витрати
	fuelCost := tripData.DistanceKm * tripData.FuelConsumption / 100 * tripData.FuelCostPerKm
	driverCost := tripData.DriverCostPerTrip
	otherCosts := 0.0 // Інші витрати (амортизація, страхування тощо)

	// Розраховуємо прибуток та рентабельність
	totalCosts := fuelCost + driverCost + otherCosts
	profit := tripData.Revenue - totalCosts
	profitabilityPercent := 0.0
	if totalCosts > 0 {
		profitabilityPercent = (profit / totalCosts) * 100
	}

	// Розраховуємо середню завантаженість
	avgOccupancyRate := 0.0
	if tripData.BusCapacity > 0 {
		avgOccupancyRate = (float64(tripData.MaxPassengers) / float64(tripData.BusCapacity)) * 100
	}

	analytics := &model.TripAnalytics{
		TripID:               tripID,
		TotalPassengers:      tripData.TotalPassengers,
		MaxPassengers:        tripData.MaxPassengers,
		AvgOccupancyRate:     avgOccupancyRate,
		Revenue:              tripData.Revenue,
		FuelCost:             fuelCost,
		DriverCost:           driverCost,
		OtherCosts:           otherCosts,
		Profit:               profit,
		ProfitabilityPercent: profitabilityPercent,
		CalculatedAt:         time.Now(),
	}

	// Зберігаємо результати в БД
	insertQuery := `
		INSERT INTO trip_analytics (
			trip_id, total_passengers, max_passengers, avg_occupancy_rate,
			revenue, fuel_cost, driver_cost, other_costs, profit, profitability_percent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (trip_id) DO UPDATE SET
			total_passengers = EXCLUDED.total_passengers,
			max_passengers = EXCLUDED.max_passengers,
			avg_occupancy_rate = EXCLUDED.avg_occupancy_rate,
			revenue = EXCLUDED.revenue,
			fuel_cost = EXCLUDED.fuel_cost,
			driver_cost = EXCLUDED.driver_cost,
			other_costs = EXCLUDED.other_costs,
			profit = EXCLUDED.profit,
			profitability_percent = EXCLUDED.profitability_percent,
			calculated_at = CURRENT_TIMESTAMP
		RETURNING id, calculated_at`

	err = r.db.QueryRowContext(ctx, insertQuery,
		analytics.TripID, analytics.TotalPassengers, analytics.MaxPassengers,
		analytics.AvgOccupancyRate, analytics.Revenue, analytics.FuelCost,
		analytics.DriverCost, analytics.OtherCosts, analytics.Profit,
		analytics.ProfitabilityPercent,
	).Scan(&analytics.ID, &analytics.CalculatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to save trip analytics: %w", err)
	}

	return analytics, nil
}

// GetTripAnalytics повертає аналітику рейсу
func (r *analyticsRepository) GetTripAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error) {
	var analytics model.TripAnalytics
	query := `SELECT * FROM trip_analytics WHERE trip_id = $1`

	err := r.db.GetContext(ctx, &analytics, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip analytics: %w", err)
	}

	return &analytics, nil
}

// GetProfitabilityByRoute повертає аналітику рентабельності за маршрутом
func (r *analyticsRepository) GetProfitabilityByRoute(ctx context.Context, routeID int64, from, to time.Time) ([]model.TripAnalytics, error) {
	var analytics []model.TripAnalytics
	query := `
		SELECT ta.* 
		FROM trip_analytics ta
		JOIN trips t ON ta.trip_id = t.id
		WHERE t.route_id = $1 
		AND t.scheduled_departure BETWEEN $2 AND $3
		ORDER BY t.scheduled_departure DESC`

	err := r.db.SelectContext(ctx, &analytics, query, routeID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get profitability data: %w", err)
	}

	return analytics, nil
}

// GetHistoricalPassengers повертає історичні дані про пасажирів
func (r *analyticsRepository) GetHistoricalPassengers(ctx context.Context, routeID int64, dayOfWeek int, weeks int) ([]int, error) {
	var passengers []int
	query := `
		SELECT COALESCE(ta.total_passengers, 0) as passengers
		FROM generate_series(
			CURRENT_DATE - INTERVAL '%d weeks',
			CURRENT_DATE - INTERVAL '1 week',
			INTERVAL '1 week'
		) AS week_start
		LEFT JOIN trips t ON t.route_id = $1 
			AND DATE_PART('dow', t.scheduled_departure) = $2
			AND t.scheduled_departure >= week_start 
			AND t.scheduled_departure < week_start + INTERVAL '1 week'
		LEFT JOIN trip_analytics ta ON ta.trip_id = t.id
		ORDER BY week_start DESC
		LIMIT $3`

	formattedQuery := fmt.Sprintf(query, weeks)
	err := r.db.SelectContext(ctx, &passengers, formattedQuery, routeID, dayOfWeek, weeks)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical passengers: %w", err)
	}

	return passengers, nil
}

// SaveDemandForecast зберігає прогноз попиту
func (r *analyticsRepository) SaveDemandForecast(ctx context.Context, forecast *model.DemandForecast) error {
	query := `
		INSERT INTO demand_forecasts (
			route_id, forecast_date, day_of_week, predicted_passengers,
			confidence_lower, confidence_upper
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (route_id, forecast_date) DO UPDATE SET
			predicted_passengers = EXCLUDED.predicted_passengers,
			confidence_lower = EXCLUDED.confidence_lower,
			confidence_upper = EXCLUDED.confidence_upper,
			created_at = CURRENT_TIMESTAMP
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		forecast.RouteID, forecast.ForecastDate, forecast.DayOfWeek,
		forecast.PredictedPassengers, forecast.ConfidenceLower, forecast.ConfidenceUpper,
	).Scan(&forecast.ID, &forecast.CreatedAt)
}

// GetDemandForecasts повертає прогнози попиту
func (r *analyticsRepository) GetDemandForecasts(ctx context.Context, routeID int64, from, to time.Time) ([]model.DemandForecast, error) {
	var forecasts []model.DemandForecast
	query := `
		SELECT * FROM demand_forecasts 
		WHERE route_id = $1 
		AND forecast_date BETWEEN $2 AND $3
		ORDER BY forecast_date`

	err := r.db.SelectContext(ctx, &forecasts, query, routeID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get demand forecasts: %w", err)
	}

	return forecasts, nil
}
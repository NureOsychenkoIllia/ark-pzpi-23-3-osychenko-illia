package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"busoptima/internal/model"
)

// RouteRepository інтерфейс для роботи з маршрутами
type RouteRepository interface {
	Create(ctx context.Context, route *model.Route) error
	GetByID(ctx context.Context, id int64) (*model.Route, error)
	GetAll(ctx context.Context, activeOnly bool) ([]model.Route, error)
	Update(ctx context.Context, route *model.Route) error
	Delete(ctx context.Context, id int64) error
}

// routeRepository реалізація RouteRepository
type routeRepository struct {
	db *sqlx.DB
}

// NewRouteRepository створює новий екземпляр репозиторію маршрутів
func NewRouteRepository(db *sqlx.DB) RouteRepository {
	return &routeRepository{db: db}
}

// Create додає новий маршрут до бази даних
func (r *routeRepository) Create(ctx context.Context, route *model.Route) error {
	query := `
		INSERT INTO routes (origin_city, destination_city, distance_km, 
			base_price, fuel_cost_per_km, driver_cost_per_trip, 
			estimated_duration_minutes, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`
	
	return r.db.QueryRowContext(ctx, query,
		route.OriginCity, route.DestinationCity, route.DistanceKm,
		route.BasePrice, route.FuelCostPerKm, route.DriverCostPerTrip,
		route.EstimatedDurationMin, route.IsActive,
	).Scan(&route.ID, &route.CreatedAt, &route.UpdatedAt)
}

// GetByID повертає маршрут за його ідентифікатором
func (r *routeRepository) GetByID(ctx context.Context, id int64) (*model.Route, error) {
	var route model.Route
	query := `SELECT * FROM routes WHERE id = $1`
	
	err := r.db.GetContext(ctx, &route, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("route with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	
	return &route, nil
}

// GetAll повертає список маршрутів з можливістю фільтрації
func (r *routeRepository) GetAll(ctx context.Context, activeOnly bool) ([]model.Route, error) {
	var routes []model.Route
	query := `SELECT * FROM routes`
	
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY origin_city, destination_city`
	
	err := r.db.SelectContext(ctx, &routes, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}
	
	return routes, nil
}

// Update оновлює існуючий маршрут
func (r *routeRepository) Update(ctx context.Context, route *model.Route) error {
	query := `
		UPDATE routes SET 
			origin_city = $1, destination_city = $2, distance_km = $3,
			base_price = $4, fuel_cost_per_km = $5, driver_cost_per_trip = $6,
			estimated_duration_minutes = $7, is_active = $8,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $9
		RETURNING updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		route.OriginCity, route.DestinationCity, route.DistanceKm,
		route.BasePrice, route.FuelCostPerKm, route.DriverCostPerTrip,
		route.EstimatedDurationMin, route.IsActive, route.ID,
	).Scan(&route.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("route with id %d not found", route.ID)
		}
		return fmt.Errorf("failed to update route: %w", err)
	}
	
	return nil
}

// Delete видаляє маршрут (soft delete)
func (r *routeRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE routes SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete route: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("route with id %d not found", id)
	}
	
	return nil
}
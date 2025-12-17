package repository

import (
	"busoptima/internal/model"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TripRepository інтерфейс для роботи з рейсами
type TripRepository interface {
	Create(ctx context.Context, trip *model.Trip) error
	GetByID(ctx context.Context, id int64) (*model.Trip, error)
	GetAll(ctx context.Context, filters map[string]interface{}) ([]model.Trip, error)
	Update(ctx context.Context, trip *model.Trip) error
	UpdatePassengerCount(ctx context.Context, tripID int64, count int) error
}

// tripRepository реалізація TripRepository
type tripRepository struct {
	db *sqlx.DB
}

// NewTripRepository створює новий екземпляр репозиторію рейсів
func NewTripRepository(db *sqlx.DB) TripRepository {
	return &tripRepository{db: db}
}

// Create додає новий рейс до бази даних
func (r *tripRepository) Create(ctx context.Context, trip *model.Trip) error {
	query := `
		INSERT INTO trips (route_id, bus_id, scheduled_departure, actual_departure, 
			actual_arrival, status, current_passengers, driver_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	return r.db.QueryRowContext(ctx, query,
		trip.RouteID, trip.BusID, trip.ScheduledDeparture, trip.ActualDeparture,
		trip.ActualArrival, trip.Status, trip.CurrentPassengers, trip.DriverName,
	).Scan(&trip.ID)
}

// GetByID повертає рейс за його ідентифікатором з даними маршруту та автобуса
func (r *tripRepository) GetByID(ctx context.Context, id int64) (*model.Trip, error) {
	var trip model.Trip
	query := `
		SELECT t.id, t.route_id, t.bus_id, t.scheduled_departure, 
			t.actual_departure, t.actual_arrival, t.status, 
			t.current_passengers, t.driver_name,
			r.origin_city, r.destination_city, r.distance_km, r.base_price,
			b.registration_number, b.capacity, b.model
		FROM trips t
		LEFT JOIN routes r ON t.route_id = r.id
		LEFT JOIN buses b ON t.bus_id = b.id
		WHERE t.id = $1`

	row := r.db.QueryRowContext(ctx, query, id)

	// Use nullable types for LEFT JOIN fields
	var routeOriginCity, routeDestinationCity *string
	var routeDistanceKm, routeBasePrice *float64
	var busRegistrationNumber, busModel *string
	var busCapacity *int

	err := row.Scan(
		&trip.ID, &trip.RouteID, &trip.BusID, &trip.ScheduledDeparture,
		&trip.ActualDeparture, &trip.ActualArrival, &trip.Status,
		&trip.CurrentPassengers, &trip.DriverName,
		&routeOriginCity, &routeDestinationCity, &routeDistanceKm, &routeBasePrice,
		&busRegistrationNumber, &busCapacity, &busModel,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get trip: %w", err)
	}

	if routeOriginCity != nil {
		route := model.Route{
			ID:              trip.RouteID,
			OriginCity:      *routeOriginCity,
			DestinationCity: *routeDestinationCity,
			DistanceKm:      *routeDistanceKm,
			BasePrice:       *routeBasePrice,
		}
		trip.Route = &route
	}

	if busRegistrationNumber != nil {
		bus := model.Bus{
			ID:                 trip.BusID,
			RegistrationNumber: *busRegistrationNumber,
			Capacity:           *busCapacity,
			Model:              *busModel,
		}
		trip.Bus = &bus
	}

	return &trip, nil
}

// GetAll повертає список рейсів з фільтрами
func (r *tripRepository) GetAll(ctx context.Context, filters map[string]interface{}) ([]model.Trip, error) {
	var trips []model.Trip
	query := `
		SELECT t.id, t.route_id, t.bus_id, t.scheduled_departure, 
			t.actual_departure, t.actual_arrival, t.status, 
			t.current_passengers, t.driver_name,
			r.origin_city, r.destination_city, r.distance_km, r.base_price,
			b.registration_number, b.capacity, b.model
		FROM trips t
		LEFT JOIN routes r ON t.route_id = r.id
		LEFT JOIN buses b ON t.bus_id = b.id
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	// Додаємо фільтри
	if routeID, ok := filters["route_id"]; ok {
		query += fmt.Sprintf(" AND t.route_id = $%d", argIndex)
		args = append(args, routeID)
		argIndex++
	}

	if status, ok := filters["status"]; ok {
		query += fmt.Sprintf(" AND t.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	if dateFrom, ok := filters["date_from"]; ok {
		query += fmt.Sprintf(" AND t.scheduled_departure >= $%d", argIndex)
		args = append(args, dateFrom)
		argIndex++
	}

	if dateTo, ok := filters["date_to"]; ok {
		query += fmt.Sprintf(" AND t.scheduled_departure <= $%d", argIndex)
		args = append(args, dateTo)
		argIndex++
	}

	query += " ORDER BY t.scheduled_departure DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get trips: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var trip model.Trip

		// Use nullable types for LEFT JOIN fields
		var routeOriginCity, routeDestinationCity *string
		var routeDistanceKm, routeBasePrice *float64
		var busRegistrationNumber, busModel *string
		var busCapacity *int

		err := rows.Scan(
			&trip.ID, &trip.RouteID, &trip.BusID, &trip.ScheduledDeparture,
			&trip.ActualDeparture, &trip.ActualArrival, &trip.Status,
			&trip.CurrentPassengers, &trip.DriverName,
			&routeOriginCity, &routeDestinationCity, &routeDistanceKm, &routeBasePrice,
			&busRegistrationNumber, &busCapacity, &busModel,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan trip: %w", err)
		}

		// Only set route if we have route data
		if routeOriginCity != nil {
			route := model.Route{
				ID:              trip.RouteID,
				OriginCity:      *routeOriginCity,
				DestinationCity: *routeDestinationCity,
				DistanceKm:      *routeDistanceKm,
				BasePrice:       *routeBasePrice,
			}
			trip.Route = &route
		}

		// Only set bus if we have bus data
		if busRegistrationNumber != nil {
			bus := model.Bus{
				ID:                 trip.BusID,
				RegistrationNumber: *busRegistrationNumber,
				Capacity:           *busCapacity,
				Model:              *busModel,
			}
			trip.Bus = &bus
		}

		trips = append(trips, trip)
	}

	return trips, nil
}

// Update оновлює існуючий рейс
func (r *tripRepository) Update(ctx context.Context, trip *model.Trip) error {
	query := `
		UPDATE trips SET 
			route_id = $1, bus_id = $2, scheduled_departure = $3,
			actual_departure = $4, actual_arrival = $5, status = $6,
			current_passengers = $7, driver_name = $8
		WHERE id = $9`

	result, err := r.db.ExecContext(ctx, query,
		trip.RouteID, trip.BusID, trip.ScheduledDeparture,
		trip.ActualDeparture, trip.ActualArrival, trip.Status,
		trip.CurrentPassengers, trip.DriverName, trip.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update trip: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("trip with id %d not found", trip.ID)
	}

	return nil
}

// UpdatePassengerCount оновлює кількість пасажирів у рейсі
func (r *tripRepository) UpdatePassengerCount(ctx context.Context, tripID int64, count int) error {
	query := `UPDATE trips SET current_passengers = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, count, tripID)
	if err != nil {
		return fmt.Errorf("failed to update passenger count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("trip with id %d not found", tripID)
	}

	return nil
}

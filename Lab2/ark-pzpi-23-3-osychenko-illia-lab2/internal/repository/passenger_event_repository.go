package repository

import (
	"context"
	"fmt"
	"busoptima/internal/model"
	"github.com/jmoiron/sqlx"
)

// PassengerEventRepository інтерфейс для роботи з подіями пасажирів
type PassengerEventRepository interface {
	BatchCreate(ctx context.Context, events []model.PassengerEvent) error
	GetByTripID(ctx context.Context, tripID int64) ([]model.PassengerEvent, error)
}

// passengerEventRepository реалізація PassengerEventRepository
type passengerEventRepository struct {
	db *sqlx.DB
}

// NewPassengerEventRepository створює новий екземпляр репозиторію подій
func NewPassengerEventRepository(db *sqlx.DB) PassengerEventRepository {
	return &passengerEventRepository{db: db}
}

// BatchCreate створює пакет подій через транзакцію
func (r *passengerEventRepository) BatchCreate(ctx context.Context, events []model.PassengerEvent) error {
	if len(events) == 0 {
		return nil
	}
	
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	query := `
		INSERT INTO passenger_events (trip_id, event_type, timestamp, latitude, 
			longitude, passenger_count_after, device_local_id, is_synced)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	for _, event := range events {
		_, err := tx.ExecContext(ctx, query,
			event.TripID, event.EventType, event.Timestamp, event.Latitude,
			event.Longitude, event.PassengerCountAfter, event.DeviceLocalID, true,
		)
		if err != nil {
			return fmt.Errorf("failed to insert event: %w", err)
		}
	}
	
	// Оновлюємо поточну кількість пасажирів у рейсі
	if len(events) > 0 {
		lastEvent := events[len(events)-1]
		updateQuery := `UPDATE trips SET current_passengers = $1 WHERE id = $2`
		_, err := tx.ExecContext(ctx, updateQuery, lastEvent.PassengerCountAfter, lastEvent.TripID)
		if err != nil {
			return fmt.Errorf("failed to update trip passenger count: %w", err)
		}
	}
	
	return tx.Commit()
}

// GetByTripID повертає всі події для конкретного рейсу
func (r *passengerEventRepository) GetByTripID(ctx context.Context, tripID int64) ([]model.PassengerEvent, error) {
	var events []model.PassengerEvent
	query := `
		SELECT * FROM passenger_events 
		WHERE trip_id = $1 
		ORDER BY timestamp ASC`
	
	err := r.db.SelectContext(ctx, &events, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get passenger events: %w", err)
	}
	
	return events, nil
}
package repository

import (
	"context"
	"fmt"
	"busoptima/internal/model"
	"github.com/jmoiron/sqlx"
)

// BusRepository інтерфейс для роботи з автобусами
type BusRepository interface {
	Create(ctx context.Context, bus *model.Bus) error
	GetByID(ctx context.Context, id int64) (*model.Bus, error)
	GetAll(ctx context.Context, activeOnly bool) ([]model.Bus, error)
	Update(ctx context.Context, bus *model.Bus) error
	Delete(ctx context.Context, id int64) error
}

// busRepository реалізація BusRepository
type busRepository struct {
	db *sqlx.DB
}

// NewBusRepository створює новий екземпляр репозиторію автобусів
func NewBusRepository(db *sqlx.DB) BusRepository {
	return &busRepository{db: db}
}

// Create додає новий автобус до бази даних
func (r *busRepository) Create(ctx context.Context, bus *model.Bus) error {
	query := `
		INSERT INTO buses (registration_number, capacity, model, fuel_consumption_per_100km, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	
	return r.db.QueryRowContext(ctx, query,
		bus.RegistrationNumber, bus.Capacity, bus.Model, 
		bus.FuelConsumptionPer100km, bus.IsActive,
	).Scan(&bus.ID)
}

// GetByID повертає автобус за його ідентифікатором
func (r *busRepository) GetByID(ctx context.Context, id int64) (*model.Bus, error) {
	var bus model.Bus
	query := `SELECT * FROM buses WHERE id = $1`
	
	err := r.db.GetContext(ctx, &bus, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get bus: %w", err)
	}
	
	return &bus, nil
}

// GetAll повертає список автобусів з можливістю фільтрації
func (r *busRepository) GetAll(ctx context.Context, activeOnly bool) ([]model.Bus, error) {
	var buses []model.Bus
	query := `SELECT * FROM buses`
	
	if activeOnly {
		query += ` WHERE is_active = true`
	}
	query += ` ORDER BY registration_number`
	
	err := r.db.SelectContext(ctx, &buses, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get buses: %w", err)
	}
	
	return buses, nil
}

// Update оновлює існуючий автобус
func (r *busRepository) Update(ctx context.Context, bus *model.Bus) error {
	query := `
		UPDATE buses SET 
			registration_number = $1, capacity = $2, model = $3,
			fuel_consumption_per_100km = $4, is_active = $5
		WHERE id = $6`
	
	result, err := r.db.ExecContext(ctx, query,
		bus.RegistrationNumber, bus.Capacity, bus.Model,
		bus.FuelConsumptionPer100km, bus.IsActive, bus.ID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update bus: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("bus with id %d not found", bus.ID)
	}
	
	return nil
}

// Delete видаляє автобус (soft delete)
func (r *busRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE buses SET is_active = false WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete bus: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("bus with id %d not found", id)
	}
	
	return nil
}
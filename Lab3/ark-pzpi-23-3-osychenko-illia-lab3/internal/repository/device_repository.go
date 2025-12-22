package repository

import (
	"busoptima/internal/model"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DeviceRepository інтерфейс для роботи з IoT-пристроями
type DeviceRepository interface {
	GetBySerialNumber(ctx context.Context, serialNumber string) (*model.Device, error)
	UpdateLastSync(ctx context.Context, deviceID int64) error
}

// deviceRepository реалізація DeviceRepository
type deviceRepository struct {
	db *sqlx.DB
}

// NewDeviceRepository створює новий екземпляр репозиторію пристроїв
func NewDeviceRepository(db *sqlx.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

// GetBySerialNumber повертає пристрій за серійним номером
func (r *deviceRepository) GetBySerialNumber(ctx context.Context, serialNumber string) (*model.Device, error) {
	var device model.Device
	query := `
		SELECT d.id, d.serial_number, d.auth_token_hash, d.bus_id, 
			d.firmware_version, d.last_sync_at, d.is_active,
			b.registration_number, b.capacity, b.model, b.fuel_consumption_per_100km
		FROM devices d
		LEFT JOIN buses b ON d.bus_id = b.id
		WHERE d.serial_number = $1 AND d.is_active = true`

	row := r.db.QueryRowContext(ctx, query, serialNumber)

	// Use nullable types for LEFT JOIN fields
	var busRegistrationNumber, busModel *string
	var busCapacity *int
	var busFuelConsumption *float64

	err := row.Scan(
		&device.ID, &device.SerialNumber, &device.AuthTokenHash, &device.BusID,
		&device.FirmwareVersion, &device.LastSyncAt, &device.IsActive,
		&busRegistrationNumber, &busCapacity, &busModel, &busFuelConsumption,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// Only set bus if we have bus data
	if device.BusID != nil && busRegistrationNumber != nil {
		bus := model.Bus{
			ID:                      *device.BusID,
			RegistrationNumber:      *busRegistrationNumber,
			Capacity:                *busCapacity,
			Model:                   *busModel,
			FuelConsumptionPer100km: *busFuelConsumption,
		}
		device.Bus = &bus
	}

	return &device, nil
}

// UpdateLastSync оновлює час останньої синхронізації пристрою
func (r *deviceRepository) UpdateLastSync(ctx context.Context, deviceID int64) error {
	query := `UPDATE devices SET last_sync_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to update last sync: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("device with id %d not found", deviceID)
	}

	return nil
}

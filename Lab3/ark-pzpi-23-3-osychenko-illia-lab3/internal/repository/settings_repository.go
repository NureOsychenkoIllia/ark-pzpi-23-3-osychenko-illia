package repository

import (
	"busoptima/internal/model"
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// SettingsRepository інтерфейс для роботи з системними налаштуваннями
type SettingsRepository interface {
	GetSettings(ctx context.Context) (*model.SystemSettings, error)
	UpdateSettings(ctx context.Context, settings *model.SystemSettings) error
}

type settingsRepository struct {
	db *sqlx.DB
}

// NewSettingsRepository створює новий репозиторій налаштувань
func NewSettingsRepository(db *sqlx.DB) SettingsRepository {
	return &settingsRepository{db: db}
}

// GetSettings отримує поточні системні налаштування
func (r *settingsRepository) GetSettings(ctx context.Context) (*model.SystemSettings, error) {
	var settings model.SystemSettings
	var seasonalCoeffsJSON []byte

	query := `
		SELECT id, fuel_price_per_liter, peak_hours_coefficient, weekend_coefficient,
			   high_demand_threshold, low_demand_threshold, price_min_coefficient,
			   price_max_coefficient, seasonal_coefficients, updated_at, updated_by
		FROM system_settings 
		ORDER BY updated_at DESC 
		LIMIT 1`

	err := r.db.QueryRowxContext(ctx, query).Scan(
		&settings.ID,
		&settings.FuelPricePerLiter,
		&settings.PeakHoursCoefficient,
		&settings.WeekendCoefficient,
		&settings.HighDemandThreshold,
		&settings.LowDemandThreshold,
		&settings.PriceMinCoefficient,
		&settings.PriceMaxCoefficient,
		&seasonalCoeffsJSON,
		&settings.UpdatedAt,
		&settings.UpdatedBy,
	)

	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil // Налаштування не знайдено
		}
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Розпарсити JSON для сезонних коефіцієнтів
	if err := json.Unmarshal(seasonalCoeffsJSON, &settings.SeasonalCoefficients); err != nil {
		return nil, fmt.Errorf("failed to unmarshal seasonal coefficients: %w", err)
	}

	return &settings, nil
}

// UpdateSettings оновлює системні налаштування
func (r *settingsRepository) UpdateSettings(ctx context.Context, settings *model.SystemSettings) error {
	seasonalCoeffsJSON, err := json.Marshal(settings.SeasonalCoefficients)
	if err != nil {
		return fmt.Errorf("failed to marshal seasonal coefficients: %w", err)
	}

	settings.UpdatedAt = time.Now()

	query := `
		INSERT INTO system_settings (
			fuel_price_per_liter, peak_hours_coefficient, weekend_coefficient,
			high_demand_threshold, low_demand_threshold, price_min_coefficient,
			price_max_coefficient, seasonal_coefficients, updated_at, updated_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			fuel_price_per_liter = EXCLUDED.fuel_price_per_liter,
			peak_hours_coefficient = EXCLUDED.peak_hours_coefficient,
			weekend_coefficient = EXCLUDED.weekend_coefficient,
			high_demand_threshold = EXCLUDED.high_demand_threshold,
			low_demand_threshold = EXCLUDED.low_demand_threshold,
			price_min_coefficient = EXCLUDED.price_min_coefficient,
			price_max_coefficient = EXCLUDED.price_max_coefficient,
			seasonal_coefficients = EXCLUDED.seasonal_coefficients,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
		RETURNING id, updated_at`

	err = r.db.QueryRowxContext(ctx, query,
		settings.FuelPricePerLiter,
		settings.PeakHoursCoefficient,
		settings.WeekendCoefficient,
		settings.HighDemandThreshold,
		settings.LowDemandThreshold,
		settings.PriceMinCoefficient,
		settings.PriceMaxCoefficient,
		seasonalCoeffsJSON,
		settings.UpdatedAt,
		settings.UpdatedBy,
	).Scan(&settings.ID, &settings.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}

// JSONMap тип для роботи з JSON полями в PostgreSQL
type JSONMap map[string]interface{}

// Value реалізує driver.Valuer інтерфейс
func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan реалізує sql.Scanner інтерфейс
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

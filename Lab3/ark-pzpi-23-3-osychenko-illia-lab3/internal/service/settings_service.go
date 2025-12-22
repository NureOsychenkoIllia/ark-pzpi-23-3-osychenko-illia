package service

import (
	"busoptima/internal/model"
	"busoptima/internal/repository"
	"context"
	"fmt"
)

// SettingsService інтерфейс для роботи з системними налаштуваннями
type SettingsService interface {
	GetSettings(ctx context.Context) (*model.SystemSettings, error)
	UpdateSettings(ctx context.Context, settings *model.SystemSettings, userID int64) error
	ValidateSettings(settings *model.SystemSettings) error
}

type settingsService struct {
	settingsRepo repository.SettingsRepository
}

// NewSettingsService створює новий сервіс налаштувань
func NewSettingsService(settingsRepo repository.SettingsRepository) SettingsService {
	return &settingsService{
		settingsRepo: settingsRepo,
	}
}

// GetSettings повертає поточні системні налаштування
func (s *settingsService) GetSettings(ctx context.Context) (*model.SystemSettings, error) {
	settings, err := s.settingsRepo.GetSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	// Якщо налаштування не знайдено, повертаємо значення за замовчуванням
	if settings == nil {
		settings = s.getDefaultSettings()
	}

	return settings, nil
}

// UpdateSettings оновлює системні налаштування
func (s *settingsService) UpdateSettings(ctx context.Context, settings *model.SystemSettings, userID int64) error {
	if err := s.ValidateSettings(settings); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	settings.UpdatedBy = &userID

	if err := s.settingsRepo.UpdateSettings(ctx, settings); err != nil {
		return fmt.Errorf("failed to update settings: %w", err)
	}

	return nil
}

// ValidateSettings валідує системні налаштування
func (s *settingsService) ValidateSettings(settings *model.SystemSettings) error {
	if settings.FuelPricePerLiter < 10 || settings.FuelPricePerLiter > 200 {
		return fmt.Errorf("fuel price must be between 10-200 UAH")
	}

	if settings.PeakHoursCoefficient < 0.5 || settings.PeakHoursCoefficient > 3.0 {
		return fmt.Errorf("peak hours coefficient must be between 0.5-3.0")
	}

	if settings.WeekendCoefficient < 0.5 || settings.WeekendCoefficient > 3.0 {
		return fmt.Errorf("weekend coefficient must be between 0.5-3.0")
	}

	if settings.HighDemandThreshold <= settings.LowDemandThreshold {
		return fmt.Errorf("high demand threshold must be greater than low demand threshold")
	}

	if settings.LowDemandThreshold < 0 || settings.LowDemandThreshold > 100 {
		return fmt.Errorf("low demand threshold must be between 0-100")
	}

	if settings.HighDemandThreshold < 0 || settings.HighDemandThreshold > 100 {
		return fmt.Errorf("high demand threshold must be between 0-100")
	}

	if settings.PriceMinCoefficient < 0.1 || settings.PriceMinCoefficient > 1.0 {
		return fmt.Errorf("price min coefficient must be between 0.1-1.0")
	}

	if settings.PriceMaxCoefficient < 1.0 || settings.PriceMaxCoefficient > 5.0 {
		return fmt.Errorf("price max coefficient must be between 1.0-5.0")
	}

	if settings.PriceMinCoefficient >= settings.PriceMaxCoefficient {
		return fmt.Errorf("price min coefficient must be less than price max coefficient")
	}

	// Валідація сезонних коефіцієнтів
	for season, coeff := range settings.SeasonalCoefficients {
		if coeff < 0.5 || coeff > 3.0 {
			return fmt.Errorf("seasonal coefficient for %s must be between 0.5-3.0", season)
		}
	}

	return nil
}

// getDefaultSettings повертає налаштування за замовчуванням
func (s *settingsService) getDefaultSettings() *model.SystemSettings {
	return &model.SystemSettings{
		FuelPricePerLiter:    50.00,
		PeakHoursCoefficient: 1.20,
		WeekendCoefficient:   1.15,
		HighDemandThreshold:  85,
		LowDemandThreshold:   30,
		PriceMinCoefficient:  0.70,
		PriceMaxCoefficient:  1.50,
		SeasonalCoefficients: map[string]float64{
			"new_year": 1.30,
			"summer":   1.15,
			"regular":  1.00,
		},
	}
}

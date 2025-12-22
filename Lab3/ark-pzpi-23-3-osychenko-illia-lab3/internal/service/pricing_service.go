package service

import (
	"context"
	"math"
	"time"
)

// PricingService інтерфейс для динамічного ціноутворення
type PricingService interface {
	CalculatePrice(ctx context.Context, basePrice float64, currentPassengers, capacity int, departureTime time.Time) (*PriceRecommendation, error)
	CalculatePriceWithCoefficients(basePrice, demandCoeff, timeCoeff, dayCoeff, minCoeff, maxCoeff float64) float64
}

// PriceRecommendation структура рекомендації ціни
type PriceRecommendation struct {
	BasePrice        float64 `json:"base_price"`
	RecommendedPrice float64 `json:"recommended_price"`
	OccupancyRate    float64 `json:"occupancy_rate"`
	DemandCoeff      float64 `json:"demand_coefficient"`
	TimeCoeff        float64 `json:"time_coefficient"`
	DayCoeff         float64 `json:"day_coefficient"`
	PriceChange      float64 `json:"price_change"`
	PriceChangePerc  float64 `json:"price_change_percent"`
	Category         string  `json:"category"`
	Recommendation   string  `json:"recommendation"`
}

type pricingService struct {
	settingsService SettingsService
}

// NewPricingService створює новий сервіс ціноутворення
func NewPricingService(settingsService SettingsService) PricingService {
	return &pricingService{
		settingsService: settingsService,
	}
}

// CalculatePrice розраховує рекомендовану ціну на основі завантаженості та часу
func (s *pricingService) CalculatePrice(ctx context.Context, basePrice float64, currentPassengers, capacity int, departureTime time.Time) (*PriceRecommendation, error) {
	// Отримуємо поточні системні налаштування
	settings, err := s.settingsService.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	// Розрахунок завантаженості
	occupancyRate := 0.0
	if capacity > 0 {
		occupancyRate = float64(currentPassengers) / float64(capacity) * 100
	}

	// Коефіцієнт попиту на основі завантаженості та налаштувань
	demandCoeff := s.calculateDemandCoefficient(occupancyRate, settings.LowDemandThreshold, settings.HighDemandThreshold)

	// Коефіцієнт часу (пікові години)
	timeCoeff := s.calculateTimeCoefficient(departureTime, settings.PeakHoursCoefficient)

	// Коефіцієнт дня (вихідні/будні)
	dayCoeff := s.calculateDayCoefficient(departureTime, settings.WeekendCoefficient, settings.SeasonalCoefficients)

	// Розрахунок рекомендованої ціни
	recommendedPrice := s.CalculatePriceWithCoefficients(basePrice, demandCoeff, timeCoeff, dayCoeff, settings.PriceMinCoefficient, settings.PriceMaxCoefficient)

	// Розрахунок зміни ціни
	priceChange := recommendedPrice - basePrice
	priceChangePerc := 0.0
	if basePrice > 0 {
		priceChangePerc = (priceChange / basePrice) * 100
	}

	category := GetPriceCategory(priceChangePerc)
	recommendation := GetPriceRecommendation(priceChangePerc, occupancyRate)

	return &PriceRecommendation{
		BasePrice:        basePrice,
		RecommendedPrice: recommendedPrice,
		OccupancyRate:    occupancyRate,
		DemandCoeff:      demandCoeff,
		TimeCoeff:        timeCoeff,
		DayCoeff:         dayCoeff,
		PriceChange:      priceChange,
		PriceChangePerc:  priceChangePerc,
		Category:         category,
		Recommendation:   recommendation,
	}, nil
}

// CalculatePriceWithCoefficients розраховує ціну з заданими коефіцієнтами
func (s *pricingService) CalculatePriceWithCoefficients(basePrice, demandCoeff, timeCoeff, dayCoeff, minCoeff, maxCoeff float64) float64 {
	// Розрахунок за формулою: P_рек = P_баз × K_попит × K_час × K_день
	recommendedPrice := basePrice * demandCoeff * timeCoeff * dayCoeff

	// Обмеження діапазону на основі системних налаштувань
	minPrice := basePrice * minCoeff
	maxPrice := basePrice * maxCoeff
	recommendedPrice = math.Max(minPrice, math.Min(maxPrice, recommendedPrice))

	// Округлення до найближчих 5 грн
	recommendedPrice = math.Round(recommendedPrice/5) * 5

	return recommendedPrice
}

// calculateDemandCoefficient розраховує коефіцієнт попиту на основі завантаженості та налаштувань
func (s *pricingService) calculateDemandCoefficient(occupancyRate float64, lowThreshold, highThreshold int) float64 {
	switch {
	case occupancyRate < float64(lowThreshold):
		return 0.75 // Низька завантаженість - знижка
	case occupancyRate < 60:
		return 0.95 // Помірна завантаженість - невелика знижка
	case occupancyRate < float64(highThreshold):
		return 1.10 // Висока завантаженість - підвищення
	default:
		return 1.40 // Критична завантаженість - значне підвищення
	}
}

// calculateTimeCoefficient розраховує коефіцієнт часу на основі години доби та налаштувань
func (s *pricingService) calculateTimeCoefficient(departureTime time.Time, peakCoeff float64) float64 {
	hour := departureTime.Hour()

	// Пікові години (ранок та вечір)
	if (hour >= 7 && hour <= 9) || (hour >= 17 && hour <= 19) {
		return peakCoeff
	}

	// Нічні години
	if hour >= 23 || hour <= 6 {
		return 0.80
	}

	// Звичайний час
	return 1.00
}

// calculateDayCoefficient розраховує коефіцієнт дня на основі дня тижня та сезонних налаштувань
func (s *pricingService) calculateDayCoefficient(departureTime time.Time, weekendCoeff float64, seasonalCoeffs map[string]float64) float64 {
	weekday := departureTime.Weekday()
	baseCoeff := 1.00

	// Вихідні дні (субота, неділя)
	if weekday == time.Saturday || weekday == time.Sunday {
		baseCoeff = weekendCoeff
	}

	// Сезонні коефіцієнти
	seasonCoeff := s.getSeasonalCoefficient(departureTime, seasonalCoeffs)

	return baseCoeff * seasonCoeff
}

// getSeasonalCoefficient повертає сезонний коефіцієнт
func (s *pricingService) getSeasonalCoefficient(date time.Time, seasonalCoeffs map[string]float64) float64 {
	month := date.Month()
	day := date.Day()

	// Новорічні свята
	if (month == time.December && day >= 25) || (month == time.January && day <= 10) {
		if coeff, exists := seasonalCoeffs["new_year"]; exists {
			return coeff
		}
	}

	// Літній період
	if month >= time.June && month <= time.August {
		if coeff, exists := seasonalCoeffs["summer"]; exists {
			return coeff
		}
	}

	// Звичайний період
	if coeff, exists := seasonalCoeffs["regular"]; exists {
		return coeff
	}

	return 1.00
}

// GetPriceCategory повертає категорію ціни для відображення
func GetPriceCategory(priceChangePerc float64) string {
	switch {
	case priceChangePerc <= -20:
		return "very_low"
	case priceChangePerc <= -10:
		return "low"
	case priceChangePerc >= 30:
		return "very_high"
	case priceChangePerc >= 15:
		return "high"
	default:
		return "normal"
	}
}

// GetPriceRecommendation повертає текстову рекомендацію
func GetPriceRecommendation(priceChangePerc, occupancyRate float64) string {
	switch {
	case priceChangePerc >= 20:
		return "Підвищення ціни через високу завантаженість та пікові години"
	case priceChangePerc >= 10:
		return "Помірне підвищення ціни через підвищений попит"
	case priceChangePerc <= -15:
		return "Значна знижка для стимулювання попиту"
	case priceChangePerc <= -5:
		return "Невелика знижка через низьку завантаженість"
	default:
		return "Стандартна ціна відповідає поточному попиту"
	}
}

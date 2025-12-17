package repository

import (
	"busoptima/internal/model"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// PriceRecommendationRepository інтерфейс для роботи з рекомендаціями цін
type PriceRecommendationRepository interface {
	Create(ctx context.Context, recommendation *model.PriceRecommendation) error
	GetByTripID(ctx context.Context, tripID int64) ([]model.PriceRecommendation, error)
}

// priceRecommendationRepository реалізація PriceRecommendationRepository
type priceRecommendationRepository struct {
	db *sqlx.DB
}

// NewPriceRecommendationRepository створює новий екземпляр репозиторію рекомендацій цін
func NewPriceRecommendationRepository(db *sqlx.DB) PriceRecommendationRepository {
	return &priceRecommendationRepository{db: db}
}

// Create зберігає нову рекомендацію ціни
func (r *priceRecommendationRepository) Create(ctx context.Context, recommendation *model.PriceRecommendation) error {
	query := `
		INSERT INTO price_recommendations (
			trip_id, base_price, recommended_price, occupancy_rate,
			demand_coefficient, time_coefficient, day_coefficient
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := r.db.QueryRowContext(ctx, query,
		recommendation.TripID,
		recommendation.BasePrice,
		recommendation.RecommendedPrice,
		recommendation.OccupancyRate,
		recommendation.DemandCoeff,
		recommendation.TimeCoeff,
		recommendation.DayCoeff,
	).Scan(&recommendation.ID, &recommendation.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create price recommendation: %w", err)
	}

	return nil
}

// GetByTripID повертає всі рекомендації цін для рейсу
func (r *priceRecommendationRepository) GetByTripID(ctx context.Context, tripID int64) ([]model.PriceRecommendation, error) {
	var recommendations []model.PriceRecommendation
	query := `
		SELECT * FROM price_recommendations 
		WHERE trip_id = $1 
		ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &recommendations, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("failed to get price recommendations: %w", err)
	}

	return recommendations, nil
}

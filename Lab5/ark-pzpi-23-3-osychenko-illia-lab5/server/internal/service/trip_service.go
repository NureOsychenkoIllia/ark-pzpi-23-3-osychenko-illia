package service

import (
	"context"
	"busoptima/internal/model"
	"busoptima/internal/repository"
)

// TripService інтерфейс для роботи з рейсами
type TripService interface {
	Create(ctx context.Context, trip *model.Trip) error
	GetByID(ctx context.Context, id int64) (*model.Trip, error)
	GetAll(ctx context.Context, filters map[string]interface{}) ([]model.Trip, error)
	Update(ctx context.Context, trip *model.Trip) error
	GetEvents(ctx context.Context, tripID int64) ([]model.PassengerEvent, error)
	GetAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error)
}

type tripService struct {
	tripRepo      repository.TripRepository
	eventRepo     repository.PassengerEventRepository
	analyticsRepo repository.AnalyticsRepository
	auditRepo     repository.AuditLogRepository
}

func NewTripService(tripRepo repository.TripRepository, eventRepo repository.PassengerEventRepository, analyticsRepo repository.AnalyticsRepository, auditRepo repository.AuditLogRepository) TripService {
	return &tripService{
		tripRepo:      tripRepo,
		eventRepo:     eventRepo,
		analyticsRepo: analyticsRepo,
		auditRepo:     auditRepo,
	}
}

func (s *tripService) Create(ctx context.Context, trip *model.Trip) error {
	return s.tripRepo.Create(ctx, trip)
}

func (s *tripService) GetByID(ctx context.Context, id int64) (*model.Trip, error) {
	return s.tripRepo.GetByID(ctx, id)
}

func (s *tripService) GetAll(ctx context.Context, filters map[string]interface{}) ([]model.Trip, error) {
	return s.tripRepo.GetAll(ctx, filters)
}

func (s *tripService) Update(ctx context.Context, trip *model.Trip) error {
	return s.tripRepo.Update(ctx, trip)
}

func (s *tripService) GetEvents(ctx context.Context, tripID int64) ([]model.PassengerEvent, error) {
	return s.eventRepo.GetByTripID(ctx, tripID)
}

func (s *tripService) GetAnalytics(ctx context.Context, tripID int64) (*model.TripAnalytics, error) {
	return s.analyticsRepo.GetTripAnalytics(ctx, tripID)
}
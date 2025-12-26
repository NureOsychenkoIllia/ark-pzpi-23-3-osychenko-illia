package service

import (
	"context"
	"busoptima/internal/model"
	"busoptima/internal/repository"
)

// BusService інтерфейс для роботи з автобусами
type BusService interface {
	Create(ctx context.Context, bus *model.Bus) error
	GetByID(ctx context.Context, id int64) (*model.Bus, error)
	GetAll(ctx context.Context, activeOnly bool) ([]model.Bus, error)
	Update(ctx context.Context, bus *model.Bus) error
	Delete(ctx context.Context, id int64) error
}

type busService struct {
	busRepo   repository.BusRepository
	auditRepo repository.AuditLogRepository
}

func NewBusService(busRepo repository.BusRepository, auditRepo repository.AuditLogRepository) BusService {
	return &busService{busRepo: busRepo, auditRepo: auditRepo}
}

func (s *busService) Create(ctx context.Context, bus *model.Bus) error {
	return s.busRepo.Create(ctx, bus)
}

func (s *busService) GetByID(ctx context.Context, id int64) (*model.Bus, error) {
	return s.busRepo.GetByID(ctx, id)
}

func (s *busService) GetAll(ctx context.Context, activeOnly bool) ([]model.Bus, error) {
	return s.busRepo.GetAll(ctx, activeOnly)
}

func (s *busService) Update(ctx context.Context, bus *model.Bus) error {
	return s.busRepo.Update(ctx, bus)
}

func (s *busService) Delete(ctx context.Context, id int64) error {
	return s.busRepo.Delete(ctx, id)
}
package service

import (
	"context"
	"busoptima/internal/model"
	"busoptima/internal/repository"
)

// RouteService інтерфейс для роботи з маршрутами
type RouteService interface {
	Create(ctx context.Context, route *model.Route) error
	GetByID(ctx context.Context, id int64) (*model.Route, error)
	GetAll(ctx context.Context, activeOnly bool) ([]model.Route, error)
	Update(ctx context.Context, route *model.Route) error
	Delete(ctx context.Context, id int64) error
}

// routeService реалізація RouteService
type routeService struct {
	routeRepo repository.RouteRepository
	auditRepo repository.AuditLogRepository
}

// NewRouteService створює новий сервіс маршрутів
func NewRouteService(routeRepo repository.RouteRepository, auditRepo repository.AuditLogRepository) RouteService {
	return &routeService{
		routeRepo: routeRepo,
		auditRepo: auditRepo,
	}
}

func (s *routeService) Create(ctx context.Context, route *model.Route) error {
	return s.routeRepo.Create(ctx, route)
}

func (s *routeService) GetByID(ctx context.Context, id int64) (*model.Route, error) {
	return s.routeRepo.GetByID(ctx, id)
}

func (s *routeService) GetAll(ctx context.Context, activeOnly bool) ([]model.Route, error) {
	return s.routeRepo.GetAll(ctx, activeOnly)
}

func (s *routeService) Update(ctx context.Context, route *model.Route) error {
	return s.routeRepo.Update(ctx, route)
}

func (s *routeService) Delete(ctx context.Context, id int64) error {
	return s.routeRepo.Delete(ctx, id)
}
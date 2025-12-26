package service

import (
	"busoptima/internal/model"
	"busoptima/internal/repository"
	"context"
	"strconv"
)

// AuditService інтерфейс для роботи з журналом аудиту
type AuditService interface {
	LogAction(ctx context.Context, userID int64, action, entityType, entityID string, oldValues, newValues map[string]any, ipAddress string) error
	GetAuditLogs(ctx context.Context, filters map[string]any) ([]model.AuditLog, error)
	GetAuditLogsCount(ctx context.Context, filters map[string]any) (int64, error)
}

// auditService реалізація AuditService
type auditService struct {
	auditRepo repository.AuditLogRepository
}

// NewAuditService створює новий сервіс аудиту
func NewAuditService(auditRepo repository.AuditLogRepository) AuditService {
	return &auditService{
		auditRepo: auditRepo,
	}
}

// LogAction записує дію в журнал аудиту
func (s *auditService) LogAction(ctx context.Context, userID int64, action, entityType, entityID string, oldValues, newValues map[string]any, ipAddress string) error {
	log := &model.AuditLog{
		UserID:     &userID,
		Action:     action,
		EntityType: entityType,
		IPAddress:  ipAddress,
		OldValues:  oldValues,
		NewValues:  newValues,
	}

	if entityID != "" {
		if id, err := strconv.ParseInt(entityID, 10, 64); err == nil {
			log.EntityID = &id
		}
	}

	return s.auditRepo.Create(ctx, log)
}

// GetAuditLogs повертає записи аудиту з фільтрами
func (s *auditService) GetAuditLogs(ctx context.Context, filters map[string]any) ([]model.AuditLog, error) {
	return s.auditRepo.GetAll(ctx, filters)
}

// GetAuditLogsCount повертає загальну кількість записів аудиту з фільтрами
func (s *auditService) GetAuditLogsCount(ctx context.Context, filters map[string]any) (int64, error) {
	return s.auditRepo.GetCount(ctx, filters)
}

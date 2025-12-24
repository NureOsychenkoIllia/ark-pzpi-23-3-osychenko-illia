package middleware

import (
	"busoptima/internal/service"
	"context"

	"github.com/gofiber/fiber/v2"
)

// AuditHelper provides utilities for manual audit logging with proper old/new values
type AuditHelper struct {
	auditService service.AuditService
}

// NewAuditHelper creates a new audit helper
func NewAuditHelper(auditService service.AuditService) *AuditHelper {
	return &AuditHelper{
		auditService: auditService,
	}
}

// LogCreate logs a CREATE operation
func (h *AuditHelper) LogCreate(c *fiber.Ctx, entityType string, entityID string, newValues map[string]any) {
	h.logAction(c, "CREATE", entityType, entityID, make(map[string]any), newValues)
}

// LogUpdate logs an UPDATE operation with proper old and new values
func (h *AuditHelper) LogUpdate(c *fiber.Ctx, entityType string, entityID string, oldValues, newValues map[string]any) {
	h.logAction(c, "UPDATE", entityType, entityID, oldValues, newValues)
}

// LogDelete logs a DELETE operation
func (h *AuditHelper) LogDelete(c *fiber.Ctx, entityType string, entityID string, oldValues map[string]any) {
	h.logAction(c, "DELETE", entityType, entityID, oldValues, make(map[string]any))
}

// logAction is the internal method that performs the actual logging
func (h *AuditHelper) logAction(c *fiber.Ctx, action, entityType, entityID string, oldValues, newValues map[string]any) {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return
	}

	ipAddress := c.IP()

	// Use goroutine to avoid blocking the request
	go func() {
		if err := h.auditService.LogAction(context.Background(), userID, action, entityType, entityID, oldValues, newValues, ipAddress); err != nil {
			// Log error but don't fail the request
			// In production, you might want to use a proper logger here
			_ = err
		}
	}()
}

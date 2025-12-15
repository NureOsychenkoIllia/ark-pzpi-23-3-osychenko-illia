package middleware

import (
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuditLog middleware для журналювання дій користувачів
func AuditLog() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Зберігаємо оригінальне тіло запиту для аудиту
		if shouldAudit(c.Method(), c.Path()) {
			body := c.Body()
			if len(body) > 0 {
				var oldValues map[string]interface{}
				json.Unmarshal(body, &oldValues)
				c.Locals("audit_old_values", oldValues)
			}
		}

		// Продовжуємо обробку запиту
		err := c.Next()

		// Після обробки запиту записуємо в аудит лог
		if shouldAudit(c.Method(), c.Path()) && c.Response().StatusCode() < 400 {
			go logAuditEvent(c)
		}

		return err
	}
}

// shouldAudit визначає, чи потрібно логувати цю дію
func shouldAudit(method, path string) bool {
	// Тільки модифікуючі операції
	if method != "POST" && method != "PUT" && method != "DELETE" {
		return false
	}

	// Не логуємо автентифікацію
	if strings.Contains(path, "/auth/") {
		return false
	}

	return true
}

// logAuditEvent записує подію в журнал аудиту
func logAuditEvent(c *fiber.Ctx) {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return
	}

	action := getActionFromMethod(c.Method())
	entityType := getEntityTypeFromPath(c.Path())
	entityID := getEntityIDFromPath(c.Path())

	// TODO: Implement audit service call
	// auditService.LogAction(userID, action, entityType, entityID, oldValues, newValues, c.IP())
	
	// Suppress unused variable warnings
	_ = userID
	_ = action
	_ = entityType
	_ = entityID
}

func getActionFromMethod(method string) string {
	switch method {
	case "POST":
		return "CREATE"
	case "PUT":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

func getEntityTypeFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 {
		return parts[2] // /api/routes -> routes
	}
	return "unknown"
}

func getEntityIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 4 {
		return parts[3] // /api/routes/123 -> 123
	}
	return ""
}
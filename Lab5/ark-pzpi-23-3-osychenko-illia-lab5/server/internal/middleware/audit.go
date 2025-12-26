package middleware

import (
	"busoptima/internal/repository"
	"busoptima/internal/service"
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuditLog middleware для журналювання дій користувачів з отриманням старих значень
func AuditLog(auditService service.AuditService, repos *repository.Repositories) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var requestBody []byte
		var oldValues map[string]any

		// Зберігаємо оригінальне тіло запиту для аудиту
		if shouldAudit(c.Method(), c.Path()) {
			requestBody = c.Body()

			// For UPDATE and DELETE operations, fetch old values before the operation
			if c.Method() == "PUT" || c.Method() == "DELETE" {
				entityType := getEntityTypeFromPath(c.Path())
				entityID := getEntityIDFromPath(c.Path())

				if entityID != "" {
					oldValues = fetchOldValues(c.Context(), repos, entityType, entityID)
				}
			}
		}

		// Продовжуємо обробку запиту
		err := c.Next()

		// Після обробки запиту записуємо в аудит лог
		if shouldAudit(c.Method(), c.Path()) && c.Response().StatusCode() < 400 {
			// Capture all necessary data before starting goroutine
			userID, ok := c.Locals("user_id").(int64)
			if ok {
				action := getActionFromMethod(c.Method())
				entityType := getEntityTypeFromPath(c.Path())
				entityID := getEntityIDFromPath(c.Path())
				ipAddress := c.IP()

				// Initialize values
				if oldValues == nil {
					oldValues = make(map[string]any)
				}
				newValues := make(map[string]any)

				if c.Method() == "POST" {
					// CREATE operation: new values are the request body
					if len(requestBody) > 0 {
						json.Unmarshal(requestBody, &newValues)
					}
				} else if c.Method() == "PUT" {
					// UPDATE operation: new values are the request body, old values fetched above
					if len(requestBody) > 0 {
						json.Unmarshal(requestBody, &newValues)
					}
				}
				// For DELETE: old values fetched above, new values remain empty

				// Now start goroutine with captured data
				go func() {
					if err := auditService.LogAction(context.Background(), userID, action, entityType, entityID, oldValues, newValues, ipAddress); err != nil {
						// Log error but don't fail the request
						_ = err
					}
				}()
			}
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

// fetchOldValues retrieves the current state of an entity from the database
func fetchOldValues(ctx context.Context, repos *repository.Repositories, entityType, entityID string) map[string]any {
	id, err := strconv.ParseInt(entityID, 10, 64)
	if err != nil {
		return make(map[string]any)
	}

	switch entityType {
	case "buses":
		if bus, err := repos.Bus.GetByID(ctx, id); err == nil {
			return entityToMap(bus)
		}
	case "routes":
		if route, err := repos.Route.GetByID(ctx, id); err == nil {
			return entityToMap(route)
		}
	case "trips":
		if trip, err := repos.Trip.GetByID(ctx, id); err == nil {
			return entityToMap(trip)
		}
	case "users":
		if user, err := repos.User.GetByID(ctx, id); err == nil {
			return entityToMap(user)
		}
	}

	return make(map[string]any)
}

// entityToMap converts an entity struct to a map for audit logging
func entityToMap(entity any) map[string]any {
	result := make(map[string]any)

	// Convert entity to JSON and back to map
	if jsonBytes, err := json.Marshal(entity); err == nil {
		json.Unmarshal(jsonBytes, &result)
	}

	return result
}

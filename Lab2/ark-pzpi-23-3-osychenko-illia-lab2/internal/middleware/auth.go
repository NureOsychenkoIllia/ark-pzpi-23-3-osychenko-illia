package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth middleware для перевірки JWT токенів
func JWTAuth(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(401).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid authorization header format",
			})
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(401, "Invalid signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Встановлюємо дані користувача в контекст
		c.Locals("user_id", int64(claims["user_id"].(float64)))
		c.Locals("role", claims["role"].(string))
		c.Locals("permissions", claims["permissions"])

		return c.Next()
	}
}

// RequirePermission middleware для перевірки дозволів
func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissions, ok := c.Locals("permissions").([]interface{})
		if !ok {
			return c.Status(403).JSON(fiber.Map{
				"error":               "Access denied",
				"required_permission": permission,
			})
		}

		// Перевіряємо наявність необхідного дозволу
		for _, p := range permissions {
			if p.(string) == permission {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{
			"error":               "Insufficient permissions",
			"required_permission": permission,
		})
	}
}
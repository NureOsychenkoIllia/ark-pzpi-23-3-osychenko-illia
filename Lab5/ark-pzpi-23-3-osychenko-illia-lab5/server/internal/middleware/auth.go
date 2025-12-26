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

		// Перевіряємо тип токена
		tokenType, hasType := claims["type"].(string)
		if hasType && tokenType == "device" {
			// Це токен пристрою
			deviceID, ok := claims["device_id"].(float64)
			if !ok {
				return c.Status(401).JSON(fiber.Map{
					"error": "Invalid device token claims",
				})
			}
			serialNumber, ok := claims["serial_number"].(string)
			if !ok {
				return c.Status(401).JSON(fiber.Map{
					"error": "Invalid device token claims",
				})
			}

			c.Locals("device_id", int64(deviceID))
			c.Locals("serial_number", serialNumber)
			c.Locals("token_type", "device")
		} else {
			// Це токен користувача
			userID, ok := claims["user_id"].(float64)
			if !ok {
				return c.Status(401).JSON(fiber.Map{
					"error": "Invalid user token claims",
				})
			}
			role, ok := claims["role"].(string)
			if !ok {
				return c.Status(401).JSON(fiber.Map{
					"error": "Invalid user token claims",
				})
			}

			c.Locals("user_id", int64(userID))
			c.Locals("role", role)
			c.Locals("permissions", claims["permissions"])
			c.Locals("token_type", "user")
		}

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

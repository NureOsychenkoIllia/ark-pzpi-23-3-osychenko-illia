package handler

import (
	"busoptima/internal/service"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler обробляє запити автентифікації
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler створює новий обробник автентифікації
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest структура запиту на вхід
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// DeviceAuthRequest структура запиту автентифікації пристрою
type DeviceAuthRequest struct {
	SerialNumber string `json:"serial_number" validate:"required"`
	Token        string `json:"token" validate:"required"`
}

// RefreshTokenRequest структура запиту оновлення токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// Login обробляє вхід користувача
//
//	@Summary		Автентифікація користувача
//	@Description	Автентифікує користувача та повертає JWT токени
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		LoginRequest	true	"Облікові дані користувача"
//	@Success		200			{object}	service.LoginResponse
//	@Failure		400			{object}	map[string]string
//	@Failure		401			{object}	map[string]string
//	@Router			/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	response, err := h.authService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

// DeviceAuth обробляє автентифікацію IoT-пристрою
//
//	@Summary		Автентифікація IoT-пристрою
//	@Description	Автентифікує IoT-пристрій та повертає токен доступу
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			device	body		DeviceAuthRequest	true	"Дані пристрою"
//	@Success		200		{object}	service.DeviceAuthResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/auth/device [post]
func (h *AuthHandler) DeviceAuth(c *fiber.Ctx) error {
	var req DeviceAuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	response, err := h.authService.DeviceAuth(c.Context(), req.SerialNumber, req.Token)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

// RefreshToken обробляє оновлення токена
//
//	@Summary		Оновити токен доступу
//	@Description	Оновлює токен доступу за допомогою refresh токена
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			refresh	body		RefreshTokenRequest	true	"Refresh токен"
//	@Success		200		{object}	service.LoginResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	response, err := h.authService.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

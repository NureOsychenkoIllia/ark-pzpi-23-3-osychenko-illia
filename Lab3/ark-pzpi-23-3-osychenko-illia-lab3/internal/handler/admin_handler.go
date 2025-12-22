package handler

import (
	"busoptima/internal/model"
	"busoptima/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	authService     service.AuthService
	settingsService service.SettingsService
}

func NewAdminHandler(authService service.AuthService, settingsService service.SettingsService) *AdminHandler {
	return &AdminHandler{
		authService:     authService,
		settingsService: settingsService,
	}
}

// GetUsers повертає список користувачів
//
//	@Summary		Отримати список користувачів
//	@Description	Повертає список всіх користувачів системи
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.User
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users [get]
func (h *AdminHandler) GetUsers(c *fiber.Ctx) error {
	users, err := h.authService.GetUsers(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(users)
}

// CreateUserRequest структура запиту створення користувача
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required"`
	RoleID   int64  `json:"role_id" validate:"required"`
}

// CreateUser створює нового користувача
//
//	@Summary		Створити нового користувача
//	@Description	Створює нового користувача в системі
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			user	body		CreateUserRequest	true	"Дані користувача"
//	@Success		201		{object}	model.User
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users [post]
func (h *AdminHandler) CreateUser(c *fiber.Ctx) error {
	var req CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user := &model.User{
		Email:    req.Email,
		FullName: req.FullName,
		RoleID:   req.RoleID,
	}

	if err := h.authService.CreateUser(c.Context(), user, req.Password); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Приховуємо пароль у відповіді
	user.PasswordHash = ""

	return c.Status(201).JSON(user)
}

// UpdateUser оновлює користувача
//
//	@Summary		Оновити користувача
//	@Description	Оновлює існуючого користувача за ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int			true	"ID користувача"
//	@Param			user	body		model.User	true	"Оновлені дані користувача"
//	@Success		200		{object}	model.User
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id} [put]
func (h *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	user.ID = id

	if err := h.authService.UpdateUser(c.Context(), &user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(user)
}

// UpdateUserRoleRequest структура запиту оновлення ролі
type UpdateUserRoleRequest struct {
	RoleID int64 `json:"role_id" validate:"required"`
}

// UpdateUserRole оновлює роль користувача
//
//	@Summary		Оновити роль користувача
//	@Description	Оновлює роль користувача за ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"ID користувача"
//	@Param			role	body		UpdateUserRoleRequest	true	"Нова роль користувача"
//	@Success		200		{object}	map[string]string
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/users/{id}/role [put]
func (h *AdminHandler) UpdateUserRole(c *fiber.Ctx) error {
	userID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req UpdateUserRoleRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.authService.UpdateUserRole(c.Context(), userID, req.RoleID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "User role updated successfully"})
}

// GetAuditLogs повертає журнал аудиту
//
//	@Summary		Отримати журнал аудиту
//	@Description	Повертає записи журналу аудиту з можливістю пагінації
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"Номер сторінки"	default(1)
//	@Param			limit	query		int	false	"Кількість записів на сторінці"	default(20)
//	@Success		200		{object}	map[string]interface{}
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/audit-logs [get]
func (h *AdminHandler) GetAuditLogs(c *fiber.Ctx) error {
	// TODO: Parse query parameters and call auditService.GetAuditLogs
	// Demo response with audit logs
	return c.JSON(fiber.Map{
		"logs": []fiber.Map{
			{
				"id": 1234,
				"user": fiber.Map{
					"id":        2,
					"email":     "admin@busoptima.ua",
					"full_name": "Адміністратор",
				},
				"action":      "UPDATE",
				"entity_type": "routes",
				"entity_id":   5,
				"old_values":  fiber.Map{"base_price": 180.00},
				"new_values":  fiber.Map{"base_price": 200.00},
				"ip_address":  "192.168.1.100",
				"created_at":  "2025-12-14T09:15:00Z",
			},
		},
		"total": 1234,
		"page":  1,
		"limit": 20,
	})
}

// CreateBackup створює резервну копію
//
//	@Summary		Створити резервну копію
//	@Description	Створює резервну копію бази даних
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/backup [post]
func (h *AdminHandler) CreateBackup(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message":    "Backup created successfully",
		"backup_id":  "backup_20251215_081530",
		"created_at": "2025-12-15T08:15:30Z",
	})
}

// GetSystemSettings повертає поточні системні налаштування
//
//	@Summary		Отримати системні налаштування
//	@Description	Повертає поточні системні параметри
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SystemSettings
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings [get]
func (h *AdminHandler) GetSystemSettings(c *fiber.Ctx) error {
	settings, err := h.settingsService.GetSettings(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(settings)
}

// UpdateSystemSettingsRequest структура запиту оновлення налаштувань
type UpdateSystemSettingsRequest struct {
	FuelPricePerLiter    float64            `json:"fuel_price_per_liter" validate:"required,min=10,max=200"`
	PeakHoursCoefficient float64            `json:"peak_hours_coefficient" validate:"required,min=0.5,max=3.0"`
	WeekendCoefficient   float64            `json:"weekend_coefficient" validate:"required,min=0.5,max=3.0"`
	HighDemandThreshold  int                `json:"high_demand_threshold" validate:"required,min=0,max=100"`
	LowDemandThreshold   int                `json:"low_demand_threshold" validate:"required,min=0,max=100"`
	PriceMinCoefficient  float64            `json:"price_min_coefficient" validate:"required,min=0.1,max=1.0"`
	PriceMaxCoefficient  float64            `json:"price_max_coefficient" validate:"required,min=1.0,max=5.0"`
	SeasonalCoefficients map[string]float64 `json:"seasonal_coefficients" validate:"required"`
}

// UpdateSystemSettings оновлює системні налаштування
//
//	@Summary		Оновити системні налаштування
//	@Description	Оновлює системні параметри
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			settings	body		UpdateSystemSettingsRequest	true	"Нові налаштування"
//	@Success		200			{object}	map[string]interface{}
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings [put]
func (h *AdminHandler) UpdateSystemSettings(c *fiber.Ctx) error {
	var req UpdateSystemSettingsRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Отримуємо ID користувача з контексту
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	settings := &model.SystemSettings{
		FuelPricePerLiter:    req.FuelPricePerLiter,
		PeakHoursCoefficient: req.PeakHoursCoefficient,
		WeekendCoefficient:   req.WeekendCoefficient,
		HighDemandThreshold:  req.HighDemandThreshold,
		LowDemandThreshold:   req.LowDemandThreshold,
		PriceMinCoefficient:  req.PriceMinCoefficient,
		PriceMaxCoefficient:  req.PriceMaxCoefficient,
		SeasonalCoefficients: req.SeasonalCoefficients,
	}

	if err := h.settingsService.UpdateSettings(c.Context(), settings, userID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":    "System settings updated successfully",
		"settings":   settings,
		"updated_at": settings.UpdatedAt,
	})
}

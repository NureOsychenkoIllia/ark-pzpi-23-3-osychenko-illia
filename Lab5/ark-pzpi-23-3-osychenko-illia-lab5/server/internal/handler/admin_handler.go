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
	backupService   service.BackupService
	auditService    service.AuditService
}

func NewAdminHandler(authService service.AuthService, settingsService service.SettingsService, backupService service.BackupService, auditService service.AuditService) *AdminHandler {
	return &AdminHandler{
		authService:     authService,
		settingsService: settingsService,
		backupService:   backupService,
		auditService:    auditService,
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
//	@Success		200		{object}	UpdateUserRoleResponse
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

	return c.JSON(UpdateUserRoleResponse{
		Message: "User role updated successfully",
	})
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
//	@Success		200		{object}	AuditLogsResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/audit-logs [get]
func (h *AdminHandler) GetAuditLogs(c *fiber.Ctx) error {
	// Розбираємо параметри запиту
	filters := make(map[string]any)

	// Сторінка та ліміт для пагінації
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit
	filters["limit"] = limit
	filters["offset"] = offset

	// Додаткові фільтри
	if userID := c.QueryInt("user_id", 0); userID > 0 {
		filters["user_id"] = int64(userID)
	}

	if action := c.Query("action"); action != "" {
		filters["action"] = action
	}

	if entityType := c.Query("entity_type"); entityType != "" {
		filters["entity_type"] = entityType
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}

	logs, err := h.auditService.GetAuditLogs(c.Context(), filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Отримуємо загальну кількість (без фільтрів пагінації)
	countFilters := make(map[string]any)
	for k, v := range filters {
		if k != "limit" && k != "offset" {
			countFilters[k] = v
		}
	}

	total, err := h.auditService.GetAuditLogsCount(c.Context(), countFilters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(AuditLogsResponse{
		Logs:  logs,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// // CreateBackup створює резервну копію
// //
// //	@Summary		Створити резервну копію
// //	@Description	Створює резервну копію бази даних
// //	@Tags			Admin
// //	@Accept			json
// //	@Produce		json
// //	@Success		200	{object}	service.BackupInfo
// //	@Failure 500 {object} ErrorResponse
// //	@Security		BearerAuth
// //	@Router			/admin/backup [post]
// func (h *AdminHandler) CreateBackup(c *fiber.Ctx) error {
// 	backup, err := h.backupService.CreateBackup(c.Context())
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(backup)
// }

// // ListBackups повертає список резервних копій
// //
// //	@Summary		Список резервних копій
// //	@Description	Повертає список всіх резервних копій
// //	@Tags			Admin
// //	@Accept			json
// //	@Produce		json
// //	@Success		200	{array}	service.BackupInfo
// //	@Failure 500 {object} ErrorResponse
// //	@Security		BearerAuth
// //	@Router			/admin/backups [get]
// func (h *AdminHandler) ListBackups(c *fiber.Ctx) error {
// 	backups, err := h.backupService.ListBackups(c.Context())
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(backups)
// }

// // RestoreBackup відновлює базу даних з резервної копії
// //
// //	@Summary		Відновити з резервної копії
// //	@Description	Відновлює базу даних з вказаної резервної копії
// //	@Tags			Admin
// //	@Accept			json
// //	@Produce		json
// //	@Param			backup_id	path		string	true	"ID резервної копії"
// //	@Success		200			{object}	map[string]interface{}
// //	@Failure 400 {object} ErrorResponse
// //	@Failure 500 {object} ErrorResponse
// //	@Security		BearerAuth
// //	@Router			/admin/backups/{backup_id}/restore [post]
// func (h *AdminHandler) RestoreBackup(c *fiber.Ctx) error {
// 	backupID := c.Params("backup_id")
// 	if backupID == "" {
// 		return c.Status(400).JSON(fiber.Map{"error": "backup_id is required"})
// 	}

// 	if err := h.backupService.RestoreBackup(c.Context(), backupID); err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{
// 		"message":     "Database restored successfully",
// 		"backup_id":   backupID,
// 		"restored_at": time.Now(),
// 	})
// }

// GetSystemSettings повертає поточні системні налаштування
//
//	@Summary		Отримати системні налаштування
//	@Description	Повертає поточні системні параметри
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	SystemSettingsResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings [get]
func (h *AdminHandler) GetSystemSettings(c *fiber.Ctx) error {
	settings, err := h.settingsService.GetSettings(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Конвертуємо модель в структуру відповіді
	settingsResponse := &SystemSettingsResponse{
		ID:                   settings.ID,
		FuelPricePerLiter:    settings.FuelPricePerLiter,
		PeakHoursCoefficient: settings.PeakHoursCoefficient,
		WeekendCoefficient:   settings.WeekendCoefficient,
		HighDemandThreshold:  settings.HighDemandThreshold,
		LowDemandThreshold:   settings.LowDemandThreshold,
		PriceMinCoefficient:  settings.PriceMinCoefficient,
		PriceMaxCoefficient:  settings.PriceMaxCoefficient,
		SeasonalCoefficients: SeasonalCoefficients{
			Spring: settings.SeasonalCoefficients["spring"],
			Summer: settings.SeasonalCoefficients["summer"],
			Autumn: settings.SeasonalCoefficients["autumn"],
			Winter: settings.SeasonalCoefficients["winter"],
		},
		UpdatedAt:     settings.UpdatedAt,
		UpdatedBy:     settings.UpdatedBy,
		UpdatedByUser: settings.UpdatedByUser,
	}

	return c.JSON(settingsResponse)
}

// SeasonalCoefficients представляє сезонні коефіцієнти ціноутворення
type SeasonalCoefficients struct {
	Spring float64 `json:"spring" example:"1.0"`
	Summer float64 `json:"summer" example:"1.2"`
	Autumn float64 `json:"autumn" example:"1.1"`
	Winter float64 `json:"winter" example:"0.9"`
}

// UpdateSystemSettingsRequest структура запиту оновлення налаштувань
type UpdateSystemSettingsRequest struct {
	FuelPricePerLiter    float64              `json:"fuel_price_per_liter" validate:"required,min=10,max=200"`
	PeakHoursCoefficient float64              `json:"peak_hours_coefficient" validate:"required,min=0.5,max=3.0"`
	WeekendCoefficient   float64              `json:"weekend_coefficient" validate:"required,min=0.5,max=3.0"`
	HighDemandThreshold  int                  `json:"high_demand_threshold" validate:"required,min=0,max=100"`
	LowDemandThreshold   int                  `json:"low_demand_threshold" validate:"required,min=0,max=100"`
	PriceMinCoefficient  float64              `json:"price_min_coefficient" validate:"required,min=0.1,max=1.0"`
	PriceMaxCoefficient  float64              `json:"price_max_coefficient" validate:"required,min=1.0,max=5.0"`
	SeasonalCoefficients SeasonalCoefficients `json:"seasonal_coefficients" validate:"required"`
}

// UpdateSystemSettings оновлює системні налаштування
//
//	@Summary		Оновити системні налаштування
//	@Description	Оновлює системні параметри
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			settings	body		UpdateSystemSettingsRequest	true	"Нові налаштування"
//	@Success		200			{object}	UpdateSystemSettingsResponse
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
		SeasonalCoefficients: map[string]float64{
			"spring": req.SeasonalCoefficients.Spring,
			"summer": req.SeasonalCoefficients.Summer,
			"autumn": req.SeasonalCoefficients.Autumn,
			"winter": req.SeasonalCoefficients.Winter,
		},
	}

	if err := h.settingsService.UpdateSettings(c.Context(), settings, userID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Конвертуємо модель в структуру відповіді
	settingsResponse := &SystemSettingsResponse{
		ID:                   settings.ID,
		FuelPricePerLiter:    settings.FuelPricePerLiter,
		PeakHoursCoefficient: settings.PeakHoursCoefficient,
		WeekendCoefficient:   settings.WeekendCoefficient,
		HighDemandThreshold:  settings.HighDemandThreshold,
		LowDemandThreshold:   settings.LowDemandThreshold,
		PriceMinCoefficient:  settings.PriceMinCoefficient,
		PriceMaxCoefficient:  settings.PriceMaxCoefficient,
		SeasonalCoefficients: SeasonalCoefficients{
			Spring: settings.SeasonalCoefficients["spring"],
			Summer: settings.SeasonalCoefficients["summer"],
			Autumn: settings.SeasonalCoefficients["autumn"],
			Winter: settings.SeasonalCoefficients["winter"],
		},
		UpdatedAt:     settings.UpdatedAt,
		UpdatedBy:     settings.UpdatedBy,
		UpdatedByUser: settings.UpdatedByUser,
	}

	return c.JSON(UpdateSystemSettingsResponse{
		Message:   "System settings updated successfully",
		Settings:  settingsResponse,
		UpdatedAt: &settings.UpdatedAt,
	})
}

// ExportSystemSettings експортує системні налаштування у JSON
//
//	@Summary		Експортувати системні налаштування
//	@Description	Експортує поточні системні налаштування у форматі JSON для резервного копіювання або перенесення
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	service.SettingsExport
//	@Failure		500	{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/export [get]
func (h *AdminHandler) ExportSystemSettings(c *fiber.Ctx) error {
	export, err := h.settingsService.ExportSettings(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	c.Set("Content-Disposition", "attachment; filename=busoptima_settings.json")
	return c.JSON(export)
}

// ImportSystemSettings імпортує системні налаштування з JSON
//
//	@Summary		Імпортувати системні налаштування
//	@Description	Імпортує системні налаштування з раніше експортованого JSON файлу
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			settings	body		service.SettingsExport	true	"Експортовані налаштування"
//	@Success		200			{object}	ImportSettingsResponse
//	@Failure		400			{object}	ErrorResponse
//	@Failure		500			{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/admin/settings/import [post]
func (h *AdminHandler) ImportSystemSettings(c *fiber.Ctx) error {
	var export service.SettingsExport

	if err := c.BodyParser(&export); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body: " + err.Error()})
	}

	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "User not authenticated"})
	}

	if err := h.settingsService.ImportSettings(c.Context(), &export, userID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(ImportSettingsResponse{
		Message:    "Settings imported successfully",
		ImportedAt: c.Context().Value("request_time"),
	})
}

// ImportSettingsResponse структура відповіді імпорту налаштувань
type ImportSettingsResponse struct {
	Message    string      `json:"message"`
	ImportedAt interface{} `json:"imported_at"`
}

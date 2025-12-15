package handler

import (
	"busoptima/internal/model"
	"busoptima/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type TripHandler struct {
	tripService service.TripService
}

func NewTripHandler(tripService service.TripService) *TripHandler {
	return &TripHandler{tripService: tripService}
}

// GetAll повертає список рейсів з фільтрами
//
//	@Summary		Отримати список рейсів
//	@Description	Повертає список всіх рейсів з можливістю фільтрації
//	@Tags			Trips
//	@Accept			json
//	@Produce		json
//	@Param			route_id	query		int		false	"ID маршруту для фільтрації"
//	@Param			status		query		string	false	"Статус рейсу (scheduled, in_progress, completed, cancelled)"
//	@Param			date_from	query		string	false	"Дата початку періоду (YYYY-MM-DD)"
//	@Param			date_to		query		string	false	"Дата кінця періоду (YYYY-MM-DD)"
//	@Success		200			{array}		model.Trip
//	@Failure		500			{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/trips [get]
func (h *TripHandler) GetAll(c *fiber.Ctx) error {
	filters := make(map[string]interface{})

	if routeID := c.QueryInt("route_id", 0); routeID > 0 {
		filters["route_id"] = routeID
	}

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		filters["date_from"] = dateFrom
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		filters["date_to"] = dateTo
	}

	trips, err := h.tripService.GetAll(c.Context(), filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(trips)
}

// GetByID повертає рейс за ID
//
//	@Summary		Отримати рейс за ID
//	@Description	Повертає рейс за вказаним ідентифікатором
//	@Tags			Trips
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"ID рейсу"
//	@Success		200	{object}	model.Trip
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/trips/{id} [get]
func (h *TripHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid trip ID"})
	}

	trip, err := h.tripService.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Trip not found"})
	}

	return c.JSON(trip)
}

// Create створює новий рейс
//
//	@Summary		Створити новий рейс
//	@Description	Створює новий рейс в системі
//	@Tags			Trips
//	@Accept			json
//	@Produce		json
//	@Param			trip	body		model.Trip	true	"Дані рейсу"
//	@Success		201		{object}	model.Trip
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/trips [post]
func (h *TripHandler) Create(c *fiber.Ctx) error {
	var trip model.Trip
	if err := c.BodyParser(&trip); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if trip.Status == "" {
		trip.Status = "scheduled"
	}

	if err := h.tripService.Create(c.Context(), &trip); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(trip)
}

// Update оновлює рейс
//
//	@Summary		Оновити рейс
//	@Description	Оновлює існуючий рейс за ID
//	@Tags			Trips
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int			true	"ID рейсу"
//	@Param			trip	body		model.Trip	true	"Оновлені дані рейсу"
//	@Success		200		{object}	model.Trip
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/trips/{id} [put]
func (h *TripHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid trip ID"})
	}

	var trip model.Trip
	if err := c.BodyParser(&trip); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	trip.ID = id

	if err := h.tripService.Update(c.Context(), &trip); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(trip)
}

// GetEvents повертає події пасажирів для рейсу
//
//	@Summary		Отримати події пасажирів рейсу
//	@Description	Повертає список подій входу/виходу пасажирів для конкретного рейсу
//	@Tags			Trips
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"ID рейсу"
//	@Success		200	{array}		model.PassengerEvent
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/trips/{id}/events [get]
func (h *TripHandler) GetEvents(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid trip ID"})
	}

	events, err := h.tripService.GetEvents(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(events)
}

// GetAnalytics повертає аналітику рейсу
//
//	@Summary		Отримати аналітику рейсу
//	@Description	Повертає детальну аналітику для конкретного рейсу
//	@Tags			Trips
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"ID рейсу"
//	@Success		200	{object}	model.TripAnalytics
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/trips/{id}/analytics [get]
func (h *TripHandler) GetAnalytics(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid trip ID"})
	}

	analytics, err := h.tripService.GetAnalytics(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(analytics)
}

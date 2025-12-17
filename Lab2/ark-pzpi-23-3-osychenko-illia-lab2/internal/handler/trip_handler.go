package handler

import (
	"busoptima/internal/model"
	"busoptima/internal/service"
	"strconv"
	"time"

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
//	@Failure 500 {object} ErrorResponse
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
//	@Failure 400 {object} ErrorResponse
//	@Failure 404 {object} ErrorResponse
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

// CreateTripRequest структура запиту створення рейсу
type CreateTripRequest struct {
	RouteID            int64     `json:"route_id" validate:"required" example:"1"`
	BusID              int64     `json:"bus_id" validate:"required" example:"3"`
	ScheduledDeparture time.Time `json:"scheduled_departure" validate:"required" example:"2025-12-15T08:00:00Z"`
	DriverName         string    `json:"driver_name" validate:"required" example:"Петренко І.П."`
}

// Create створює новий рейс
//
//	@Summary		Створити новий рейс
//	@Description	Створює новий рейс в системі
//	@Tags			Trips
//	@Accept			json
//	@Produce		json
//	@Param			trip	body		CreateTripRequest	true	"Дані рейсу"
//	@Success		201		{object}	model.Trip
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/trips [post]
func (h *TripHandler) Create(c *fiber.Ctx) error {
	var req CreateTripRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	trip := &model.Trip{
		RouteID:            req.RouteID,
		BusID:              req.BusID,
		ScheduledDeparture: req.ScheduledDeparture,
		DriverName:         req.DriverName,
		Status:             "scheduled",
		CurrentPassengers:  0,
	}

	if err := h.tripService.Create(c.Context(), trip); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(trip)
}

// UpdateTripRequest структура запиту оновлення рейсу
type UpdateTripRequest struct {
	RouteID            *int64     `json:"route_id,omitempty" example:"1"`
	BusID              *int64     `json:"bus_id,omitempty" example:"3"`
	ScheduledDeparture *time.Time `json:"scheduled_departure,omitempty" example:"2025-12-15T08:00:00Z"`
	ActualDeparture    *time.Time `json:"actual_departure,omitempty" example:"2025-12-15T08:05:00Z"`
	ActualArrival      *time.Time `json:"actual_arrival,omitempty" example:"2025-12-15T14:05:00Z"`
	Status             *string    `json:"status,omitempty" example:"completed" enums:"scheduled,in_progress,completed,cancelled"`
	CurrentPassengers  *int       `json:"current_passengers,omitempty" example:"35"`
	DriverName         *string    `json:"driver_name,omitempty" example:"Петро Петренко"`
}

// Update оновлює рейс
//
//	@Summary		Оновити рейс
//	@Description	Оновлює існуючий рейс за ID
//	@Tags			Trips
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"ID рейсу"
//	@Param			trip	body		UpdateTripRequest	true	"Оновлені дані рейсу"
//	@Success		200		{object}	model.Trip
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/trips/{id} [put]
func (h *TripHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid trip ID"})
	}

	var req UpdateTripRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Get existing trip first
	existingTrip, err := h.tripService.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Trip not found"})
	}

	// Update only provided fields
	if req.RouteID != nil {
		existingTrip.RouteID = *req.RouteID
	}
	if req.BusID != nil {
		existingTrip.BusID = *req.BusID
	}
	if req.ScheduledDeparture != nil {
		existingTrip.ScheduledDeparture = *req.ScheduledDeparture
	}
	if req.ActualDeparture != nil {
		existingTrip.ActualDeparture = req.ActualDeparture
	}
	if req.ActualArrival != nil {
		existingTrip.ActualArrival = req.ActualArrival
	}
	if req.Status != nil {
		existingTrip.Status = *req.Status
	}
	if req.CurrentPassengers != nil {
		existingTrip.CurrentPassengers = *req.CurrentPassengers
	}
	if req.DriverName != nil {
		existingTrip.DriverName = *req.DriverName
	}

	if err := h.tripService.Update(c.Context(), existingTrip); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(existingTrip)
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
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
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
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
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

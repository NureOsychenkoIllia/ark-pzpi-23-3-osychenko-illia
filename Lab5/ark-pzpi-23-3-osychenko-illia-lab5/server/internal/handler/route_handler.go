package handler

import (
	"busoptima/internal/model"
	"busoptima/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// RouteHandler обробляє запити для маршрутів
type RouteHandler struct {
	routeService service.RouteService
}

// NewRouteHandler створює новий обробник маршрутів
func NewRouteHandler(routeService service.RouteService) *RouteHandler {
	return &RouteHandler{routeService: routeService}
}

// GetAll повертає список маршрутів
//
//	@Summary		Отримати список маршрутів
//	@Description	Повертає список всіх маршрутів з можливістю фільтрації
//	@Tags			Routes
//	@Accept			json
//	@Produce		json
//	@Param			active_only	query		bool	false	"Тільки активні маршрути"
//	@Success		200			{array}		model.Route
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/routes [get]
func (h *RouteHandler) GetAll(c *fiber.Ctx) error {
	activeOnly := c.QueryBool("active_only", true)

	routes, err := h.routeService.GetAll(c.Context(), activeOnly)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(routes)
}

// GetByID повертає маршрут за ID
//
//	@Summary		Отримати маршрут за ID
//	@Description	Повертає маршрут за вказаним ідентифікатором
//	@Tags			Routes
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"ID маршруту"
//	@Success		200	{object}	model.Route
//	@Failure 400 {object} ErrorResponse
//	@Failure 404 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/routes/{id} [get]
func (h *RouteHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid route ID"})
	}

	route, err := h.routeService.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Route not found"})
	}

	return c.JSON(route)
}

// Create створює новий маршрут
//
//	@Summary		Створити новий маршрут
//	@Description	Створює новий маршрут в системі
//	@Tags			Routes
//	@Accept			json
//	@Produce		json
//	@Param			route	body		model.Route	true	"Дані маршруту"
//	@Success		201		{object}	model.Route
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/routes [post]
func (h *RouteHandler) Create(c *fiber.Ctx) error {
	var route model.Route
	if err := c.BodyParser(&route); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	route.IsActive = true

	if err := h.routeService.Create(c.Context(), &route); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(route)
}

// Update оновлює маршрут
//
//	@Summary		Оновити маршрут
//	@Description	Оновлює існуючий маршрут за ID
//	@Tags			Routes
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int			true	"ID маршруту"
//	@Param			route	body		model.Route	true	"Оновлені дані маршруту"
//	@Success		200		{object}	model.Route
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/routes/{id} [put]
func (h *RouteHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid route ID"})
	}

	var route model.Route
	if err := c.BodyParser(&route); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	route.ID = id

	if err := h.routeService.Update(c.Context(), &route); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(route)
}

// Delete видаляє маршрут
//
//	@Summary		Видалити маршрут
//	@Description	Видаляє маршрут за ID
//	@Tags			Routes
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"ID маршруту"
//	@Success		204	"No Content"
//	@Failure 400 {object} ErrorResponse
//	@Failure 500 {object} ErrorResponse
//	@Security		BearerAuth
//	@Router			/routes/{id} [delete]
func (h *RouteHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid route ID"})
	}

	if err := h.routeService.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(204).Send(nil)
}

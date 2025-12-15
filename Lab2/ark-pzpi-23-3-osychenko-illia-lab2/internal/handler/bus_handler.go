package handler

import (
	"busoptima/internal/model"
	"busoptima/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type BusHandler struct {
	busService service.BusService
}

func NewBusHandler(busService service.BusService) *BusHandler {
	return &BusHandler{busService: busService}
}

// GetAll повертає список автобусів
//
//	@Summary		Отримати список автобусів
//	@Description	Повертає список всіх автобусів з можливістю фільтрації
//	@Tags			Buses
//	@Accept			json
//	@Produce		json
//	@Param			active_only	query		bool	false	"Тільки активні автобуси"
//	@Success		200			{array}		model.Bus
//	@Failure		500			{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/buses [get]
func (h *BusHandler) GetAll(c *fiber.Ctx) error {
	activeOnly := c.QueryBool("active_only", true)

	buses, err := h.busService.GetAll(c.Context(), activeOnly)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(buses)
}

// GetByID повертає автобус за ID
//
//	@Summary		Отримати автобус за ID
//	@Description	Повертає автобус за вказаним ідентифікатором
//	@Tags			Buses
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"ID автобуса"
//	@Success		200	{object}	model.Bus
//	@Failure		400	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/buses/{id} [get]
func (h *BusHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid bus ID"})
	}

	bus, err := h.busService.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Bus not found"})
	}

	return c.JSON(bus)
}

// Create створює новий автобус
//
//	@Summary		Створити новий автобус
//	@Description	Створює новий автобус в системі
//	@Tags			Buses
//	@Accept			json
//	@Produce		json
//	@Param			bus	body		model.Bus	true	"Дані автобуса"
//	@Success		201	{object}	model.Bus
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/buses [post]
func (h *BusHandler) Create(c *fiber.Ctx) error {
	var bus model.Bus
	if err := c.BodyParser(&bus); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	bus.IsActive = true

	if err := h.busService.Create(c.Context(), &bus); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(bus)
}

// Update оновлює автобус
//
//	@Summary		Оновити автобус
//	@Description	Оновлює існуючий автобус за ID
//	@Tags			Buses
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int			true	"ID автобуса"
//	@Param			bus	body		model.Bus	true	"Оновлені дані автобуса"
//	@Success		200	{object}	model.Bus
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/buses/{id} [put]
func (h *BusHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid bus ID"})
	}

	var bus model.Bus
	if err := c.BodyParser(&bus); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	bus.ID = id

	if err := h.busService.Update(c.Context(), &bus); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(bus)
}

// Delete видаляє автобус
//
//	@Summary		Видалити автобус
//	@Description	Видаляє автобус за ID
//	@Tags			Buses
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"ID автобуса"
//	@Success		204	"No Content"
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/buses/{id} [delete]
func (h *BusHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid bus ID"})
	}

	if err := h.busService.Delete(c.Context(), id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(204).Send(nil)
}

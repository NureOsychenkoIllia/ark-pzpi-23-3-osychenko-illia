package handler

import (
	"busoptima/internal/model"
	"busoptima/internal/service"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type IoTHandler struct {
	iotService service.IoTService
}

func NewIoTHandler(iotService service.IoTService) *IoTHandler {
	return &IoTHandler{iotService: iotService}
}

type SyncEventsRequest struct {
	TripID int64 `json:"trip_id"`
	Events []struct {
		LocalID             int     `json:"local_id"`
		EventType           string  `json:"event_type"`
		Timestamp           string  `json:"timestamp"`
		Latitude            float64 `json:"latitude"`
		Longitude           float64 `json:"longitude"`
		PassengerCountAfter int     `json:"passenger_count_after"`
	} `json:"events"`
}

// SyncEvents синхронізує події пасажирів від IoT-пристрою
//
//	@Summary		Синхронізація подій пасажирів
//	@Description	Отримує та зберігає події входу/виходу пасажирів від IoT-пристрою
//	@Tags			IoT
//	@Accept			json
//	@Produce		json
//	@Param			events	body		SyncEventsRequest	true	"Події пасажирів"
//	@Success		201		{object}	service.SyncEventsResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/iot/events [post]
func (h *IoTHandler) SyncEvents(c *fiber.Ctx) error {
	var req SyncEventsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Конвертуємо в модель
	events := make([]model.PassengerEvent, len(req.Events))
	for i, e := range req.Events {
		timestamp, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid timestamp format"})
		}

		events[i] = model.PassengerEvent{
			TripID:              req.TripID,
			EventType:           e.EventType,
			Timestamp:           timestamp,
			Latitude:            &e.Latitude,
			Longitude:           &e.Longitude,
			PassengerCountAfter: e.PassengerCountAfter,
			DeviceLocalID:       &e.LocalID,
		}
	}

	response, err := h.iotService.SyncEvents(c.Context(), req.TripID, events)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(response)
}

// SendPriceRecommendation отримує рекомендацію ціни від IoT-пристрою
//
//	@Summary		Отримати рекомендацію ціни
//	@Description	Отримує рекомендацію ціни від IoT-пристрою
//	@Tags			IoT
//	@Accept			json
//	@Produce		json
//	@Param			recommendation	body		map[string]interface{}	true	"Рекомендація ціни"
//	@Success		200				{object}	map[string]string
//	@Failure		400				{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/iot/price [post]
func (h *IoTHandler) SendPriceRecommendation(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Price recommendation received"})
}

// GetTripConfig повертає конфігурацію рейсу для IoT-пристрою
//
//	@Summary		Отримати конфігурацію рейсу
//	@Description	Повертає конфігурацію рейсу для IoT-пристрою
//	@Tags			IoT
//	@Accept			json
//	@Produce		json
//	@Param			tripId	path		int	true	"ID рейсу"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/iot/config/{tripId} [get]
func (h *IoTHandler) GetTripConfig(c *fiber.Ctx) error {
	tripID, err := strconv.ParseInt(c.Params("tripId"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid trip ID"})
	}

	config, err := h.iotService.GetTripConfig(c.Context(), tripID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Trip not found"})
	}

	return c.JSON(config)
}

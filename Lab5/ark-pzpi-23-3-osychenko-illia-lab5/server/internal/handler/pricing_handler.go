package handler

import (
	"busoptima/internal/service"
	"time"

	"github.com/gofiber/fiber/v2"
)

type PricingHandler struct {
	pricingService service.PricingService
}

func NewPricingHandler(pricingService service.PricingService) *PricingHandler {
	return &PricingHandler{pricingService: pricingService}
}

// CalculatePriceRequest структура запиту розрахунку ціни
type CalculatePriceRequest struct {
	BasePrice         float64   `json:"base_price" validate:"required,min=0" example:"200.00"`
	CurrentPassengers int       `json:"current_passengers" validate:"min=0" example:"25"`
	Capacity          int       `json:"capacity" validate:"required,min=1" example:"50"`
	DepartureTime     time.Time `json:"departure_time" validate:"required" example:"2025-12-15T08:00:00Z"`
}

// CalculatePrice розраховує рекомендовану ціну
//
//	@Summary		Розрахувати рекомендовану ціну
//	@Description	Розраховує динамічну ціну на основі завантаженості та часу відправлення
//	@Tags			Pricing
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CalculatePriceRequest	true	"Параметри для розрахунку ціни"
//	@Success		200		{object}	service.PriceRecommendation
//	@Failure		400		{object}	ErrorResponse
//	@Security		BearerAuth
//	@Router			/pricing/calculate [post]
func (h *PricingHandler) CalculatePrice(c *fiber.Ctx) error {
	var req CalculatePriceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	recommendation, err := h.pricingService.CalculatePrice(
		c.Context(),
		req.BasePrice,
		req.CurrentPassengers,
		req.Capacity,
		req.DepartureTime,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(recommendation)
}

// generatePricingRecommendation генерує текстову рекомендацію
func generatePricingRecommendation(rec *service.PriceRecommendation) string {
	changePerc := rec.PriceChangePerc

	switch {
	case changePerc <= -20:
		return "Значна знижка через низьку завантаженість"
	case changePerc <= -10:
		return "Помірна знижка для залучення пасажирів"
	case changePerc >= 30:
		return "Максимальне підвищення через високий попит"
	case changePerc >= 15:
		return "Підвищення ціни через високу завантаженість"
	case changePerc >= 5:
		return "Невелике підвищення ціни"
	case changePerc <= -5:
		return "Невелика знижка"
	default:
		return "Ціна залишається близькою до базової"
	}
}

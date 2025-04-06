package handler

import (
	"sweng-task/internal/model"
	"sweng-task/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// TrackingHandler handles HTTP requests related to tracking
type TrackingHandler struct {
	service *service.TrackingService
	log     *zap.SugaredLogger
}

// NewTrackingHandler creates a new TrackingHandler
func NewTrackingHandler(service *service.TrackingService, log *zap.SugaredLogger) *TrackingHandler {
	return &TrackingHandler{
		service: service,
		log:     log,
	}
}

// TrackingHandler tracks events
func (h *TrackingHandler) TrackEvent(c *fiber.Ctx) error {
	var input model.TrackingEvent
	err := c.BodyParser(&input)
	if err != nil {
		return BadRequestResponse(c, "Invalid request body", err.Error())
	}

	ok, err := h.service.RecordAdInteraction(input)
	if err != nil {
		return InternalServerErrorResponse(c, "Failed to track event", err.Error())
	}

	return c.Status(fiber.StatusAccepted).JSON(map[string]any{
		"success": ok,
	})
}

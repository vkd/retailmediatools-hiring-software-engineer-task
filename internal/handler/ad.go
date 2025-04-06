package handler

import (
	"sweng-task/internal/service"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// AdHandler handles HTTP requests related to ads
type AdHandler struct {
	service *service.AdService
	log     *zap.SugaredLogger
}

// NewAdHandler creates a new AdHandler
func NewAdHandler(service *service.AdService, log *zap.SugaredLogger) *AdHandler {
	return &AdHandler{
		service: service,
		log:     log,
	}
}

// AdHandler returns winning ads
func (h *AdHandler) GetWinningAds(c *fiber.Ctx) error {
	placement := c.Query("placement")
	if placement == "" {
		return BadRequestResponse(c, "'placement' query parameter is empty", nil)
	}

	limit := c.QueryInt("limit", 1)
	if limit < 1 || limit > 10 {
		return BadRequestResponse(c, "'limit' must me in the range [1-10]", nil)
	}

	category := c.Query("category")
	keyword := c.Query("keyword")

	ads, err := h.service.GetWinningAds(placement, category, keyword, limit)
	if err != nil {
		return InternalServerErrorResponse(c, "Failed to get winning ads", err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(ads)
}

package handler

import "github.com/gofiber/fiber/v2"

func ErrorResponse(c *fiber.Ctx, status int, message string, details interface{}) error {
	jsonResponse := fiber.Map{
		"code":    status,
		"message": "Missing line item ID",
	}
	if details != nil {
		jsonResponse["details"] = details
	}
	return c.Status(status).JSON(jsonResponse)
}

func BadRequestResponse(c *fiber.Ctx, message string, details interface{}) error {
	return ErrorResponse(c, fiber.StatusBadRequest, message, details)
}

func InternalServerErrorResponse(c *fiber.Ctx, message string, details interface{}) error {
	return ErrorResponse(c, fiber.StatusInternalServerError, message, details)
}

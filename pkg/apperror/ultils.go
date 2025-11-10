package apperror

import (
	"ai-learn-english/config"
	"ai-learn-english/pkg/apperror/status"
	"ai-learn-english/pkg/logger"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

// ErrorResponse is the standardized HTTP error payload
type ErrorResponse struct {
	Error     string `json:"error"`
	ErrorCode string `json:"error_code"`
}
type FiberSuccessMessage struct {
	Code       status.SuccessCode `json:"code"`
	Message    string             `json:"message"`
	TrackingID string             `json:"tracking_id"`
	Data       any                `json:"data"`
}

// WriteError logs a structured warning and returns a standardized JSON error
func WriteError(module config.Module, c fiber.Ctx, httpStatus int, code string, message string) error {
	logger.WithFields(map[string]interface{}{
		"status_code":   httpStatus,
		"error_code":    code,
		"error_message": message,
		"http_method":   c.Method(),
		"base_url":      c.BaseURL(),
		"path":          c.Path(),
		"url":           c.OriginalURL(),
		"ip":            c.IP(),
		"headers":       c.GetReqHeaders(),
		"body":          string(c.Body()),
	}).Warnf("http error")

	return c.Status(httpStatus).JSON(ErrorResponse{
		Error:     message,
		ErrorCode: code,
	})
}

// Shorthands for common error responses
func BadRequest(module config.Module, c fiber.Ctx, code status.ErrorCode, message string) error {
	error_code := fmt.Sprintf("AI-%d", code)
	return WriteError(module, c, fiber.StatusBadRequest, error_code, message)
}

// InternalError writes a structured warning and returns a standardized JSON error
func InternalError(module config.Module, c fiber.Ctx, err error) error {
	error_code := fmt.Sprintf("AI-%d", status.ErrorCodeInternal)
	return WriteError(module, c, fiber.StatusInternalServerError, error_code, err.Error())
}

// Success writes a standardized JSON success response
func Success(module config.Module, fiberCtx fiber.Ctx, response FiberSuccessMessage) error {
	return fiberCtx.Status(fiber.StatusOK).JSON(response)
}

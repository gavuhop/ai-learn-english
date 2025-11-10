package query

import (
	"ai-learn-english/config"
	corequery "ai-learn-english/internal/core/query"
	"ai-learn-english/pkg/apperror"
	"ai-learn-english/pkg/apperror/status"
	"context"
	"encoding/json"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func HandleQuery(c fiber.Ctx) error {
	trackingID := c.Get("X-Request-ID")

	var req corequery.Request
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return apperror.BadRequest(config.ModuleQuery, c, status.FileUploadInvalidRequestBody, err.Error())
	}
	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		return apperror.BadRequest(config.ModuleQuery, c, status.FileUploadMissingParams, "question is empty")
	}
	// Delegate to core
	resp, err := corequery.Run(context.Background(), req)
	if err != nil {
		return apperror.InternalError(config.ModuleQuery, c, err)
	}

	return apperror.Success(config.ModuleQuery, c, apperror.FiberSuccessMessage{
		Code:       status.OK,
		Message:    "query ok",
		TrackingID: trackingID,
		Data:       resp,
	})
}

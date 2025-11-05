package ingest

import (
	"ai-learn-english/internal/services/ingest"
	"ai-learn-english/pkg/apperror"
	"ai-learn-english/pkg/apperror/status"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type ingestResponse struct {
	DocID int64 `json:"doc_id"`
}

func HandleIngest(c fiber.Ctx) error {
	trackingID := c.Get("X-Request-ID")

	docIDStr := c.Params("docID")
	if docIDStr == "" {
		return apperror.BadRequest(c, status.FileUploadMissingParams, "docID is required")
	}
	docID, err := strconv.ParseInt(docIDStr, 10, 64)
	if err != nil {
		return apperror.BadRequest(c, status.FileUploadMissingParams, "invalid docID")
	}

	q := c.Query("force")
	force := q == "1" || q == "true" || q == "yes"

	// Fire and forget
	go ingest.RunIngestion(docID, force)

	return apperror.Success(c, apperror.FiberSuccessMessage{
		Code:       status.OK,
		Message:    "ingest started",
		TrackingID: trackingID,
		Data:       ingestResponse{DocID: docID},
	})
}

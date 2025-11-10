package retriever

import (
	"context"
	"strconv"
	"strings"
	"time"

	"ai-learn-english/config"
	"ai-learn-english/internal/core/retriever"
	"ai-learn-english/pkg/apperror"
	"ai-learn-english/pkg/apperror/status"

	"github.com/gofiber/fiber/v3"
)

type searchResponse struct {
	Hits []retriever.Hit `json:"hits"`
}

func HandleSearch(c fiber.Ctx) error {
	trackingID := c.Get("X-Request-ID")

	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		return apperror.BadRequest(config.ModuleRetriever, c, status.FileUploadMissingParams, "q is required")
	}
	topKStr := c.Query("top_k")
	topK := 8
	if topKStr != "" {
		if v, err := strconv.Atoi(topKStr); err == nil && v > 0 && v <= 64 {
			topK = v
		}
	}
	var docIDs []int64
	if ids := strings.TrimSpace(c.Query("doc_ids")); ids != "" {
		parts := strings.Split(ids, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if id, err := strconv.ParseInt(p, 10, 64); err == nil {
				docIDs = append(docIDs, id)
			}
		}
	}

	// Embed with a longer timeout (network call), e.g., 3s
	embedCtx, cancelEmbed := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelEmbed()
	vec, err := retriever.EmbedQuestion(embedCtx, q)
	if err != nil {
		return apperror.InternalError(config.ModuleRetriever, c, err)
	}
	// Search with slightly higher timeout to account for initial collection load (1s)
	searchCtx, cancelSearch := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelSearch()
	hits, err := retriever.SearchMilvus(searchCtx, vec, topK, retriever.Filters{DocIDs: docIDs})
	if err != nil {
		return apperror.InternalError(config.ModuleRetriever, c, err)
	}

	return apperror.Success(config.ModuleRetriever, c, apperror.FiberSuccessMessage{
		Code:       status.OK,
		Message:    "search ok",
		TrackingID: trackingID,
		Data:       searchResponse{Hits: hits},
	})
}

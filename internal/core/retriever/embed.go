package retriever

import (
	"context"
	"errors"

	"ai-learn-english/config"
	"ai-learn-english/internal/core/ingest"
	"ai-learn-english/pkg/logger"
)

// EmbedQuestion embeds a single question string and returns its vector.
func EmbedQuestion(ctx context.Context, question string) ([]float32, error) {
	if question == "" {
		return nil, errors.New("question is empty")
	}
	vecs, err := ingest.EmbedOpenAI(ctx, []string{question})
	if err != nil {
		logger.Error(err, "%v: embed question failed: %s", config.ModuleRetriever, question)
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, errors.New("no embedding returned")
	}
	return vecs[0], nil
}

package ingest

import (
	"ai-learn-english/config"
	"ai-learn-english/pkg/logger"
	"context"
	"errors"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// EmbedOpenAI calls OpenAI embeddings for given inputs and returns vectors.
func EmbedOpenAI(ctx context.Context, inputs []string) ([][]float32, error) {
	if len(inputs) == 0 {
		return [][]float32{}, nil
	}
	key := config.Cfg.OpenAI.Key
	if key == "" {
		return nil, errors.New("missing openai key")
	}
	// Batch in chunks of up to 100 inputs
	var all [][]float32
	for i := 0; i < len(inputs); i += 100 {
		j := i + 100
		if j > len(inputs) {
			j = len(inputs)
		}
		batch := inputs[i:j]
		logger.WithFields(map[string]interface{}{
			"model":       config.Cfg.OpenAI.EmbeddingModel,
			"batch_start": i,
			"batch_end":   j,
			"batch_size":  len(batch),
		}).Info("openai: embedding batch start")

		vectors, err := embedBatch(ctx, key, batch)
		if err != nil {
			logger.WithFields(map[string]interface{}{
				"model":       config.Cfg.OpenAI.EmbeddingModel,
				"batch_start": i,
				"batch_end":   j,
				"error":       err,
			}).Errorf("openai: embedding batch failed")
			return nil, err
		}
		logger.WithFields(map[string]interface{}{
			"batch_start": i,
			"batch_end":   j,
			"vectors":     len(vectors),
		}).Info("openai: embedding batch done")
		all = append(all, vectors...)
	}
	return all, nil
}

// Retries are disabled; calls will be attempted only once per batch.

func embedBatch(ctx context.Context, apiKey string, batch []string) ([][]float32, error) {
	// Use official OpenAI Go SDK and disable automatic retries
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	// Build request params and call embeddings endpoint via SDK
	reqBody := openAIEmbeddingRequest{Model: config.Cfg.OpenAI.EmbeddingModel, Input: batch}
	var out openAIEmbeddingResponse
	if err := client.Post(ctx, "/embeddings", reqBody, &out); err != nil {
		return nil, err
	}
	if out.Error != nil {
		return nil, errors.New(out.Error.Message)
	}
	vectors := make([][]float32, len(out.Data))
	for i := range out.Data {
		src := out.Data[i].Embedding
		vec := make([]float32, len(src))
		for k := range src {
			vec[k] = float32(src[k])
		}
		vectors[i] = vec
	}
	return vectors, nil
}

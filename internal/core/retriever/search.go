package retriever

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ai-learn-english/config"
	"ai-learn-english/pkg/logger"

	milvusclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	milvusentity "github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// SearchMilvus performs a vector similarity search and returns topK hits with metadata.
func SearchMilvus(ctx context.Context, query []float32, topK int, filters Filters) ([]Hit, error) {
	if topK <= 0 {
		topK = 8
	}
	if len(query) == 0 {
		return []Hit{}, nil
	}
	// Guard the search by a short timeout to keep latency bounds tight.
	timeout := 200 * time.Millisecond
	var cancel context.CancelFunc
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cli, err := milvusclient.NewClient(ctx, milvusclient.Config{Address: config.Cfg.Milvus.Address})
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	collection := config.Cfg.Milvus.Collection
	if collection == "" {
		collection = "chunks"
	}

	// Ensure collection exists then load
	exists, err := cli.HasCollection(ctx, collection)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("collection %q not found", collection)
	}
	if err := cli.LoadCollection(ctx, collection, false); err != nil {
		return nil, err
	}

	// Milvus search setup
	metricType := milvusentity.MetricType(config.Cfg.Milvus.IndexHNSWConfig.MetricType)
	// Favor low latency locally; tune within 64–128 range
	ef := 64
	searchParam, err := milvusentity.NewIndexHNSWSearchParam(ef)
	if err != nil {
		return nil, err
	}

	// Filter expression
	expr := buildExpr(filters)
	outputFields := []string{"id", "doc_id", "chunk_index", "page_index", "content"}
	var vectors []milvusentity.Vector
	vectors = append(vectors, milvusentity.FloatVector(query))

	start := time.Now()
	results, err := cli.Search(
		ctx,
		collection,
		nil, // partitions
		expr,
		outputFields,
		vectors,
		"embedding",
		metricType,
		topK,
		searchParam,
	)
	elapsed := time.Since(start)

	if err != nil {
		logger.Error(err, "%v: milvus search failed: %s", config.ModuleRetriever, err.Error())
		return nil, err
	}
	logger.Info("%v: milvus search done: %s", config.ModuleRetriever, elapsed.Milliseconds())

	// Parse results
	if len(results) == 0 {
		return []Hit{}, nil
	}
	it := results[0]

	hits := make([]Hit, 0, it.ResultCount)
	for i := 0; i < it.ResultCount; i++ {
		var h Hit
		h.ChunkID = it.IDs.(*milvusentity.ColumnInt64).Data()[i]
		h.Score = float32(it.Scores[i])

		// Extract fields
		for _, field := range it.Fields {
			switch col := field.(type) {
			case *milvusentity.ColumnInt64:
				switch col.Name() {
				case "doc_id":
					h.DocID = col.Data()[i]
				}
			case *milvusentity.ColumnInt32:
				switch col.Name() {
				case "page_index":
					h.PageIndex = col.Data()[i]
				case "chunk_index":
					h.ChunkIndex = col.Data()[i]
				}
			case *milvusentity.ColumnVarChar:
				if col.Name() == "content" {
					h.Content = col.Data()[i]
				}
			}
		}
		hits = append(hits, h)
	}
	return hits, nil
}

func buildExpr(f Filters) string {
	if len(f.DocIDs) == 0 {
		return ""
	}
	// doc_id in [1,2,3]
	var b strings.Builder
	b.WriteString("doc_id in [")
	for i, id := range f.DocIDs {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(fmt.Sprintf("%d", id))
	}
	b.WriteByte(']')
	return b.String()
}

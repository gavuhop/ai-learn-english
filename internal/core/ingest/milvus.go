package ingest

import (
	"ai-learn-english/config"
	"ai-learn-english/pkg/logger"
	"context"

	milvusclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	milvusentity "github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const milvusVectorDim = 1536
const milvusContentMaxLen = 8192
const moduleName = string(config.ModuleMilvus)

// UpsertMilvusVectors ensures collection and inserts embeddings. Returns IDs and collection.
func UpsertMilvusVectors(ctx context.Context, vectors [][]float32, docID int64, chunks []Chunk) ([]int64, string, error) {
	cli, err := milvusclient.NewClient(ctx, milvusclient.Config{Address: config.Cfg.Milvus.Address})
	if err != nil {
		return nil, "", err
	}
	defer cli.Close()

	// Determine collection
	collection := config.Cfg.Milvus.Collection
	if collection == "" {
		collection = "chunks"
	}
	// Ensure collection
	exists, err := cli.HasCollection(ctx, collection)
	if err != nil {
		return nil, "", err
	}
	if !exists {
		if err := createChunksCollection(ctx, cli, collection); err != nil {
			return nil, "", err
		}
	}

	// Prepare columns
	docIDs := make([]int64, len(chunks))
	chunkIdxs := make([]int32, len(chunks))
	pageIdxs := make([]int32, len(chunks))
	contents := make([]string, len(chunks))
	for i, ch := range chunks {
		docIDs[i] = docID
		chunkIdxs[i] = ch.ChunkIndex
		pageIdxs[i] = ch.PageIndex
		contents[i] = ch.Content
	}

	// Deterministic primary keys from docID and chunkIndex to avoid AutoID API differences
	ids := make([]int64, len(chunks))
	for i := range chunks {
		ids[i] = (docID << 20) + int64(chunks[i].ChunkIndex)
	}
	colID := milvusentity.NewColumnInt64("id", ids)
	colDoc := milvusentity.NewColumnInt64("doc_id", docIDs)
	colChunk := milvusentity.NewColumnInt32("chunk_index", chunkIdxs)
	colPage := milvusentity.NewColumnInt32("page_index", pageIdxs)
	colContent := milvusentity.NewColumnVarChar("content", contents)
	colVec := milvusentity.NewColumnFloatVector("embedding", milvusVectorDim, vectors)

	if _, err := cli.Insert(ctx, collection, "", colID, colDoc, colChunk, colPage, colContent, colVec); err != nil {
		return nil, "", err
	}
	logger.WithFields(map[string]interface{}{
		"collection": collection,
		"rows":       len(chunks),
	}).Infof("%s insert done", moduleName)
	return ids, collection, nil
}

func createChunksCollection(ctx context.Context, cli milvusclient.Client, collection string) error {
	schema := milvusentity.NewSchema().WithName(collection).WithDescription("chunks")
	// Primary key (no AutoID) – we will provide IDs
	schema.WithField(milvusentity.NewField().WithName("id").WithDataType(milvusentity.FieldTypeInt64).WithIsPrimaryKey(true))
	schema.WithField(milvusentity.NewField().WithName("doc_id").WithDataType(milvusentity.FieldTypeInt64))
	schema.WithField(milvusentity.NewField().WithName("chunk_index").WithDataType(milvusentity.FieldTypeInt32))
	schema.WithField(milvusentity.NewField().WithName("page_index").WithDataType(milvusentity.FieldTypeInt32))
	schema.WithField(milvusentity.NewField().WithName("content").WithDataType(milvusentity.FieldTypeVarChar).WithMaxLength(milvusContentMaxLen))
	schema.WithField(milvusentity.NewField().WithName("embedding").WithDataType(milvusentity.FieldTypeFloatVector).WithDim(milvusVectorDim))

	if err := cli.CreateCollection(ctx, schema, 2); err != nil {
		return err
	}

	metricType := config.Cfg.Milvus.IndexHNSWConfig.MetricType
	m := config.Cfg.Milvus.IndexHNSWConfig.M
	efConstruction := config.Cfg.Milvus.IndexHNSWConfig.EfConstruction

	index, err := milvusentity.NewIndexHNSW(milvusentity.MetricType(metricType), m, efConstruction)
	if err != nil {
		logger.Errorf("%s failed to create index params: %v", moduleName, err)
		return err
	}
	if err := cli.CreateIndex(ctx, collection, "embedding", index, false); err != nil {
		logger.Errorf("%s failed to create index: %v", moduleName, err)
		return err
	}

	logger.WithFields(map[string]interface{}{
		"collection":      collection,
		"field":           "embedding",
		"metric_type":     metricType,
		"m":               m,
		"ef_construction": efConstruction,
	}).Infof("%s index created successfully", moduleName)

	return nil
}
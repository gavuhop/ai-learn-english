package ingest

import (
	"ai-learn-english/config"
	"context"

	milvusclient "github.com/milvus-io/milvus-sdk-go/v2/client"
	milvusentity "github.com/milvus-io/milvus-sdk-go/v2/entity"
)

const milvusVectorDim = 1536

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
	for i, ch := range chunks {
		docIDs[i] = docID
		chunkIdxs[i] = ch.ChunkIndex
	}

	// Deterministic primary keys from docID and chunkIndex to avoid AutoID API differences
	ids := make([]int64, len(chunks))
	for i := range chunks {
		ids[i] = (docID << 20) + int64(chunks[i].ChunkIndex)
	}
	colID := milvusentity.NewColumnInt64("id", ids)
	colDoc := milvusentity.NewColumnInt64("doc_id", docIDs)
	colChunk := milvusentity.NewColumnInt32("chunk_index", chunkIdxs)
	colVec := milvusentity.NewColumnFloatVector("embedding", milvusVectorDim, vectors)

	if _, err := cli.Insert(ctx, collection, "", colID, colDoc, colChunk, colVec); err != nil {
		return nil, "", err
	}
	return ids, collection, nil
}

func createChunksCollection(ctx context.Context, cli milvusclient.Client, collection string) error {
	schema := milvusentity.NewSchema().WithName(collection).WithDescription("chunks")
	// Primary key (no AutoID) – we will provide IDs
	schema.WithField(milvusentity.NewField().WithName("id").WithDataType(milvusentity.FieldTypeInt64).WithIsPrimaryKey(true))
	schema.WithField(milvusentity.NewField().WithName("doc_id").WithDataType(milvusentity.FieldTypeInt64))
	schema.WithField(milvusentity.NewField().WithName("chunk_index").WithDataType(milvusentity.FieldTypeInt32))
	schema.WithField(milvusentity.NewField().WithName("embedding").WithDataType(milvusentity.FieldTypeFloatVector).WithDim(milvusVectorDim))

	if err := cli.CreateCollection(ctx, schema, 2); err != nil {
		return err
	}

	// Optional: index creation can be added later
	return nil
}

package ingest

import (
	"ai-learn-english/config"
	"ai-learn-english/internal/database"
	"ai-learn-english/pkg/logger"
	"context"
	"errors"
	"time"
)

// RunIngestion orchestrates the ingestion pipeline for a document ID.
func RunIngestion(docID int64, force bool) {
	db, err := database.GetDB()
	if err != nil {
		logger.Error(err, "ingest: db unavailable")
		return
	}

	// Load document
	doc, err := GetDocumentByID(db, docID)
	if err != nil {
		logger.Error(err, "ingest: get document failed")
		return
	}
	if doc == nil {
		logger.Error(errors.New("not found"), "ingest: document not found")
		return
	}
	logger.WithFields(map[string]interface{}{
		"doc_id":    docID,
		"file_path": *doc.FilePath,
	}).Info("ingest: start")

	// Idempotency
	exists, err := HasChunks(db, docID)
	if err != nil {
		logger.Error(err, "ingest: check chunks failed")
		return
	}
	if exists && !force {
		logger.Info("ingest: chunks already exist; skip (no force)")
		return
	}
	if exists && force {
		if err := DeleteChunksByDocID(db, docID); err != nil {
			logger.Error(err, "ingest: cleanup chunks failed")
			return
		}
		// We do not delete vectors in Milvus for POC; future work can track and delete by doc_id.
	}

	// Mark processing
	_ = UpdateDocumentStatus(db, docID, "processing")

	// Fetch PDF to local temp path
	tmpPath, cleanup, err := FetchToLocalTemp(*doc.FilePath)
	if err != nil {
		logger.Error(err, "ingest: fetch file failed")
		_ = UpdateDocumentStatus(db, docID, "failed")
		return
	}
	defer cleanup()

	// Extract text pages
	pages, err := ExtractPDFTextPages(tmpPath)
	if err != nil {
		logger.Error(err, "ingest: extract text failed")
		_ = UpdateDocumentStatus(db, docID, "failed")
		return
	}
	logger.WithFields(map[string]interface{}{
		"doc_id": docID,
		"pages":  len(pages),
	}).Info("ingest: extracted pages")

	// Chunking
	targetTokens := config.Cfg.Ingest.ChunkTokens
	if targetTokens <= 0 {
		targetTokens = 600
	}
	overlap := config.Cfg.Ingest.ChunkOverlap
	if overlap < 0 {
		overlap = 80
	}
	chunks := BuildChunks(pages, targetTokens, overlap)
	logger.WithFields(map[string]interface{}{
		"doc_id":       docID,
		"chunks":       len(chunks),
		"chunk_tokens": targetTokens,
		"overlap":      overlap,
	}).Info("ingest: chunks built")

	// Prepare embedding inputs
	inputs := make([]string, 0, len(chunks))
	for _, ch := range chunks {
		inputs = append(inputs, ch.Content)
	}

	// Embed
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	vectors, err := EmbedOpenAI(ctx, inputs)
	if err != nil {
		logger.Error(err, "ingest: embedding failed")
		_ = UpdateDocumentStatus(db, docID, "failed")
		return
	}
	if len(vectors) != len(chunks) {
		logger.Error(errors.New("mismatch"), "ingest: embedding count mismatch")
		_ = UpdateDocumentStatus(db, docID, "failed")
		return
	}

	// Upsert into Milvus
	milvusIDs, collection, err := UpsertMilvusVectors(ctx, vectors, docID, chunks)
	if err != nil {
		logger.Error(err, "ingest: milvus upsert failed")
		_ = UpdateDocumentStatus(db, docID, "failed")
		return
	}

	// Persist chunks into MySQL
	if err := InsertChunks(db, docID, chunks, milvusIDs, collection); err != nil {
		logger.Error(err, "ingest: db insert chunks failed")
		_ = UpdateDocumentStatus(db, docID, "failed")
		return
	}

	// Done
	_ = UpdateDocumentStatus(db, docID, "ready")
}

package ingest

import (
	"ai-learn-english/internal/database/model"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

func GetDocumentByID(db *gorm.DB, docID int64) (*model.Document, error) {
	var doc model.Document
	if err := db.First(&doc, docID).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func HasChunks(db *gorm.DB, docID int64) (bool, error) {
	var count int64
	if err := db.Model(&model.Chunk{}).Where("document_id = ?", docID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func DeleteChunksByDocID(db *gorm.DB, docID int64) error {
	return db.Where("document_id = ?", docID).Delete(&model.Chunk{}).Error
}

func UpdateDocumentStatus(db *gorm.DB, docID int64, status string) error {
	return db.Exec("UPDATE documents SET status=? WHERE id=?", status, docID).Error
}

func InsertChunks(db *gorm.DB, docID int64, chunks []Chunk, milvusIDs []int64, collection string) error {
	records := make([]model.Chunk, 0, len(chunks))
	for i, ch := range chunks {
		content := ch.Content
		contentPreview := buildContentPreview(content, 512)
		h := sha256.Sum256([]byte(content))
		hash := hex.EncodeToString(h[:])
		var milvusID int64
		if i < len(milvusIDs) {
			milvusID = milvusIDs[i]
		}
		records = append(records, model.Chunk{
			DocumentID:       docID,
			ChunkIndex:       ch.ChunkIndex,
			PageIndex:        &ch.PageIndex,
			Content:          content,
			ContentPreview:   &contentPreview,
			TokenCount:       nil,
			MilvusCollection: collection,
			MilvusID:         milvusID,
			ContentHash:      hash,
		})
	}
	return db.Create(&records).Error
}

// buildContentPreview sanitizes the preview to valid UTF-8 printable characters
// and truncates by runes to avoid splitting multi-byte sequences.
func buildContentPreview(s string, maxRunes int) string {
	// Remove BOM and control characters except common whitespace
	var b strings.Builder
	b.Grow(len(s))
	count := 0
	for _, r := range s {
		if r == '\uFEFF' { // BOM
			continue
		}
		if r == '\n' || r == '\t' || r == '\r' {
			// keep common whitespace
		} else if !unicode.IsPrint(r) {
			// skip non-printable
			continue
		}
		b.WriteRune(r)
		count++
		if count >= maxRunes {
			break
		}
	}
	out := b.String()
	// Collapse excessive spaces and trim
	out = strings.TrimSpace(out)
	return out
}

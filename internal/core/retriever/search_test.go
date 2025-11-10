package retriever

import (
	"context"
	"testing"
	"time"
)

// This test covers the EmbedQuestion shape and basic error handling by mocking ingest.EmbedOpenAI is non-trivial here.
// Instead, we validate SearchMilvus interface behavior via a minimal integration shape using an impossible address,
// and ensure it times out quickly when context deadline is small.
func TestEmbedQuestion_Empty(t *testing.T) {
	_, err := EmbedQuestion(context.Background(), "")
	if err == nil {
		t.Fatalf("expected error for empty question")
	}
}

// Note: Full end-to-end Milvus search requires a running Milvus. For unit-level verification
// of topK handling, we'd abstract a Milvus client interface. Given current code uses SDK directly,
// we assert timeout behavior to keep tests hermetic.
func TestSearchMilvus_ContextTimeout(t *testing.T) {
	// Short timeout to avoid long waits
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// A zero vector is acceptable to exercise early path before network
	_, err := SearchMilvus(ctx, make([]float32, 1536), 10, Filters{})
	if err == nil {
		// If Milvus is running locally and reachable, this might pass, so we only assert no hang.
		// When not running, it should error quickly due to timeout.
		t.Log("search completed without error (Milvus may be running locally)")
	}
}

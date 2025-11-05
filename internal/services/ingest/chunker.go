package ingest

import (
	"strings"
)

type Chunk struct {
	ChunkIndex int32
	PageIndex  int32
	Content    string
}

// BuildChunks makes ~token-sized chunks with overlap from page texts.
// Token approximation: ~4 chars per token; this is coarse but adequate for POC.
func BuildChunks(pages []string, targetTokens int, overlapTokens int) []Chunk {
	if targetTokens <= 0 {
		targetTokens = 600
	}
	if overlapTokens < 0 {
		overlapTokens = 0
	}
	targetChars := targetTokens * 4
	overlapChars := overlapTokens * 4

	chunks := make([]Chunk, 0, 128)
	chunkIdx := int32(0)
	for pageIdx, page := range pages {
		text := strings.TrimSpace(page)
		if text == "" {
			continue
		}
		runes := []rune(text)
		for startRune := 0; startRune < len(runes); {
			endRune := startRune + targetChars
			if endRune > len(runes) {
				endRune = len(runes)
			}
			chunk := string(runes[startRune:endRune])
			chunks = append(chunks, Chunk{
				ChunkIndex: chunkIdx,
				PageIndex:  int32(pageIdx + 1),
				Content:    chunk,
			})
			chunkIdx++
			if endRune == len(runes) {
				break
			}
			// Advance with overlap (by runes)
			nextStartRune := endRune - overlapChars
			if nextStartRune <= startRune {
				nextStartRune = endRune
			}
			startRune = nextStartRune
		}
	}
	return chunks
}

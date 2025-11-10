package retriever

// Filters represents optional constraints applied during search.
// Currently supports filtering by a set of document IDs.
type Filters struct {
	DocIDs []int64
}

// Hit represents a single search result from Milvus with associated metadata.
type Hit struct {
	ChunkID    int64
	Score      float32
	DocID      int64
	PageIndex  int32
	ChunkIndex int32
	Content    string
}

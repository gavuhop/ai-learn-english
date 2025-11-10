package query

type Request struct {
	Question string  `json:"question"`
	DocIDs   []int64 `json:"doc_ids"`
	TopK     int     `json:"top_k"`
}

type ContextSnippet struct {
	DocID   int64  `json:"doc_id"`
	Page    int32  `json:"page"`
	Snippet string `json:"snippet"`
}

type Response struct {
	Answer   string           `json:"answer"`
	Contexts []ContextSnippet `json:"contexts"`
}

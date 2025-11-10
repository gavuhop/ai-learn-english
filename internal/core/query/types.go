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

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float32       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}
type chatChoice struct {
	Index   int `json:"index"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}
type chatResponse struct {
	Choices []chatChoice `json:"choices"`
}

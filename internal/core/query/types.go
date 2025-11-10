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
	Learning *LearningAnswer  `json:"learning,omitempty"`
}

// Learning-oriented structured output
type LearningAnswer struct {
	ShortAnswer  string             `json:"short_answer"`
	Wordlist     []WordItem         `json:"wordlist"`
	Phrases      []string           `json:"phrases"`
	Examples     []BilingualExample `json:"examples"`
	Exercises    []FillBlank        `json:"exercises"`
	Insufficient bool               `json:"insufficient"`
}

type WordItem struct {
	En  string `json:"en"`
	Vi  string `json:"vi"`
	Def string `json:"def"`
}

type BilingualExample struct {
	En string `json:"en"`
	Vi string `json:"vi"`
}

type FillBlank struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
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

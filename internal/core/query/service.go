package query

import (
	"ai-learn-english/config"
	"ai-learn-english/internal/api/upload"
	"ai-learn-english/internal/core/retriever"
	"ai-learn-english/internal/database"
	"ai-learn-english/internal/database/model"
	"ai-learn-english/pkg/logger"
	"context"
	"fmt"
	"strings"
	"time"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)


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
// Run executes the query flow: embed → search → prompt → LLM → persist.
func Run(ctx context.Context, req Request) (Response, error) {
	if req.TopK <= 0 || req.TopK > 64 {
		req.TopK = 12
	}
	// Embed
	embedCtx, cancelEmbed := context.WithTimeout(ctx, 3*time.Second)
	defer cancelEmbed()
	vec, err := retriever.EmbedQuestion(embedCtx, strings.TrimSpace(req.Question))
	if err != nil {
		logger.Error(err, "%v: embed question failed", config.ModuleQuery)
		return Response{}, err
	}
	// Search
	searchCtx, cancelSearch := context.WithTimeout(ctx, 1*time.Second)
	defer cancelSearch()
	hits, err := retriever.SearchMilvus(searchCtx, vec, req.TopK, retriever.Filters{DocIDs: req.DocIDs})
	if err != nil {
		logger.Error(err, "%v: search milvus failed", config.ModuleQuery)
		return Response{}, err
	}
	// Build contexts
	ctxs := make([]ContextSnippet, 0, len(hits))
	for _, h := range hits {
		ctxs = append(ctxs, ContextSnippet{
			DocID:   h.DocID,
			Page:    h.PageIndex,
			Snippet: h.Content,
		})
	}
	// Guard hallucination
	if len(ctxs) == 0 {
		answer := "Chưa đủ bằng chứng để trả lời từ tài liệu."
		if err := persistMessages(req.Question, answer, nil); err != nil {
			logger.Error(err, "%v: persist messages failed", config.ModuleQuery)
		}
		return Response{Answer: answer, Contexts: []ContextSnippet{}}, nil
	}
	// Prompt + LLM
	sysMsg, userMsg := buildPrompt(req.Question, ctxs)
	llmCtx, cancelLLM := context.WithTimeout(ctx, 10*time.Second)
	defer cancelLLM()
	answer, err := callLLM(llmCtx, sysMsg, userMsg)
	if err != nil {
		logger.Error(err, "%v: call llm failed", config.ModuleQuery)
		return Response{}, err
	}
	// Persist
	if err := persistMessages(req.Question, answer, ctxs); err != nil {
		logger.Error(err, "%v: persist messages failed", config.ModuleQuery)
	}
	return Response{Answer: answer, Contexts: ctxs}, nil
}

func buildPrompt(question string, ctxs []ContextSnippet) (systemMsg, userMsg string) {
	var b strings.Builder
	b.WriteString("Bạn là gia sư tiếng Anh. Hãy trả lời ngắn gọn, dễ hiểu, song ngữ (VI trước, English simplified sau). ")
	b.WriteString("Chỉ dựa trên các trích đoạn (contexts) bên dưới. Nếu không đủ bằng chứng, trả: \"Chưa đủ bằng chứng để trả lời từ tài liệu.\".\n\n")
	b.WriteString("Contexts:\n")
	for i, c := range ctxs {
		b.WriteString(fmt.Sprintf("[%d] (doc_id=%d, page=%d): %s\n\n", i+1, c.DocID, c.Page, sanitize(c.Snippet)))
	}
	systemMsg = b.String()
	userMsg = fmt.Sprintf("Câu hỏi: %s\nYêu cầu: trả lời ngắn gọn, song ngữ, có thể trích dẫn ngắn từ contexts nếu cần.", question)
	return
}

func sanitize(s string) string {
	out := strings.ReplaceAll(s, "\x00", "")
	return strings.TrimSpace(out)
}

func callLLM(ctx context.Context, promptSystem, promptUser string) (string, error) {
	client := openai.NewClient(option.WithAPIKey(config.Cfg.OpenAI.Key))
	req := chatRequest{
		Model:       config.Cfg.OpenAI.Model,
		Temperature: 0.2,
		MaxTokens:   512,
		Messages: []chatMessage{
			{Role: "system", Content: promptSystem},
			{Role: "user", Content: promptUser},
		},
	}
	var out chatResponse
	if err := client.Post(ctx, "/chat/completions", req, &out); err != nil {
		logger.Error(err, "%v: call llm failed", config.ModuleQuery)
		return "", err
	}
	if len(out.Choices) == 0 {
		logger.Error(fmt.Errorf("no choices returned"), "%v: no choices returned", config.ModuleQuery)
		return "", fmt.Errorf("no choices returned")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func persistMessages(question string, answer string, ctxs []ContextSnippet) error {
	db, err := database.GetDB()
	if err != nil {
		logger.Error(err, "%v: get db failed", config.ModuleQuery)
		return err
	}
	userID, err := upload.EnsureDefaultUser(db)
	if err != nil {
		logger.Error(err, "%v: ensure default user failed", config.ModuleQuery)
		return err
	}
	now := time.Now()
	msgUser := model.Message{
		UserID:    userID,
		Role:      "user",
		Content:   question,
		CreatedAt: &now,
	}
	if err := db.Create(&msgUser).Error; err != nil {
		logger.Error(err, "%v: persist messages failed", config.ModuleQuery)
		return err
	}
	msgAssistant := model.Message{
		UserID:    userID,
		Role:      "assistant",
		Content:   answer,
		CreatedAt: &now,
	}
	if err := db.Create(&msgAssistant).Error; err != nil {
		logger.Error(err, "%v: persist messages failed", config.ModuleQuery)
		return err
	}
	for _, cs := range ctxs {
		content := cs.Snippet
		var docID *int64
		docID = &cs.DocID
		msgCtx := model.Message{
			UserID:     userID,
			Role:       "context",
			Content:    content,
			DocumentID: docID,
			CreatedAt:  &now,
		}
		if err := db.Create(&msgCtx).Error; err != nil {
			logger.Error(err, "%v: persist messages failed", config.ModuleQuery)
			return err
		}
	}
	return nil
}

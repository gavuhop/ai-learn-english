package query

import (
	"ai-learn-english/config"
	"ai-learn-english/internal/api/upload"
	"ai-learn-english/internal/core/retriever"
	"ai-learn-english/internal/database"
	"ai-learn-english/internal/database/model"
	"ai-learn-english/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

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
		// learning struct on insufficient
		la := &LearningAnswer{
			ShortAnswer:  answer,
			Wordlist:     []WordItem{},
			Phrases:      []string{},
			Examples:     []BilingualExample{},
			Exercises:    []FillBlank{},
			Insufficient: true,
		}
		if err := persistMessages(req.Question, answer, nil); err != nil {
			logger.Error(err, "%v: persist messages failed", config.ModuleQuery)
		}
		return Response{Answer: answer, Contexts: []ContextSnippet{}, Learning: la}, nil
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
	// Try parse structured learning JSON
	var learning LearningAnswer
	var parsed bool
	if e := json.Unmarshal([]byte(answer), &learning); e == nil && (learning.ShortAnswer != "" || learning.Insufficient) {
		parsed = true
	}
	// Compose legacy answer for backward compat
	legacy := answer
	if parsed {
		legacy = composeLegacyAnswer(learning)
	}
	// Persist legacy
	if err := persistMessages(req.Question, legacy, ctxs); err != nil {
		logger.Error(err, "%v: persist messages failed", config.ModuleQuery)
	}
	var laPtr *LearningAnswer
	if parsed {
		laPtr = &learning
	}
	return Response{Answer: legacy, Contexts: ctxs, Learning: laPtr}, nil
}

func buildPrompt(question string, ctxs []ContextSnippet) (systemMsg, userMsg string) {
	var b strings.Builder
	b.WriteString("Bạn là gia sư tiếng Anh. Trả lời song ngữ (VI trước, English simplified sau). ")
	b.WriteString("Chỉ dựa trên các trích đoạn (contexts) bên dưới. ")
	b.WriteString("Luôn xuất ra JSON DUY NHẤT theo schema bên dưới, không thêm ký tự ngoài JSON. ")
	b.WriteString("Nếu không đủ bằng chứng, set \"insufficient\": true, short_answer=\"Chưa đủ bằng chứng để trả lời từ tài liệu.\", các mảng rỗng.\n\n")
	b.WriteString("Contexts:\n")
	for i, c := range ctxs {
		b.WriteString(fmt.Sprintf("[%d] (doc_id=%d, page=%d): %s\n\n", i+1, c.DocID, c.Page, sanitize(c.Snippet)))
	}
	systemMsg = b.String()
	var sb strings.Builder
	sb.WriteString("Câu hỏi: ")
	sb.WriteString(question)
	sb.WriteString("\nHãy trả về JSON đúng schema và ràng buộc:\n")
	sb.WriteString(`{
  "short_answer": "string (VI trước, rồi English simplified, <=3 câu)",
  "wordlist": [{"en": "string", "vi": "string", "def": "short english definition"}],  // 5-8 mục
  "phrases": ["string"],                             // 3-5 mục (cấu trúc/cụm từ)
  "examples": [{"en": "string", "vi": "string"}],   // 2-3 ví dụ mới
  "exercises": [{"question": "câu có ___", "answer": "từ cần điền"}], // 1-2 bài tập
  "insufficient": false
}`)
	userMsg = sb.String()
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

func composeLegacyAnswer(la LearningAnswer) string {
	var b strings.Builder
	// 1) short answer
	if la.ShortAnswer != "" {
		b.WriteString("1) ")
		b.WriteString(la.ShortAnswer)
		b.WriteString("\n")
	}
	// 2) wordlist
	if len(la.Wordlist) > 0 {
		b.WriteString("2) Wordlist:\n")
		for _, w := range la.Wordlist {
			b.WriteString("- ")
			b.WriteString(w.En)
			if w.Vi != "" {
				b.WriteString(" → ")
				b.WriteString(w.Vi)
			}
			if w.Def != "" {
				b.WriteString(" (")
				b.WriteString(w.Def)
				b.WriteString(")")
			}
			b.WriteString("\n")
		}
	}
	// 3) phrases
	if len(la.Phrases) > 0 {
		b.WriteString("3) Cấu trúc/cụm từ:\n")
		for _, p := range la.Phrases {
			b.WriteString("- ")
			b.WriteString(p)
			b.WriteString("\n")
		}
	}
	// 4) examples
	if len(la.Examples) > 0 {
		b.WriteString("4) Ví dụ:\n")
		for _, ex := range la.Examples {
			b.WriteString("- EN: ")
			b.WriteString(ex.En)
			if ex.Vi != "" {
				b.WriteString(" | VI: ")
				b.WriteString(ex.Vi)
			}
			b.WriteString("\n")
		}
	}
	// 5) exercises
	if len(la.Exercises) > 0 {
		b.WriteString("5) Bài tập điền từ:\n")
		for _, ex := range la.Exercises {
			b.WriteString("- ")
			b.WriteString(ex.Question)
			if ex.Answer != "" {
				b.WriteString(" (Đáp án: ")
				b.WriteString(ex.Answer)
				b.WriteString(")")
			}
			b.WriteString("\n")
		}
	}
	return strings.TrimSpace(b.String())
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

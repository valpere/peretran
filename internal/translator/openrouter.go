package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/valpere/peretran/internal/postprocess"
)

var DefaultOpenRouterModels = []string{
	"google/gemini-2.0-flash-exp:free",
	"qwen/qwen2.5-72b-instruct:free",
	"mistralai/mistral-nemo:free",
	"meta-llama/llama-3.1-8b-instruct:free",
}

type OpenRouterService struct {
	apiKey  string
	baseURL string
	models  []string
	client  *http.Client
}

func NewOpenRouterService(apiKey string, baseURL string, models []string) *OpenRouterService {
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}
	if len(models) == 0 {
		models = DefaultOpenRouterModels
	}
	return &OpenRouterService{
		apiKey:  apiKey,
		baseURL: baseURL,
		models:  models,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (s *OpenRouterService) Name() string {
	return "openrouter"
}

func (s *OpenRouterService) getRandomModel() string {
	if len(s.models) == 0 {
		return "google/gemini-2.0-flash-exp:free"
	}
	return s.models[rand.Intn(len(s.models))]
}

func (s *OpenRouterService) SetModels(models []string) {
	if len(models) > 0 {
		s.models = models
	}
}

func (s *OpenRouterService) Translate(ctx context.Context, cfg ServiceConfig, req TranslateRequest) (*ServiceResult, error) {
	result := &ServiceResult{ServiceName: s.Name()}
	start := time.Now()
	defer func() { result.Latency = time.Since(start) }()

	apiKey := s.apiKey
	if apiKey == "" && cfg.APIKey != "" {
		apiKey = cfg.APIKey
	}

	if apiKey == "" {
		result.Error = "OpenRouter API key required"
		return result, fmt.Errorf("OpenRouter API key required")
	}

	model := s.getRandomModel()

	sourceLang := req.SourceLang
	if sourceLang == "" || sourceLang == "auto" {
		sourceLang = "the detected language"
	}

	systemPrompt := buildOpenRouterSystemPrompt(sourceLang, req.TargetLang, req.PreviousContext, req.GlossaryTerms, req.Instructions)

	openrouterReq := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": req.Text},
		},
		"max_tokens": 4096,
	}

	jsonData, err := json.Marshal(openrouterReq)
	if err != nil {
		result.Error = fmt.Sprintf("failed to marshal request: %v", err)
		return result, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat/completions", s.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	httpReq.Header.Set("HTTP-Referer", "https://peretran.local")
	httpReq.Header.Set("X-Title", "PereTran")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		result.Error = fmt.Sprintf("API returned status %d: %v", resp.StatusCode, errResp)
		return result, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var openrouterResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openrouterResp); err != nil {
		result.Error = fmt.Sprintf("failed to decode response: %v", err)
		return result, err
	}

	if len(openrouterResp.Choices) == 0 {
		result.Error = "empty response from API"
		return result, fmt.Errorf("empty response from API")
	}

	result.TranslatedText = postprocess.Clean(openrouterResp.Choices[0].Message.Content)
	result.Confidence = 0.7
	result.Metadata = map[string]string{
		"model":             model,
		"prompt_tokens":     fmt.Sprintf("%d", openrouterResp.Usage.PromptTokens),
		"completion_tokens": fmt.Sprintf("%d", openrouterResp.Usage.CompletionTokens),
	}

	return result, nil
}

func (s *OpenRouterService) IsAvailable(ctx context.Context) error {
	if s.apiKey == "" {
		return fmt.Errorf("OpenRouter API key not configured")
	}
	return nil
}

func (s *OpenRouterService) SupportedLanguages(ctx context.Context) ([]string, error) {
	return []string{"en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko", "ar", "uk"}, nil
}

func (s *OpenRouterService) GetModels() []string {
	return s.models
}

// buildOpenRouterSystemPrompt constructs the system prompt, optionally
// injecting glossary terms, a sliding-window context, and extra instructions.
func buildOpenRouterSystemPrompt(sourceLang, targetLang, previousContext string, glossary map[string]string, instructions string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("You are a professional translator. Translate the following text from %s to %s.\n", sourceLang, targetLang))
	sb.WriteString("Only respond with the translation, nothing else. No explanations, no quotes, just the translation.")

	if instructions != "" {
		sb.WriteString(" ")
		sb.WriteString(instructions)
	}

	if len(glossary) > 0 {
		sb.WriteString("\n\nTERMINOLOGY (use these exact translations):\n")
		for src, tgt := range glossary {
			sb.WriteString(fmt.Sprintf("  %s → %s\n", src, tgt))
		}
	}

	if previousContext != "" {
		sb.WriteString(fmt.Sprintf("\n\nCONTEXT (previous passage for continuity — do NOT retranslate this):\n...%s", previousContext))
	}

	return sb.String()
}

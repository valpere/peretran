package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/valpere/peretran/internal/postprocess"
)

var DefaultOllamaModels = []string{
	"llama3.2",
	"gemma2:2b",
	"qwen2.5:3b",
	"mistral:7b",
	"phi4:14b",
}

type OllamaTranslator struct {
	baseURL string
	models  []string
	client  *http.Client
}

func NewOllamaTranslator(baseURL string, models []string) *OllamaTranslator {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if len(models) == 0 {
		models = DefaultOllamaModels
	}
	return &OllamaTranslator{
		baseURL: baseURL,
		models:  models,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (s *OllamaTranslator) Name() string {
	return "ollama"
}

func (s *OllamaTranslator) getRandomModel() string {
	if len(s.models) == 0 {
		return "llama3.2"
	}
	return s.models[rand.Intn(len(s.models))]
}

func (s *OllamaTranslator) SetModels(models []string) {
	if len(models) > 0 {
		s.models = models
	}
}

func (s *OllamaTranslator) Translate(ctx context.Context, cfg ServiceConfig, req TranslateRequest) (*ServiceResult, error) {
	result := &ServiceResult{ServiceName: s.Name()}
	start := time.Now()
	defer func() { result.Latency = time.Since(start) }()

	model := cfg.Model
	if model == "" {
		model = s.getRandomModel()
	}

	sourceLang := req.SourceLang
	if sourceLang == "" || sourceLang == "auto" {
		sourceLang = "detect"
	}

	prompt := fmt.Sprintf(`Translate the following text from %s to %s.
Only respond with the translation, nothing else.

Text: "%s"

Translation:`, sourceLang, req.TargetLang, req.Text)

	ollamaReq := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		result.Error = fmt.Sprintf("failed to marshal request: %v", err)
		return result, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/generate", s.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Error = fmt.Sprintf("API returned status %d", resp.StatusCode)
		return result, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		result.Error = fmt.Sprintf("failed to decode response: %v", err)
		return result, err
	}

	result.TranslatedText = postprocess.Clean(ollamaResp.Response)
	result.Confidence = 0.7
	result.Metadata = map[string]string{"model": model}

	return result, nil
}

func (s *OllamaTranslator) IsAvailable(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/tags", s.baseURL), nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama not available: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}
	return nil
}

func (s *OllamaTranslator) SupportedLanguages(ctx context.Context) ([]string, error) {
	return []string{"en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko", "ar", "uk"}, nil
}

func (s *OllamaTranslator) GetModels() []string {
	return s.models
}

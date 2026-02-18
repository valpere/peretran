package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SystranService struct {
	apiKey string
	client *http.Client
}

func NewSystranService(apiKey string) *SystranService {
	return &SystranService{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SystranService) Name() string {
	return "systran"
}

func (s *SystranService) Translate(ctx context.Context, cfg ServiceConfig, req TranslateRequest) (*ServiceResult, error) {
	result := &ServiceResult{ServiceName: s.Name()}
	start := time.Now()
	defer func() { result.Latency = time.Since(start) }()

	if s.apiKey == "" && cfg.APIKey == "" {
		result.Error = "Systran API key required"
		return result, fmt.Errorf("Systran API key required")
	}

	apiKey := s.apiKey
	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	systranReq := map[string]interface{}{
		"text":   []string{req.Text},
		"source": req.SourceLang,
		"target": req.TargetLang,
		"format": "text",
	}

	jsonData, err := json.Marshal(systranReq)
	if err != nil {
		result.Error = fmt.Sprintf("failed to marshal request: %v", err)
		return result, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api-systran-systran-translation-v1.p.rapidapi.com/translation/text/translate", bytes.NewBuffer(jsonData))
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-RapidAPI-Key", apiKey)
	httpReq.Header.Set("X-RapidAPI-Host", "api-systran-systran-translation-v1.p.rapidapi.com")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		result.Error = fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body))
		return result, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var systranResp struct {
		Outputs []struct {
			Output string `json:"output"`
		} `json:"outputs"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&systranResp); err != nil {
		result.Error = fmt.Sprintf("failed to decode response: %v", err)
		return result, err
	}

	if len(systranResp.Outputs) == 0 || systranResp.Outputs[0].Output == "" {
		result.Error = "empty translation response"
		return result, fmt.Errorf("empty translation response")
	}

	result.TranslatedText = systranResp.Outputs[0].Output
	result.Confidence = 1.0

	return result, nil
}

func (s *SystranService) IsAvailable(ctx context.Context) error {
	if s.apiKey == "" {
		return fmt.Errorf("Systran API key not configured")
	}
	return nil
}

func (s *SystranService) SupportedLanguages(ctx context.Context) ([]string, error) {
	return []string{"en", "fr", "es", "de", "it", "pt", "ru", "zh", "ja", "ko", "ar"}, nil
}

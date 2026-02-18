package translator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type MyMemoryService struct {
	email  string
	client *http.Client
}

func NewMyMemoryService(email string) *MyMemoryService {
	return &MyMemoryService{
		email:  email,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *MyMemoryService) Name() string {
	return "mymemory"
}

func (s *MyMemoryService) Translate(ctx context.Context, cfg ServiceConfig, req TranslateRequest) (*ServiceResult, error) {
	result := &ServiceResult{ServiceName: s.Name()}
	start := time.Now()
	defer func() { result.Latency = time.Since(start) }()

	sourceLang := req.SourceLang
	if sourceLang == "" || sourceLang == "auto" {
		sourceLang = "en"
	}

	langPair := fmt.Sprintf("%s|%s", sourceLang, req.TargetLang)

	apiURL := fmt.Sprintf("https://api.mymemory.translated.net/get?q=%s&langpair=%s",
		url.QueryEscape(req.Text),
		url.QueryEscape(langPair))

	if s.email != "" {
		apiURL += fmt.Sprintf("&de=%s", url.QueryEscape(s.email))
	}

	httpReq, err := http.NewRequestWithContext(ctx, apiURL, "GET", nil)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		return result, err
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		return result, err
	}
	defer resp.Body.Close()

	var mymemResp struct {
		ResponseData struct {
			TranslatedText string  `json:"translatedText"`
			Match          float64 `json:"match"`
		} `json:"responseData"`
		ResponseStatus  int    `json:"responseStatus"`
		ResponseDetails string `json:"responseDetails"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&mymemResp); err != nil {
		result.Error = fmt.Sprintf("failed to decode response: %v", err)
		return result, err
	}

	if mymemResp.ResponseStatus != 200 {
		result.Error = fmt.Sprintf("API error: %s (%d)", mymemResp.ResponseDetails, mymemResp.ResponseStatus)
		return result, fmt.Errorf("API error: %s", mymemResp.ResponseDetails)
	}

	result.TranslatedText = mymemResp.ResponseData.TranslatedText
	result.Confidence = mymemResp.ResponseData.Match

	if result.Confidence < 0 {
		result.Confidence = 0
	}
	if result.Confidence > 1 {
		result.Confidence = 1
	}

	return result, nil
}

func (s *MyMemoryService) IsAvailable(ctx context.Context) error {
	return nil
}

func (s *MyMemoryService) SupportedLanguages(ctx context.Context) ([]string, error) {
	return []string{
		"en", "es", "fr", "de", "it", "pt", "ru", "ja", "ko", "zh",
		"ar", "nl", "pl", "tr", "sv", "da", "no", "fi", "el", "he",
		"th", "vi", "id", "ms", "cs", "hu", "ro", "uk", "bg", "ca",
	}, nil
}

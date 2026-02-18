package translator

import (
	"context"
	"fmt"
	"time"
)

type DoclingoService struct {
	apiKey string
}

func NewDoclingoService(apiKey string) *DoclingoService {
	return &DoclingoService{apiKey: apiKey}
}

func (s *DoclingoService) Name() string {
	return "doclingo"
}

func (s *DoclingoService) Translate(ctx context.Context, cfg ServiceConfig, req TranslateRequest) (*ServiceResult, error) {
	result := &ServiceResult{ServiceName: s.Name()}
	start := time.Now()
	defer func() { result.Latency = time.Since(start) }()

	result.Error = "DocLingo is document/PDF focused, not plain text - not implemented in this version"
	return result, fmt.Errorf("DocLingo is document/PDF focused, not plain text - not implemented in this version")
}

func (s *DoclingoService) IsAvailable(ctx context.Context) error {
	if s.apiKey == "" {
		return fmt.Errorf("DocLingo API key not configured")
	}
	return nil
}

func (s *DoclingoService) SupportedLanguages(ctx context.Context) ([]string, error) {
	return []string{"en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko"}, nil
}

package translator

import (
	"context"
	"fmt"
	"time"
)

type AmazonService struct{}

func NewAmazonService() *AmazonService {
	return &AmazonService{}
}

func (s *AmazonService) Name() string {
	return "amazon"
}

func (s *AmazonService) Translate(ctx context.Context, cfg ServiceConfig, req TranslateRequest) (*ServiceResult, error) {
	result := &ServiceResult{ServiceName: s.Name()}
	start := time.Now()
	defer func() { result.Latency = time.Since(start) }()

	result.Error = "Amazon Translate: Not Implemented Yet"
	return result, fmt.Errorf("Amazon Translate: Not Implemented Yet")
}

func (s *AmazonService) IsAvailable(ctx context.Context) error {
	return fmt.Errorf("Amazon Translate: Not Implemented Yet")
}

func (s *AmazonService) SupportedLanguages(ctx context.Context) ([]string, error) {
	return []string{
		"en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko",
		"ar", "hi", "id", "ms", "th", "tr", "vi", "uk", "cs", "pl",
		"nl", "sv", "da", "no", "fi", "el", "he", "hu", "ro",
	}, nil
}

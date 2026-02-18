package translator

import (
	"context"
	"fmt"
	"time"
)

type IBMService struct{}

func NewIBMService() *IBMService {
	return &IBMService{}
}

func (s *IBMService) Name() string {
	return "ibm"
}

func (s *IBMService) Translate(ctx context.Context, cfg ServiceConfig, req TranslateRequest) (*ServiceResult, error) {
	result := &ServiceResult{ServiceName: s.Name()}
	start := time.Now()
	defer func() { result.Latency = time.Since(start) }()

	result.Error = "IBM Watson Translate: Not Implemented Yet"
	return result, fmt.Errorf("IBM Watson Translate: Not Implemented Yet")
}

func (s *IBMService) IsAvailable(ctx context.Context) error {
	return fmt.Errorf("IBM Watson Translate: Not Implemented Yet")
}

func (s *IBMService) SupportedLanguages(ctx context.Context) ([]string, error) {
	return []string{
		"en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko",
		"ar", "nl", "pl", "tr", "sv", "da", "no", "fi", "el", "he",
		"th", "vi", "id", "ms", "cs", "hu", "ro", "uk", "bg", "ca",
	}, nil
}

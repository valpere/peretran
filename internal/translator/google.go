package translator

import (
	"context"
	"fmt"
	"os"
	"time"

	translate "cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

type GoogleService struct{}

func NewGoogleService() *GoogleService {
	return &GoogleService{}
}

func (s *GoogleService) Name() string {
	return "google"
}

func (s *GoogleService) Translate(ctx context.Context, cfg ServiceConfig, req TranslateRequest) (*ServiceResult, error) {
	result := &ServiceResult{ServiceName: s.Name()}
	start := time.Now()
	defer func() { result.Latency = time.Since(start) }()

	if cfg.Credentials != "" {
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cfg.Credentials)
	}

	targetLangTag, err := language.Parse(req.TargetLang)
	if err != nil {
		result.Error = fmt.Sprintf("invalid target language: %v", err)
		return result, fmt.Errorf("invalid target language: %v", err)
	}

	opts := []option.ClientOption{}
	if cfg.Credentials != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.Credentials))
	}

	client, err := translate.NewClient(ctx, opts...)
	if err != nil {
		result.Error = fmt.Sprintf("failed to create client: %v", err)
		return result, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	var translations []translate.Translation
	if req.SourceLang == "" || req.SourceLang == "auto" {
		translations, err = client.Translate(ctx, []string{req.Text}, targetLangTag, nil)
	} else {
		sourceLangTag, _ := language.Parse(req.SourceLang)
		translations, err = client.Translate(ctx, []string{req.Text}, targetLangTag, &translate.Options{
			Source: sourceLangTag,
		})
	}

	if err != nil {
		result.Error = fmt.Sprintf("translation failed: %v", err)
		return result, fmt.Errorf("translation failed: %v", err)
	}

	if len(translations) == 0 {
		result.Error = "no translation returned"
		return result, fmt.Errorf("no translation returned")
	}

	result.TranslatedText = translations[0].Text
	result.Confidence = 1.0

	return result, nil
}

func (s *GoogleService) IsAvailable(ctx context.Context) error {
	return nil
}

func (s *GoogleService) SupportedLanguages(ctx context.Context) ([]string, error) {
	return nil, nil
}

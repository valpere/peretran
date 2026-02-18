package translator

import (
	"context"
	"time"
)

type ServiceConfig struct {
	Credentials string        `mapstructure:"credentials" json:"credentials"`
	APIKey      string        `mapstructure:"api_key" json:"api_key"`
	Model       string        `mapstructure:"model" json:"model"`
	BaseURL     string        `mapstructure:"base_url" json:"base_url"`
	Timeout     time.Duration `mapstructure:"timeout" json:"timeout"`
	ProjectID   string        `mapstructure:"project_id" json:"project_id"`
}

type TranslateRequest struct {
	Text       string `json:"text"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
}

type ServiceResult struct {
	ServiceName    string            `json:"service_name"`
	TranslatedText string            `json:"translated_text"`
	Confidence     float64           `json:"confidence"`
	Metadata       map[string]string `json:"metadata"`
	Latency        time.Duration     `json:"latency"`
	Error          string            `json:"error,omitempty"`
}

type TranslationService interface {
	Name() string
	Translate(ctx context.Context, cfg ServiceConfig, req TranslateRequest) (*ServiceResult, error)
	IsAvailable(ctx context.Context) error
	SupportedLanguages(ctx context.Context) ([]string, error)
}

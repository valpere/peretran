package translator

import (
	"context"
	"testing"
)

func TestAmazonService_Translate(t *testing.T) {
	svc := NewAmazonService()

	result, err := svc.Translate(context.Background(), ServiceConfig{}, TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	})

	if err == nil {
		t.Error("expected error for not implemented service")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Error == "" {
		t.Error("expected error message in result")
	}
	if result.ServiceName != "amazon" {
		t.Errorf("expected service name 'amazon', got %q", result.ServiceName)
	}
	if result.Latency <= 0 {
		t.Error("expected positive latency")
	}
}

func TestAmazonService_IsAvailable(t *testing.T) {
	svc := NewAmazonService()

	err := svc.IsAvailable(context.Background())
	if err == nil {
		t.Error("expected error for not implemented service")
	}
}

func TestAmazonService_SupportedLanguages(t *testing.T) {
	svc := NewAmazonService()

	langs, err := svc.SupportedLanguages(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(langs) == 0 {
		t.Error("expected non-empty language list")
	}
}

func TestAmazonService_Name(t *testing.T) {
	svc := NewAmazonService()

	if svc.Name() != "amazon" {
		t.Errorf("expected 'amazon', got %q", svc.Name())
	}
}

func TestIBMService_Translate(t *testing.T) {
	svc := NewIBMService()

	result, err := svc.Translate(context.Background(), ServiceConfig{}, TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	})

	if err == nil {
		t.Error("expected error for not implemented service")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Error == "" {
		t.Error("expected error message in result")
	}
	if result.ServiceName != "ibm" {
		t.Errorf("expected service name 'ibm', got %q", result.ServiceName)
	}
}

func TestIBMService_IsAvailable(t *testing.T) {
	svc := NewIBMService()

	err := svc.IsAvailable(context.Background())
	if err == nil {
		t.Error("expected error for not implemented service")
	}
}

func TestIBMService_SupportedLanguages(t *testing.T) {
	svc := NewIBMService()

	langs, err := svc.SupportedLanguages(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(langs) == 0 {
		t.Error("expected non-empty language list")
	}
}

func TestIBMService_Name(t *testing.T) {
	svc := NewIBMService()

	if svc.Name() != "ibm" {
		t.Errorf("expected 'ibm', got %q", svc.Name())
	}
}

func TestDoclingoService_Translate(t *testing.T) {
	svc := NewDoclingoService("test-key")

	result, err := svc.Translate(context.Background(), ServiceConfig{}, TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	})

	if err == nil {
		t.Error("expected error for not implemented service")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Error == "" {
		t.Error("expected error message in result")
	}
	if result.ServiceName != "doclingo" {
		t.Errorf("expected service name 'doclingo', got %q", result.ServiceName)
	}
}

func TestDoclingoService_IsAvailable_NoAPIKey(t *testing.T) {
	svc := NewDoclingoService("")

	err := svc.IsAvailable(context.Background())
	if err == nil {
		t.Error("expected error when no API key")
	}
}

func TestDoclingoService_IsAvailable_WithAPIKey(t *testing.T) {
	svc := NewDoclingoService("test-key")

	err := svc.IsAvailable(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDoclingoService_SupportedLanguages(t *testing.T) {
	svc := NewDoclingoService("test-key")

	langs, err := svc.SupportedLanguages(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(langs) == 0 {
		t.Error("expected non-empty language list")
	}
}

func TestDoclingoService_Name(t *testing.T) {
	svc := NewDoclingoService("test-key")

	if svc.Name() != "doclingo" {
		t.Errorf("expected 'doclingo', got %q", svc.Name())
	}
}

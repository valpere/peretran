package translator

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMyMemoryService_IsAvailable(t *testing.T) {
	svc := NewMyMemoryService("test@example.com")

	err := svc.IsAvailable(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMyMemoryService_SupportedLanguages(t *testing.T) {
	svc := NewMyMemoryService("")

	langs, err := svc.SupportedLanguages(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(langs) == 0 {
		t.Error("expected non-empty language list")
	}
}

func TestMyMemoryService_Name(t *testing.T) {
	svc := NewMyMemoryService("")

	if svc.Name() != "mymemory" {
		t.Errorf("expected 'mymemory', got %q", svc.Name())
	}
}

func TestSystranService_Translate_NoAPIKey(t *testing.T) {
	svc := NewSystranService("")

	result, err := svc.Translate(context.Background(), ServiceConfig{}, TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "fr",
	})

	if err == nil {
		t.Error("expected error when no API key")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Error == "" {
		t.Error("expected error message in result")
	}
}

func TestSystranService_Translate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
	}))
	defer server.Close()

	svc := &SystranService{
		apiKey: "test-key",
		client: server.Client(),
	}

	result, err := svc.Translate(context.Background(), ServiceConfig{}, TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "fr",
	})

	if err == nil {
		t.Error("expected error for non-OK status")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestSystranService_IsAvailable_NoAPIKey(t *testing.T) {
	svc := NewSystranService("")

	err := svc.IsAvailable(context.Background())
	if err == nil {
		t.Error("expected error when no API key")
	}
}

func TestSystranService_IsAvailable_WithAPIKey(t *testing.T) {
	svc := NewSystranService("test-key")

	err := svc.IsAvailable(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSystranService_SupportedLanguages(t *testing.T) {
	svc := NewSystranService("test-key")

	langs, err := svc.SupportedLanguages(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(langs) == 0 {
		t.Error("expected non-empty language list")
	}
}

func TestSystranService_Name(t *testing.T) {
	svc := NewSystranService("test-key")

	if svc.Name() != "systran" {
		t.Errorf("expected 'systran', got %q", svc.Name())
	}
}

func TestOllamaTranslator_Translate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"response": "Привіт",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := &OllamaTranslator{
		baseURL: server.URL,
		models:  []string{"llama3.2"},
		client:  server.Client(),
	}

	result, err := svc.Translate(context.Background(), ServiceConfig{}, TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.TranslatedText != "Привіт" {
		t.Errorf("expected 'Привіт', got %q", result.TranslatedText)
	}
	if result.Metadata["model"] != "llama3.2" {
		t.Errorf("expected model in metadata, got %v", result.Metadata)
	}
}

func TestOllamaTranslator_Translate_AutoSourceLang(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		prompt := req["prompt"].(string)
		if len(prompt) == 0 {
			t.Error("expected prompt in request")
		}
		resp := map[string]interface{}{"response": "Привіт"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := &OllamaTranslator{
		baseURL: server.URL,
		models:  []string{"llama3.2"},
		client:  server.Client(),
	}

	_, err := svc.Translate(context.Background(), ServiceConfig{}, TranslateRequest{
		Text:       "Hello",
		SourceLang: "auto",
		TargetLang: "uk",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOllamaTranslator_Translate_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := &OllamaTranslator{
		baseURL: server.URL,
		models:  []string{"llama3.2"},
		client:  server.Client(),
	}

	result, err := svc.Translate(context.Background(), ServiceConfig{}, TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	})

	if err == nil {
		t.Error("expected error for non-OK status")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestOllamaTranslator_IsAvailable_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := &OllamaTranslator{
		baseURL: server.URL,
		client:  server.Client(),
	}

	err := svc.IsAvailable(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOllamaTranslator_IsAvailable_NotRunning(t *testing.T) {
	svc := &OllamaTranslator{
		baseURL: "http://localhost:19999",
		client:  &http.Client{Timeout: 100 * time.Millisecond},
	}

	err := svc.IsAvailable(context.Background())
	if err == nil {
		t.Error("expected error when Ollama not available")
	}
}

func TestOllamaTranslator_SupportedLanguages(t *testing.T) {
	svc := NewOllamaTranslator("", nil)

	langs, err := svc.SupportedLanguages(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(langs) == 0 {
		t.Error("expected non-empty language list")
	}
}

func TestOllamaTranslator_Name(t *testing.T) {
	svc := NewOllamaTranslator("", nil)

	if svc.Name() != "ollama" {
		t.Errorf("expected 'ollama', got %q", svc.Name())
	}
}

func TestOllamaTranslator_GetModels(t *testing.T) {
	models := []string{"llama3.2", "gemma2:2b"}
	svc := NewOllamaTranslator("", models)

	got := svc.GetModels()
	if len(got) != len(models) {
		t.Errorf("expected %d models, got %d", len(models), len(got))
	}
}

func TestOllamaTranslator_SetModels(t *testing.T) {
	svc := NewOllamaTranslator("", []string{"llama3.2"})

	svc.SetModels([]string{"gemma2:2b", "qwen2.5:3b"})

	got := svc.GetModels()
	if len(got) != 2 {
		t.Errorf("expected 2 models, got %d", len(got))
	}
}

func TestOllamaTranslator_SetModels_Empty(t *testing.T) {
	svc := NewOllamaTranslator("", []string{"llama3.2"})

	svc.SetModels([]string{})

	got := svc.GetModels()
	if len(got) != 1 {
		t.Errorf("expected 1 model (unchanged), got %d", len(got))
	}
}

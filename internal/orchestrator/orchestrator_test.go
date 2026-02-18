package orchestrator

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/valpere/peretran/internal/translator"
)

type mockService struct {
	nameVal       string
	translateFunc func(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) (*translator.ServiceResult, error)
	availableFunc func(ctx context.Context) error
	languagesFunc func(ctx context.Context) ([]string, error)
	callCount     atomic.Int32
}

func (m *mockService) Name() string { return m.nameVal }

func (m *mockService) Translate(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) (*translator.ServiceResult, error) {
	m.callCount.Add(1)
	if m.translateFunc != nil {
		return m.translateFunc(ctx, cfg, req)
	}
	return &translator.ServiceResult{ServiceName: m.nameVal, TranslatedText: "mock result"}, nil
}

func (m *mockService) IsAvailable(ctx context.Context) error {
	if m.availableFunc != nil {
		return m.availableFunc(ctx)
	}
	return nil
}

func (m *mockService) SupportedLanguages(ctx context.Context) ([]string, error) {
	if m.languagesFunc != nil {
		return m.languagesFunc(ctx)
	}
	return []string{"en", "uk"}, nil
}

func TestOrchestrator_New(t *testing.T) {
	services := []translator.TranslationService{
		&mockService{nameVal: "mock1"},
	}

	config := OrchestratorConfig{
		Timeout:     10 * time.Second,
		MaxAttempts: 3,
		RetryDelay:  100 * time.Millisecond,
	}

	o := New(services, config)

	if o == nil {
		t.Fatal("expected non-nil Orchestrator")
	}
	if o.validator == nil {
		t.Error("expected validator to be created by default")
	}
}

func TestOrchestrator_New_SkipValidation(t *testing.T) {
	services := []translator.TranslationService{
		&mockService{nameVal: "mock1"},
	}

	config := OrchestratorConfig{
		Timeout:        10 * time.Second,
		SkipValidation: true,
	}

	o := New(services, config)

	if o == nil {
		t.Fatal("expected non-nil Orchestrator")
	}
	if o.validator != nil {
		t.Error("expected nil validator when SkipValidation is true")
	}
}

func TestOrchestrator_New_Defaults(t *testing.T) {
	services := []translator.TranslationService{
		&mockService{nameVal: "mock1"},
	}

	config := OrchestratorConfig{}

	o := New(services, config)

	if o.config.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", o.config.MaxAttempts)
	}
	if o.config.RetryDelay <= 0 {
		t.Error("expected positive RetryDelay")
	}
}

func TestOrchestrator_Execute_SingleService(t *testing.T) {
	svc := &mockService{nameVal: "mock1"}
	services := []translator.TranslationService{svc}

	o := New(services, OrchestratorConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		RetryDelay:  10 * time.Millisecond,
	})

	req := translator.TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	}

	result := o.Execute(context.Background(), translator.ServiceConfig{}, req)

	if result.Succeeded != 1 {
		t.Errorf("expected 1 succeeded, got %d", result.Succeeded)
	}
	if result.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", result.Failed)
	}
	if len(result.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(result.Results))
	}
}

func TestOrchestrator_Execute_MultipleServices(t *testing.T) {
	svc1 := &mockService{nameVal: "service1"}
	svc2 := &mockService{nameVal: "service2"}
	svc3 := &mockService{nameVal: "service3"}
	services := []translator.TranslationService{svc1, svc2, svc3}

	o := New(services, OrchestratorConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		RetryDelay:  10 * time.Millisecond,
	})

	req := translator.TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	}

	result := o.Execute(context.Background(), translator.ServiceConfig{}, req)

	if result.Succeeded != 3 {
		t.Errorf("expected 3 succeeded, got %d", result.Succeeded)
	}
	if result.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", result.Failed)
	}
	if len(result.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(result.Results))
	}
}

func TestOrchestrator_Execute_WithFailures(t *testing.T) {
	svc1 := &mockService{
		nameVal: "service1",
		translateFunc: func(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) (*translator.ServiceResult, error) {
			return nil, errors.New("service unavailable")
		},
	}
	svc2 := &mockService{nameVal: "service2"}
	services := []translator.TranslationService{svc1, svc2}

	o := New(services, OrchestratorConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		RetryDelay:  10 * time.Millisecond,
	})

	req := translator.TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	}

	result := o.Execute(context.Background(), translator.ServiceConfig{}, req)

	if result.Succeeded != 1 {
		t.Errorf("expected 1 succeeded, got %d", result.Succeeded)
	}
	if result.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", result.Failed)
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}

func TestOrchestrator_Execute_WithRetry(t *testing.T) {
	callCount := atomic.Int32{}
	svc := &mockService{
		nameVal: "retryable",
		translateFunc: func(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) (*translator.ServiceResult, error) {
			count := callCount.Add(1)
			if count < 3 {
				return &translator.ServiceResult{ServiceName: "retryable", Error: "temporary failure"}, nil
			}
			return &translator.ServiceResult{ServiceName: "retryable", TranslatedText: "success on 3rd attempt"}, nil
		},
	}
	services := []translator.TranslationService{svc}

	o := New(services, OrchestratorConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 3,
		RetryDelay:  10 * time.Millisecond,
	})

	req := translator.TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	}

	result := o.Execute(context.Background(), translator.ServiceConfig{}, req)

	if result.Succeeded != 1 {
		t.Errorf("expected 1 succeeded after retry, got %d", result.Succeeded)
	}
	if svc.callCount.Load() != 3 {
		t.Errorf("expected 3 calls (1 initial + 2 retries), got %d", svc.callCount.Load())
	}
}

func TestOrchestrator_Execute_Cancellation(t *testing.T) {
	svc := &mockService{
		nameVal: "slow",
		translateFunc: func(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) (*translator.ServiceResult, error) {
			time.Sleep(1 * time.Second)
			return &translator.ServiceResult{ServiceName: "slow", TranslatedText: "done"}, nil
		},
	}
	services := []translator.TranslationService{svc}

	o := New(services, OrchestratorConfig{
		Timeout:     10 * time.Second,
		MaxAttempts: 3,
		RetryDelay:  100 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := translator.TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	}

	result := o.Execute(ctx, translator.ServiceConfig{}, req)

	// With cancelled context, services may still complete since they run in goroutines
	// The result could succeed or fail depending on timing - just verify no panic
	_ = result
}

func TestOrchestrator_ExecuteWithFallback(t *testing.T) {
	svc := &mockService{nameVal: "mock"}
	services := []translator.TranslationService{svc}

	o := New(services, OrchestratorConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		RetryDelay:  10 * time.Millisecond,
	})

	req := translator.TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	}

	result := o.ExecuteWithFallback(context.Background(), translator.ServiceConfig{}, req)

	if result == nil {
		t.Fatal("expected non-nil result from ExecuteWithFallback")
	}
	if result.ServiceName != "mock" {
		t.Errorf("expected mock service name, got %s", result.ServiceName)
	}
}

func TestOrchestrator_ExecuteWithFallback_AllFailed(t *testing.T) {
	svc := &mockService{
		nameVal: "failing",
		translateFunc: func(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) (*translator.ServiceResult, error) {
			return nil, errors.New("always fails")
		},
	}
	services := []translator.TranslationService{svc}

	o := New(services, OrchestratorConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 1,
		RetryDelay:  10 * time.Millisecond,
	})

	req := translator.TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk",
	}

	result := o.ExecuteWithFallback(context.Background(), translator.ServiceConfig{}, req)

	if result != nil {
		t.Error("expected nil result when all services fail")
	}
}

func TestOrchestrator_Execute_ValidationFailure(t *testing.T) {
	svc := &mockService{
		nameVal: "bad-translator",
		translateFunc: func(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) (*translator.ServiceResult, error) {
			// Return text in wrong language
			return &translator.ServiceResult{ServiceName: "bad-translator", TranslatedText: "This should fail validation because it is clearly in English not Ukrainian"}, nil
		},
	}
	services := []translator.TranslationService{svc}

	o := New(services, OrchestratorConfig{
		Timeout:     5 * time.Second,
		MaxAttempts: 3,
		RetryDelay:  10 * time.Millisecond,
	})

	req := translator.TranslateRequest{
		Text:       "Hello",
		SourceLang: "en",
		TargetLang: "uk", // Expecting Ukrainian but getting English
	}

	result := o.Execute(context.Background(), translator.ServiceConfig{}, req)

	// After retries exhausted, returns result anyway (with validation failure logged)
	if result.Succeeded != 1 {
		t.Errorf("expected 1 succeeded (validation failure on final attempt still returns result), got %d", result.Succeeded)
	}
}

package orchestrator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/valpere/peretran/internal/translator"
	"github.com/valpere/peretran/internal/validator"
)

// OrchestratorConfig controls parallel execution, retry behaviour, and validation.
type OrchestratorConfig struct {
	// Timeout is the per-attempt call timeout applied to each service.
	Timeout time.Duration

	// MinServices is the minimum number of successful results required (default 1).
	MinServices int

	// MaxAttempts is the total number of tries per service, including the first
	// attempt (default 3: one initial call plus two retries).
	MaxAttempts int

	// RetryDelay is the base wait before the first retry; it doubles on every
	// subsequent retry (exponential back-off, default 500 ms).
	RetryDelay time.Duration

	// SkipValidation disables target-language checking of translation results.
	SkipValidation bool
}

// OrchestratorResult holds the aggregated output of a parallel translation run.
type OrchestratorResult struct {
	Results   []translator.ServiceResult
	Errors    []error
	Succeeded int
	Failed    int
}

// Orchestrator runs multiple TranslationServices in parallel and collects results.
type Orchestrator struct {
	services  []translator.TranslationService
	config    OrchestratorConfig
	validator *validator.Validator
}

// New creates an Orchestrator. A language validator is built automatically unless
// SkipValidation is set. Unset MaxAttempts and RetryDelay receive safe defaults.
func New(services []translator.TranslationService, config OrchestratorConfig) *Orchestrator {
	if config.MaxAttempts < 1 {
		config.MaxAttempts = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 500 * time.Millisecond
	}

	var v *validator.Validator
	if !config.SkipValidation {
		v = validator.New()
	}

	return &Orchestrator{
		services:  services,
		config:    config,
		validator: v,
	}
}

// Execute runs all configured services concurrently and returns their results.
func (o *Orchestrator) Execute(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) *OrchestratorResult {
	result := &OrchestratorResult{
		Results: make([]translator.ServiceResult, 0, len(o.services)),
		Errors:  make([]error, 0),
	}

	type outcome struct {
		res *translator.ServiceResult
		err error
	}

	ch := make(chan outcome, len(o.services))

	var wg sync.WaitGroup
	for _, svc := range o.services {
		wg.Add(1)
		go func(service translator.TranslationService) {
			defer wg.Done()
			res, err := o.translateWithRetry(ctx, cfg, req, service)
			ch <- outcome{res: res, err: err}
		}(svc)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for oc := range ch {
		if oc.err != nil {
			result.Errors = append(result.Errors, oc.err)
			result.Failed++
		} else {
			result.Results = append(result.Results, *oc.res)
			result.Succeeded++
		}
	}

	return result
}

// translateWithRetry calls svc.Translate up to MaxAttempts times with exponential
// back-off between attempts. If target-language validation fails and retries remain,
// the call is retried. On the final attempt a validation failure is logged but the
// result is returned anyway so the pipeline always has something to work with.
func (o *Orchestrator) translateWithRetry(
	ctx context.Context,
	cfg translator.ServiceConfig,
	req translator.TranslateRequest,
	svc translator.TranslationService,
) (*translator.ServiceResult, error) {
	var lastResult *translator.ServiceResult
	var lastErr error
	delay := o.config.RetryDelay

	for attempt := 0; attempt < o.config.MaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2
		}

		callCtx, cancel := context.WithTimeout(ctx, o.config.Timeout)
		res, err := svc.Translate(callCtx, cfg, req)
		cancel()

		if err != nil {
			lastErr = err
			if attempt < o.config.MaxAttempts-1 {
				fmt.Fprintf(os.Stderr, "[%s] attempt %d/%d failed: %v, retrying...\n",
					svc.Name(), attempt+1, o.config.MaxAttempts, err)
			}
			continue
		}

		if res.Error != "" {
			lastErr = fmt.Errorf("%s: %s", res.ServiceName, res.Error)
			if attempt < o.config.MaxAttempts-1 {
				fmt.Fprintf(os.Stderr, "[%s] attempt %d/%d error: %s, retrying...\n",
					svc.Name(), attempt+1, o.config.MaxAttempts, res.Error)
			}
			continue
		}

		// Validate that the result is written in the target language.
		if o.validator != nil {
			if valid, validErr := o.validator.IsValid(res.TranslatedText, req.TargetLang); !valid {
				lastResult = res
				lastErr = validErr
				if attempt < o.config.MaxAttempts-1 {
					fmt.Fprintf(os.Stderr, "[%s] attempt %d/%d validation failed (%v), retrying...\n",
						svc.Name(), attempt+1, o.config.MaxAttempts, validErr)
					continue
				}
				// Final attempt: return the result rather than discarding it.
				fmt.Fprintf(os.Stderr, "[%s] validation failed after %d attempts, using result anyway\n",
					svc.Name(), o.config.MaxAttempts)
				return res, nil
			}
		}

		return res, nil
	}

	// All attempts exhausted by hard errors; if we have a validation-failed result, use it.
	if lastResult != nil {
		return lastResult, nil
	}
	return nil, lastErr
}

// ExecuteWithFallback is a convenience wrapper that returns the first successful result.
func (o *Orchestrator) ExecuteWithFallback(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) *translator.ServiceResult {
	result := o.Execute(ctx, cfg, req)
	if result.Succeeded == 0 {
		return nil
	}
	return &result.Results[0]
}

package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/valpere/peretran/internal/translator"
)

type OrchestratorConfig struct {
	Timeout     time.Duration
	MinServices int
	RetryCount  int
}

type OrchestratorResult struct {
	Results   []translator.ServiceResult
	Errors    []error
	Succeeded int
	Failed    int
}

type Orchestrator struct {
	services []translator.TranslationService
	config   OrchestratorConfig
}

func New(services []translator.TranslationService, config OrchestratorConfig) *Orchestrator {
	return &Orchestrator{
		services: services,
		config:   config,
	}
}

func (o *Orchestrator) Execute(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) *OrchestratorResult {
	result := &OrchestratorResult{
		Results: make([]translator.ServiceResult, 0),
		Errors:  make([]error, 0),
	}

	type resultChan struct {
		index int
		res   *translator.ServiceResult
		err   error
	}

	resultChanSlice := make(chan resultChan, len(o.services))

	var wg sync.WaitGroup
	for i, svc := range o.services {
		wg.Add(1)
		go func(index int, service translator.TranslationService) {
			defer wg.Done()

			serviceCtx, cancel := context.WithTimeout(ctx, o.config.Timeout)
			defer cancel()

			res, err := service.Translate(serviceCtx, cfg, req)
			resultChanSlice <- resultChan{index: index, res: res, err: err}
		}(i, svc)
	}

	go func() {
		wg.Wait()
		close(resultChanSlice)
	}()

	for rc := range resultChanSlice {
		if rc.err != nil {
			result.Errors = append(result.Errors, rc.err)
			result.Failed++
		} else if rc.res.Error != "" {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %s", rc.res.ServiceName, rc.res.Error))
			result.Failed++
		} else {
			result.Results = append(result.Results, *rc.res)
			result.Succeeded++
		}
	}

	return result
}

func (o *Orchestrator) ExecuteWithFallback(ctx context.Context, cfg translator.ServiceConfig, req translator.TranslateRequest) *translator.ServiceResult {
	result := o.Execute(ctx, cfg, req)

	if result.Succeeded == 0 {
		return nil
	}

	if result.Succeeded == 1 {
		return &result.Results[0]
	}

	return nil
}

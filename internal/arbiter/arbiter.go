package arbiter

import (
	"context"

	"github.com/valpere/peretran/internal/translator"
)

type EvaluationResult struct {
	SelectedService string
	CompositeText   string
	IsComposite     bool
	Reasoning       string
}

type Arbiter interface {
	Evaluate(ctx context.Context, source string, sourceLang, targetLang string, results []translator.ServiceResult) (*EvaluationResult, error)
}

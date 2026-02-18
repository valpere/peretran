// Package refiner implements Stage 2 of the two-pass translation pipeline.
// It takes a draft translation and refines it for literary quality using an LLM.
package refiner

import "context"

// Refiner reviews and improves a draft translation for literary quality.
type Refiner interface {
	Refine(ctx context.Context, sourceLang, targetLang, sourceText, draftText string) (string, error)
}

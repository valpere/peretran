// Package validator checks that a translation result is in the expected target language.
package validator

import (
	"fmt"
	"strings"

	"github.com/valpere/peretran/internal/detector"
)

// minValidationLength is the minimum rune count required to attempt language detection.
// Shorter texts produce unreliable results and are accepted without validation.
const minValidationLength = 20

// Validator checks that a translation result is written in the expected target language.
// The underlying language detector is expensive to build; reuse the instance.
type Validator struct {
	det *detector.Detector
}

// New creates a Validator backed by the lingua-go language detector.
func New() *Validator {
	return &Validator{det: detector.New()}
}

// IsValid returns true when translatedText appears to be written in targetLang.
//
// Short texts (fewer than minValidationLength runes) and texts whose language
// cannot be determined pass without error. When the detected language differs
// from targetLang the returned error names both codes.
func (v *Validator) IsValid(translatedText, targetLang string) (bool, error) {
	if targetLang == "" {
		return true, nil
	}

	text := strings.TrimSpace(translatedText)
	if text == "" {
		return false, fmt.Errorf("translation is empty")
	}

	// Detector is unreliable for very short texts; skip validation.
	if len([]rune(text)) < minValidationLength {
		return true, nil
	}

	detected, ok := v.det.DetectISO(text)
	if !ok {
		// Ambiguous language — cannot validate, pass through.
		return true, nil
	}

	if !strings.EqualFold(detected, targetLang) {
		return false, fmt.Errorf("expected %s but detected %s", targetLang, detected)
	}

	return true, nil
}

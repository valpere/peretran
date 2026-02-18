package validator

import (
	"testing"
)

func TestIsValid_EmptyTargetLang(t *testing.T) {
	v := New()

	valid, err := v.IsValid("Some translated text", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected valid=true for empty targetLang")
	}
}

func TestIsValid_EmptyTranslation(t *testing.T) {
	v := New()

	valid, err := v.IsValid("", "en")
	if err == nil {
		t.Error("expected error for empty translation")
	}
	if valid {
		t.Error("expected valid=false for empty translation")
	}
}

func TestIsValid_WhitespaceOnlyTranslation(t *testing.T) {
	v := New()

	valid, err := v.IsValid("   ", "en")
	if err == nil {
		t.Error("expected error for whitespace-only translation")
	}
	if valid {
		t.Error("expected valid=false for whitespace-only translation")
	}
}

func TestIsValid_ShortText(t *testing.T) {
	v := New()

	shortText := "Hi" // Less than minValidationLength (20 chars)
	valid, err := v.IsValid(shortText, "en")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected valid=true for short text (below threshold)")
	}
}

func TestIsValid_EnglishToEnglish(t *testing.T) {
	v := New()

	text := "This is a longer piece of text that should be detected as English."
	valid, err := v.IsValid(text, "en")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected valid=true when detecting English as English")
	}
}

func TestIsValid_MismatchedLanguage(t *testing.T) {
	v := New()

	englishText := "This is a longer piece of text that should be detected as English."
	valid, err := v.IsValid(englishText, "uk")
	if err == nil {
		t.Error("expected error for mismatched language")
	}
	if valid {
		t.Error("expected valid=false when detecting English but expecting Ukrainian")
	}
}

func TestIsValid_UkrainianText(t *testing.T) {
	v := New()

	ukrainianText := "Це є тестовий текст українською мовою для перевірки роботи валідатора."
	valid, err := v.IsValid(ukrainianText, "uk")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected valid=true when detecting Ukrainian as Ukrainian")
	}
}

func TestIsValid_CaseInsensitiveTargetLang(t *testing.T) {
	v := New()

	text := "This is a longer piece of text that should be detected as English."
	valid, err := v.IsValid(text, "EN") // uppercase
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !valid {
		t.Error("expected valid=true for case-insensitive targetLang")
	}
}

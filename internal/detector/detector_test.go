package detector

import (
	"testing"
)

func TestDetector_Detect(t *testing.T) {
	d := New()

	tests := []struct {
		name     string
		text     string
		wantLang string
		wantOK   bool
	}{
		{
			name:     "empty text",
			text:     "",
			wantLang: "",
			wantOK:   false,
		},
		{
			name:     "english text",
			text:     "Hello, this is a test in English.",
			wantLang: "English",
			wantOK:   true,
		},
		{
			name:     "ukrainian text",
			text:     "Привіт, це тест українською мовою.",
			wantLang: "Ukrainian",
			wantOK:   true,
		},
		{
			name:     "german text",
			text:     "Hallo, das ist ein Test auf Deutsch.",
			wantLang: "German",
			wantOK:   true,
		},
		{
			name:     "french text",
			text:     "Bonjour, ceci est un test en français.",
			wantLang: "French",
			wantOK:   true,
		},
		{
			name:     "spanish text",
			text:     "Hola, esto es una prueba en español.",
			wantLang: "Spanish",
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang, ok := d.Detect(tt.text)
			if ok != tt.wantOK {
				t.Errorf("Detect(%q) ok = %v, want %v", tt.text, ok, tt.wantOK)
				return
			}
			if tt.wantOK && lang.String() != tt.wantLang {
				t.Errorf("Detect(%q) = %v, want %v", tt.text, lang, tt.wantLang)
			}
		})
	}
}

func TestDetector_DetectISO(t *testing.T) {
	d := New()

	tests := []struct {
		name     string
		text     string
		wantCode string
		wantOK   bool
	}{
		{
			name:     "empty text",
			text:     "",
			wantCode: "",
			wantOK:   false,
		},
		{
			name:     "english text",
			text:     "Hello, this is a test in English.",
			wantCode: "EN",
			wantOK:   true,
		},
		{
			name:     "ukrainian text",
			text:     "Привіт, це тест українською мовою.",
			wantCode: "UK",
			wantOK:   true,
		},
		{
			name:     "german text",
			text:     "Hallo, das ist ein Test auf Deutsch.",
			wantCode: "DE",
			wantOK:   true,
		},
		{
			name:     "french text",
			text:     "Bonjour, ceci est un test en français.",
			wantCode: "FR",
			wantOK:   true,
		},
		{
			name:     "spanish text",
			text:     "Hola, esto es una prueba en español.",
			wantCode: "ES",
			wantOK:   true,
		},
		{
			name:     "polish text",
			text:     "To jest test po polsku.",
			wantCode: "PL",
			wantOK:   true,
		},
		{
			name:     "russian text",
			text:     "Это тест на русском языке.",
			wantCode: "RU",
			wantOK:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := d.DetectISO(tt.text)
			if ok != tt.wantOK {
				t.Errorf("DetectISO(%q) ok = %v, want %v", tt.text, ok, tt.wantOK)
				return
			}
			if tt.wantOK && code != tt.wantCode {
				t.Errorf("DetectISO(%q) = %q, want %q", tt.text, code, tt.wantCode)
			}
		})
	}
}

func TestDetector_ShortText(t *testing.T) {
	d := New()

	code, ok := d.DetectISO("Hi")
	// Short text may or may not be detected, just check it doesn't panic
	_ = code
	_ = ok
}

package detector

import (
	lingua "github.com/pemistahl/lingua-go"
)

type Detector struct {
	detector lingua.LanguageDetector
}

func New() *Detector {
	detector := lingua.NewLanguageDetectorBuilder().
		FromAllLanguages().
		Build()

	return &Detector{detector: detector}
}

func (d *Detector) Detect(text string) (lingua.Language, bool) {
	if text == "" {
		return lingua.Unknown, false
	}
	return d.detector.DetectLanguageOf(text)
}

func (d *Detector) DetectISO(text string) (string, bool) {
	lang, ok := d.Detect(text)
	if !ok {
		return "", false
	}
	return lang.IsoCode639_1().String(), true
}

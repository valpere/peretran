package internal

import "time"

type TranslationRequest struct {
	ID         string    `json:"id"`
	SourceText string    `json:"source_text"`
	SourceLang string    `json:"source_lang"`
	TargetLang string    `json:"target_lang"`
	Timestamp  time.Time `json:"timestamp"`
}

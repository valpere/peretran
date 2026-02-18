package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/valpere/peretran/internal"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return s, nil
}

func (s *Store) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS translation_requests (
		id TEXT PRIMARY KEY,
		source_text TEXT NOT NULL,
		source_lang TEXT NOT NULL,
		target_lang TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS translation_results (
		id TEXT PRIMARY KEY,
		request_id TEXT NOT NULL,
		service_name TEXT NOT NULL,
		translated_text TEXT NOT NULL,
		confidence REAL,
		latency_ms INTEGER,
		error TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (request_id) REFERENCES translation_requests(id)
	);

	CREATE TABLE IF NOT EXISTS final_translations (
		id TEXT PRIMARY KEY,
		request_id TEXT NOT NULL,
		selected_service TEXT,
		final_text TEXT NOT NULL,
		is_composite BOOLEAN DEFAULT FALSE,
		arbiter_reasoning TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (request_id) REFERENCES translation_requests(id)
	);

	CREATE TABLE IF NOT EXISTS translation_memory (
		id TEXT PRIMARY KEY,
		source_text TEXT NOT NULL,
		source_lang TEXT NOT NULL,
		target_lang TEXT NOT NULL,
		final_text TEXT NOT NULL,
		service_used TEXT,
		usage_count INTEGER DEFAULT 1,
		invalidated BOOLEAN DEFAULT FALSE,
		last_used TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_text, source_lang, target_lang)
	);

	CREATE INDEX IF NOT EXISTS idx_memory_lookup ON translation_memory(source_text, source_lang, target_lang);
	CREATE INDEX IF NOT EXISTS idx_results_request ON translation_results(request_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) SaveRequest(ctx context.Context, req internal.TranslationRequest) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO translation_requests (id, source_text, source_lang, target_lang, created_at) VALUES (?, ?, ?, ?, ?)`,
		req.ID, req.SourceText, req.SourceLang, req.TargetLang, req.Timestamp)
	return err
}

func (s *Store) SaveResult(ctx context.Context, requestID, serviceName, translatedText string, confidence float64, latencyMs int, errMsg string) error {
	id := fmt.Sprintf("%s_%s", requestID, serviceName)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO translation_results (id, request_id, service_name, translated_text, confidence, latency_ms, error) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, requestID, serviceName, translatedText, confidence, latencyMs, errMsg)
	return err
}

func (s *Store) SaveFinalTranslation(ctx context.Context, requestID, selectedService, finalText string, isComposite bool, reasoning string) error {
	id := fmt.Sprintf("%s_final", requestID)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO final_translations (id, request_id, selected_service, final_text, is_composite, arbiter_reasoning) VALUES (?, ?, ?, ?, ?, ?)`,
		id, requestID, selectedService, finalText, isComposite, reasoning)
	return err
}

func (s *Store) GetCachedTranslation(ctx context.Context, sourceText, sourceLang, targetLang string) (string, bool, error) {
	var finalText string
	var invalidated bool

	err := s.db.QueryRowContext(ctx,
		`SELECT final_text, invalidated FROM translation_memory WHERE source_text = ? AND source_lang = ? AND target_lang = ?`,
		normalizeText(sourceText), sourceLang, targetLang).Scan(&finalText, &invalidated)

	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	if invalidated {
		return "", false, nil
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE translation_memory SET usage_count = usage_count + 1, last_used = ? WHERE source_text = ? AND source_lang = ? AND target_lang = ?`,
		time.Now(), normalizeText(sourceText), sourceLang, targetLang)

	return finalText, true, err
}

func (s *Store) SaveToMemory(ctx context.Context, sourceText, sourceLang, targetLang, finalText, serviceUsed string) error {
	id := fmt.Sprintf("mem_%d", time.Now().UnixNano())
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO translation_memory (id, source_text, source_lang, target_lang, final_text, service_used, usage_count, invalidated, last_used, created_at) VALUES (?, ?, ?, ?, ?, ?, 1, FALSE, ?, ?)`,
		id, normalizeText(sourceText), sourceLang, targetLang, finalText, serviceUsed, time.Now(), time.Now())
	return err
}

func (s *Store) InvalidateMemory(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE translation_memory SET invalidated = TRUE WHERE id = ?`, id)
	return err
}

func (s *Store) ListMemory(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, source_text, source_lang, target_lang, final_text, service_used, usage_count, invalidated, last_used FROM translation_memory ORDER BY last_used DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, sourceText, sourceLang, targetLang, finalText, serviceUsed string
		var usageCount int
		var invalidated bool
		var lastUsed time.Time

		if err := rows.Scan(&id, &sourceText, &sourceLang, &targetLang, &finalText, &serviceUsed, &usageCount, &invalidated, &lastUsed); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"id":           id,
			"source_text":  sourceText,
			"source_lang":  sourceLang,
			"target_lang":  targetLang,
			"final_text":   finalText,
			"service_used": serviceUsed,
			"usage_count":  usageCount,
			"invalidated":  invalidated,
			"last_used":    lastUsed,
		})
	}

	return results, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func normalizeText(text string) string {
	return text
}

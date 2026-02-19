package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
	"golang.org/x/text/unicode/norm"

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
		draft_text TEXT,
		service_used TEXT,
		usage_count INTEGER DEFAULT 1,
		invalidated BOOLEAN DEFAULT FALSE,
		last_used TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_text, source_lang, target_lang)
	);

	-- stage1_cache stores primary translation drafts (pre-refinement)
	CREATE TABLE IF NOT EXISTS stage1_cache (
		id TEXT PRIMARY KEY,
		source_text TEXT NOT NULL,
		source_lang TEXT NOT NULL,
		target_lang TEXT NOT NULL,
		draft_text TEXT NOT NULL,
		service_used TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_used TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_text, source_lang, target_lang, service_used)
	);

	-- csv_checkpoints tracks progress of CSV translation jobs for resume support
	CREATE TABLE IF NOT EXISTS csv_checkpoints (
		id TEXT PRIMARY KEY,
		input_file TEXT NOT NULL,
		output_file TEXT NOT NULL,
		source_lang TEXT NOT NULL,
		target_lang TEXT NOT NULL,
		status TEXT DEFAULT 'running',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- csv_checkpoint_cells stores per-cell translated results
	CREATE TABLE IF NOT EXISTS csv_checkpoint_cells (
		checkpoint_id TEXT NOT NULL,
		row_idx INTEGER NOT NULL,
		col_idx INTEGER NOT NULL,
		translated_text TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (checkpoint_id, row_idx, col_idx),
		FOREIGN KEY (checkpoint_id) REFERENCES csv_checkpoints(id)
	);

	-- glossary stores user-defined terminology for consistent translation of specific terms
	CREATE TABLE IF NOT EXISTS glossary (
		id TEXT PRIMARY KEY,
		source_lang TEXT NOT NULL,
		target_lang TEXT NOT NULL,
		source_term TEXT NOT NULL,
		target_term TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_lang, target_lang, source_term)
	);

	CREATE INDEX IF NOT EXISTS idx_memory_lookup ON translation_memory(source_text, source_lang, target_lang);
	CREATE INDEX IF NOT EXISTS idx_stage1_lookup ON stage1_cache(source_text, source_lang, target_lang);
	CREATE INDEX IF NOT EXISTS idx_results_request ON translation_results(request_id);
	CREATE INDEX IF NOT EXISTS idx_checkpoint_cells ON csv_checkpoint_cells(checkpoint_id);
	CREATE INDEX IF NOT EXISTS idx_glossary_lookup ON glossary(source_lang, target_lang);
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

func (s *Store) SaveToMemory(ctx context.Context, sourceText, sourceLang, targetLang, finalText, draftText, serviceUsed string) error {
	id := fmt.Sprintf("mem_%d", time.Now().UnixNano())
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO translation_memory (id, source_text, source_lang, target_lang, final_text, draft_text, service_used, usage_count, invalidated, last_used, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, 1, FALSE, ?, ?)`,
		id, normalizeText(sourceText), sourceLang, targetLang, finalText, draftText, serviceUsed, time.Now(), time.Now())
	return err
}

// SaveToStage1Cache stores the primary (pre-refinement) translation draft.
func (s *Store) SaveToStage1Cache(ctx context.Context, sourceText, sourceLang, targetLang, draftText, serviceUsed string) error {
	id := fmt.Sprintf("s1_%d", time.Now().UnixNano())
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO stage1_cache (id, source_text, source_lang, target_lang, draft_text, service_used, created_at, last_used) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, normalizeText(sourceText), sourceLang, targetLang, draftText, serviceUsed, time.Now(), time.Now())
	return err
}

// GetStage1Draft returns a cached stage1 draft if available.
func (s *Store) GetStage1Draft(ctx context.Context, sourceText, sourceLang, targetLang, serviceUsed string) (string, bool, error) {
	var draftText string
	err := s.db.QueryRowContext(ctx,
		`SELECT draft_text FROM stage1_cache WHERE source_text = ? AND source_lang = ? AND target_lang = ? AND service_used = ?`,
		normalizeText(sourceText), sourceLang, targetLang, serviceUsed).Scan(&draftText)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	_, _ = s.db.ExecContext(ctx,
		`UPDATE stage1_cache SET last_used = ? WHERE source_text = ? AND source_lang = ? AND target_lang = ? AND service_used = ?`,
		time.Now(), normalizeText(sourceText), sourceLang, targetLang, serviceUsed)
	return draftText, true, nil
}

// MemoryEntry is a row from the translation_memory table.
type MemoryEntry struct {
	ID          string
	SourceText  string
	SourceLang  string
	TargetLang  string
	FinalText   string
	ServiceUsed string
	UsageCount  int
	Invalidated bool
	LastUsed    time.Time
}

// CacheStats summarises translation memory usage.
type CacheStats struct {
	TotalEntries   int
	ActiveEntries  int
	InvalidEntries int
	TotalUsage     int
}

func (s *Store) InvalidateMemory(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE translation_memory SET invalidated = TRUE WHERE id = ?`, id)
	return err
}

// DeleteMemory permanently removes a translation memory entry by ID.
func (s *Store) DeleteMemory(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM translation_memory WHERE id = ?`, id)
	return err
}

// ClearMemory removes all translation memory entries.
func (s *Store) ClearMemory(ctx context.Context) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM translation_memory`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ListMemory returns all translation memory entries ordered by most recently used.
func (s *Store) ListMemory(ctx context.Context) ([]MemoryEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, source_text, source_lang, target_lang, final_text, service_used, usage_count, invalidated, last_used FROM translation_memory ORDER BY last_used DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []MemoryEntry
	for rows.Next() {
		var e MemoryEntry
		if err := rows.Scan(&e.ID, &e.SourceText, &e.SourceLang, &e.TargetLang, &e.FinalText, &e.ServiceUsed, &e.UsageCount, &e.Invalidated, &e.LastUsed); err != nil {
			return nil, err
		}
		results = append(results, e)
	}

	return results, rows.Err()
}

// Stats returns summary statistics for the translation memory.
func (s *Store) Stats(ctx context.Context) (*CacheStats, error) {
	stats := &CacheStats{}

	err := s.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN NOT invalidated THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN invalidated THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(usage_count), 0)
		FROM translation_memory`).Scan(
		&stats.TotalEntries,
		&stats.ActiveEntries,
		&stats.InvalidEntries,
		&stats.TotalUsage,
	)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

// CSVCheckpoint represents a CSV translation job's checkpoint record.
type CSVCheckpoint struct {
	ID         string
	InputFile  string
	OutputFile string
	SourceLang string
	TargetLang string
	Status     string
	CreatedAt  time.Time
}

// CreateCSVCheckpoint creates a new checkpoint record and returns its ID.
func (s *Store) CreateCSVCheckpoint(ctx context.Context, inputFile, outputFile, sourceLang, targetLang string) (string, error) {
	id := fmt.Sprintf("cp_%d", time.Now().UnixNano())
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO csv_checkpoints (id, input_file, output_file, source_lang, target_lang) VALUES (?, ?, ?, ?, ?)`,
		id, inputFile, outputFile, sourceLang, targetLang)
	return id, err
}

// GetCSVCheckpoint retrieves a checkpoint by ID.
func (s *Store) GetCSVCheckpoint(ctx context.Context, checkpointID string) (*CSVCheckpoint, error) {
	var cp CSVCheckpoint
	err := s.db.QueryRowContext(ctx,
		`SELECT id, input_file, output_file, source_lang, target_lang, status, created_at FROM csv_checkpoints WHERE id = ?`,
		checkpointID).Scan(&cp.ID, &cp.InputFile, &cp.OutputFile, &cp.SourceLang, &cp.TargetLang, &cp.Status, &cp.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("checkpoint not found: %s", checkpointID)
	}
	return &cp, err
}

// SaveCSVCell persists the translated text for a single CSV cell.
func (s *Store) SaveCSVCell(ctx context.Context, checkpointID string, rowIdx, colIdx int, translatedText string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO csv_checkpoint_cells (checkpoint_id, row_idx, col_idx, translated_text) VALUES (?, ?, ?, ?)`,
		checkpointID, rowIdx, colIdx, translatedText)
	return err
}

// GetCSVCells returns all already-translated cells for a checkpoint as a "row:col" → text map.
func (s *Store) GetCSVCells(ctx context.Context, checkpointID string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT row_idx, col_idx, translated_text FROM csv_checkpoint_cells WHERE checkpoint_id = ?`,
		checkpointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cells := make(map[string]string)
	for rows.Next() {
		var rowIdx, colIdx int
		var translatedText string
		if err := rows.Scan(&rowIdx, &colIdx, &translatedText); err != nil {
			return nil, err
		}
		cells[fmt.Sprintf("%d:%d", rowIdx, colIdx)] = translatedText
	}
	return cells, rows.Err()
}

// CompleteCSVCheckpoint marks a checkpoint as completed.
func (s *Store) CompleteCSVCheckpoint(ctx context.Context, checkpointID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE csv_checkpoints SET status = 'completed', updated_at = ? WHERE id = ?`,
		time.Now(), checkpointID)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}

// normalizeText trims whitespace and applies Unicode NFC normalization
// for consistent cache key comparison.
func normalizeText(text string) string {
	return norm.NFC.String(strings.TrimSpace(text))
}

// levenshtein returns the edit distance between two strings (rune-aware).
// Uses a space-optimized two-row DP implementation.
func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			if ra[i-1] == rb[j-1] {
				curr[j] = prev[j-1]
			} else {
				min := prev[j]
				if prev[j-1] < min {
					min = prev[j-1]
				}
				if curr[j-1] < min {
					min = curr[j-1]
				}
				curr[j] = min + 1
			}
		}
		prev, curr = curr, prev
	}

	return prev[lb]
}

// stringSimilarity returns a similarity score in [0, 1] (1 = identical).
func stringSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}
	la, lb := len([]rune(a)), len([]rune(b))
	maxLen := la
	if lb > maxLen {
		maxLen = lb
	}
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(levenshtein(a, b))/float64(maxLen)
}

// FuzzyGetCachedTranslation returns a cached translation whose normalised source
// text has at least threshold similarity (0–1) to sourceText. Pass threshold ≤ 0
// to disable (always returns "", false, nil). To avoid O(n²) cost, texts longer
// than 1 000 runes are not fuzzy-matched.
func (s *Store) FuzzyGetCachedTranslation(ctx context.Context, sourceText, sourceLang, targetLang string, threshold float64) (string, bool, error) {
	if threshold <= 0 {
		return "", false, nil
	}

	normalized := normalizeText(sourceText)
	const maxFuzzyRunes = 1000
	if len([]rune(normalized)) > maxFuzzyRunes {
		return "", false, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT source_text, final_text FROM translation_memory
		 WHERE source_lang = ? AND target_lang = ? AND NOT invalidated`,
		sourceLang, targetLang)
	if err != nil {
		return "", false, err
	}
	defer rows.Close()

	var bestFinal string
	bestScore := 0.0

	for rows.Next() {
		var srcText, finalText string
		if err := rows.Scan(&srcText, &finalText); err != nil {
			return "", false, err
		}

		// Quick length pre-filter: if the length difference alone makes it
		// impossible to reach the threshold, skip the expensive edit distance.
		ls, lr := len([]rune(normalized)), len([]rune(srcText))
		maxL := ls
		if lr > maxL {
			maxL = lr
		}
		diff := ls - lr
		if diff < 0 {
			diff = -diff
		}
		if maxL > 0 && 1.0-float64(diff)/float64(maxL) < threshold {
			continue
		}

		score := stringSimilarity(normalized, srcText)
		if score >= threshold && score > bestScore {
			bestScore = score
			bestFinal = finalText
		}
	}
	if err := rows.Err(); err != nil {
		return "", false, err
	}

	if bestFinal != "" {
		return bestFinal, true, nil
	}
	return "", false, nil
}

// GlossaryEntry represents a row in the glossary table.
type GlossaryEntry struct {
	ID         string
	SourceLang string
	TargetLang string
	SourceTerm string
	TargetTerm string
	CreatedAt  time.Time
}

// AddGlossaryTerm inserts or replaces a glossary entry.
func (s *Store) AddGlossaryTerm(ctx context.Context, sourceLang, targetLang, sourceTerm, targetTerm string) error {
	id := fmt.Sprintf("gl_%d", time.Now().UnixNano())
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO glossary (id, source_lang, target_lang, source_term, target_term)
		 VALUES (?, ?, ?, ?, ?)`,
		id, sourceLang, targetLang, sourceTerm, targetTerm)
	return err
}

// GetGlossaryTerms returns all active glossary terms for a language pair as a
// source-term → target-term map, ready to embed in a translation prompt.
func (s *Store) GetGlossaryTerms(ctx context.Context, sourceLang, targetLang string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT source_term, target_term FROM glossary WHERE source_lang = ? AND target_lang = ?`,
		sourceLang, targetLang)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	terms := make(map[string]string)
	for rows.Next() {
		var src, tgt string
		if err := rows.Scan(&src, &tgt); err != nil {
			return nil, err
		}
		terms[src] = tgt
	}
	return terms, rows.Err()
}

// ListGlossaryTerms returns all glossary entries, optionally filtered by language
// pair (pass empty strings to return everything).
func (s *Store) ListGlossaryTerms(ctx context.Context, sourceLang, targetLang string) ([]GlossaryEntry, error) {
	query := `SELECT id, source_lang, target_lang, source_term, target_term, created_at FROM glossary`
	var args []interface{}

	switch {
	case sourceLang != "" && targetLang != "":
		query += ` WHERE source_lang = ? AND target_lang = ?`
		args = append(args, sourceLang, targetLang)
	case sourceLang != "":
		query += ` WHERE source_lang = ?`
		args = append(args, sourceLang)
	case targetLang != "":
		query += ` WHERE target_lang = ?`
		args = append(args, targetLang)
	}
	query += ` ORDER BY source_lang, target_lang, source_term`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []GlossaryEntry
	for rows.Next() {
		var e GlossaryEntry
		if err := rows.Scan(&e.ID, &e.SourceLang, &e.TargetLang, &e.SourceTerm, &e.TargetTerm, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// DeleteGlossaryTerm removes a glossary entry by ID.
func (s *Store) DeleteGlossaryTerm(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM glossary WHERE id = ?`, id)
	return err
}

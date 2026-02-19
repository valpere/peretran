package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/valpere/peretran/internal"
)

func TestStore_New(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	if s == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestStore_New_InvalidPath(t *testing.T) {
	_, err := New("/nonexistent/path/test.db")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestStore_SaveRequest(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	req := internal.TranslationRequest{
		ID:         "test-req-1",
		SourceText: "Hello world",
		SourceLang: "en",
		TargetLang: "uk",
		Timestamp:  time.Now(),
	}

	err = s.SaveRequest(context.Background(), req)
	if err != nil {
		t.Errorf("SaveRequest failed: %v", err)
	}
}

func TestStore_SaveResult(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// First save a request
	req := internal.TranslationRequest{
		ID:         "test-req-1",
		SourceText: "Hello world",
		SourceLang: "en",
		TargetLang: "uk",
		Timestamp:  time.Now(),
	}
	err = s.SaveRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("SaveRequest failed: %v", err)
	}

	// Then save a result
	err = s.SaveResult(context.Background(), "test-req-1", "google", "Привіт світ", 0.95, 150, "")
	if err != nil {
		t.Errorf("SaveResult failed: %v", err)
	}
}

func TestStore_SaveFinalTranslation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// First save a request
	req := internal.TranslationRequest{
		ID:         "test-req-1",
		SourceText: "Hello world",
		SourceLang: "en",
		TargetLang: "uk",
		Timestamp:  time.Now(),
	}
	err = s.SaveRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("SaveRequest failed: %v", err)
	}

	err = s.SaveFinalTranslation(context.Background(), "test-req-1", "google", "Привіт світ", false, "Selected best result")
	if err != nil {
		t.Errorf("SaveFinalTranslation failed: %v", err)
	}
}

func TestStore_GetCachedTranslation_Miss(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	text, found, err := s.GetCachedTranslation(context.Background(), "Hello", "en", "uk")
	if err != nil {
		t.Errorf("GetCachedTranslation failed: %v", err)
	}
	if found {
		t.Error("expected not found for uncached translation")
	}
	if text != "" {
		t.Errorf("expected empty text, got %q", text)
	}
}

func TestStore_GetCachedTranslation_Hit(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Save to memory
	err = s.SaveToMemory(context.Background(), "Hello", "en", "uk", "Привіт", "", "google")
	if err != nil {
		t.Fatalf("SaveToMemory failed: %v", err)
	}

	// Retrieve from cache
	text, found, err := s.GetCachedTranslation(context.Background(), "Hello", "en", "uk")
	if err != nil {
		t.Errorf("GetCachedTranslation failed: %v", err)
	}
	if !found {
		t.Error("expected to find cached translation")
	}
	if text != "Привіт" {
		t.Errorf("expected 'Привіт', got %q", text)
	}
}

func TestStore_GetCachedTranslation_Invalidated(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Save to memory
	err = s.SaveToMemory(context.Background(), "Hello", "en", "uk", "Привіт", "", "google")
	if err != nil {
		t.Fatalf("SaveToMemory failed: %v", err)
	}

	// Get the ID
	entries, err := s.ListMemory(context.Background())
	if err != nil {
		t.Fatalf("ListMemory failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one entry")
	}

	// Invalidate it
	err = s.InvalidateMemory(context.Background(), entries[0].ID)
	if err != nil {
		t.Fatalf("InvalidateMemory failed: %v", err)
	}

	// Should not be found now
	text, found, err := s.GetCachedTranslation(context.Background(), "Hello", "en", "uk")
	if err != nil {
		t.Errorf("GetCachedTranslation failed: %v", err)
	}
	if found {
		t.Error("expected not found for invalidated translation")
	}
	if text != "" {
		t.Errorf("expected empty text, got %q", text)
	}
}

func TestStore_SaveToStage1Cache(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	err = s.SaveToStage1Cache(context.Background(), "Hello", "en", "uk", "Draft translation", "google")
	if err != nil {
		t.Errorf("SaveToStage1Cache failed: %v", err)
	}

	// Retrieve it
	draft, found, err := s.GetStage1Draft(context.Background(), "Hello", "en", "uk", "google")
	if err != nil {
		t.Errorf("GetStage1Draft failed: %v", err)
	}
	if !found {
		t.Error("expected to find stage1 draft")
	}
	if draft != "Draft translation" {
		t.Errorf("expected 'Draft translation', got %q", draft)
	}
}

func TestStore_Stats(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Empty stats
	stats, err := s.Stats(context.Background())
	if err != nil {
		t.Errorf("Stats failed: %v", err)
	}
	if stats.TotalEntries != 0 {
		t.Errorf("expected 0 total entries, got %d", stats.TotalEntries)
	}

	// Add some memory entries
	s.SaveToMemory(context.Background(), "Hello", "en", "uk", "Привіт", "", "google")
	s.SaveToMemory(context.Background(), "World", "en", "uk", "Світ", "", "google")

	stats, err = s.Stats(context.Background())
	if err != nil {
		t.Errorf("Stats failed: %v", err)
	}
	if stats.TotalEntries != 2 {
		t.Errorf("expected 2 total entries, got %d", stats.TotalEntries)
	}
	if stats.ActiveEntries != 2 {
		t.Errorf("expected 2 active entries, got %d", stats.ActiveEntries)
	}
}

func TestStore_DeleteMemory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Add memory
	s.SaveToMemory(context.Background(), "Hello", "en", "uk", "Привіт", "", "google")

	// Get ID
	entries, err := s.ListMemory(context.Background())
	if err != nil {
		t.Fatalf("ListMemory failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one entry")
	}

	// Delete it
	err = s.DeleteMemory(context.Background(), entries[0].ID)
	if err != nil {
		t.Errorf("DeleteMemory failed: %v", err)
	}

	// Verify gone
	entries, err = s.ListMemory(context.Background())
	if err != nil {
		t.Fatalf("ListMemory failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after delete, got %d", len(entries))
	}
}

func TestStore_ClearMemory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Add memory
	s.SaveToMemory(context.Background(), "Hello", "en", "uk", "Привіт", "", "google")
	s.SaveToMemory(context.Background(), "World", "en", "uk", "Світ", "", "google")

	// Clear all
	count, err := s.ClearMemory(context.Background())
	if err != nil {
		t.Errorf("ClearMemory failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 cleared, got %d", count)
	}

	// Verify empty
	entries, err := s.ListMemory(context.Background())
	if err != nil {
		t.Fatalf("ListMemory failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after clear, got %d", len(entries))
	}
}

func TestStore_CSVCheckpoint(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Create checkpoint
	cpID, err := s.CreateCSVCheckpoint(context.Background(), "input.csv", "output.csv", "en", "uk")
	if err != nil {
		t.Fatalf("CreateCSVCheckpoint failed: %v", err)
	}

	// Get checkpoint
	cp, err := s.GetCSVCheckpoint(context.Background(), cpID)
	if err != nil {
		t.Fatalf("GetCSVCheckpoint failed: %v", err)
	}
	if cp.InputFile != "input.csv" {
		t.Errorf("expected input.csv, got %q", cp.InputFile)
	}
	if cp.Status != "running" {
		t.Errorf("expected running status, got %q", cp.Status)
	}

	// Save cell
	err = s.SaveCSVCell(context.Background(), cpID, 0, 1, "Translated cell")
	if err != nil {
		t.Errorf("SaveCSVCell failed: %v", err)
	}

	// Get cells
	cells, err := s.GetCSVCells(context.Background(), cpID)
	if err != nil {
		t.Fatalf("GetCSVCells failed: %v", err)
	}
	if cells["0:1"] != "Translated cell" {
		t.Errorf("expected 'Translated cell', got %q", cells["0:1"])
	}

	// Complete checkpoint
	err = s.CompleteCSVCheckpoint(context.Background(), cpID)
	if err != nil {
		t.Errorf("CompleteCSVCheckpoint failed: %v", err)
	}

	// Verify completed
	cp, err = s.GetCSVCheckpoint(context.Background(), cpID)
	if err != nil {
		t.Fatalf("GetCSVCheckpoint failed: %v", err)
	}
	if cp.Status != "completed" {
		t.Errorf("expected completed status, got %q", cp.Status)
	}
}

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  Hello  ", "Hello"},
		{"Hello\u0020World", "Hello World"}, // NFC normalization
		{"\t\nHello\t\n", "Hello"},
		{"", ""},
	}

	for _, tt := range tests {
		result := normalizeText(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeText(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestStore_MultipleLanguagePairs(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	// Save different language pairs
	s.SaveToMemory(context.Background(), "Hello", "en", "uk", "Привіт", "", "google")
	s.SaveToMemory(context.Background(), "Hello", "en", "de", "Hallo", "", "google")
	s.SaveToMemory(context.Background(), "Hello", "en", "fr", "Bonjour", "", "google")

	// Check each pair
	text, found, _ := s.GetCachedTranslation(context.Background(), "Hello", "en", "uk")
	if !found || text != "Привіт" {
		t.Errorf("en->uk: expected found=true and 'Привіт', got found=%v and %q", found, text)
	}

	text, found, _ = s.GetCachedTranslation(context.Background(), "Hello", "en", "de")
	if !found || text != "Hallo" {
		t.Errorf("en->de: expected found=true and 'Hallo', got found=%v and %q", found, text)
	}

	text, found, _ = s.GetCachedTranslation(context.Background(), "Hello", "en", "fr")
	if !found || text != "Bonjour" {
		t.Errorf("en->fr: expected found=true and 'Bonjour', got found=%v and %q", found, text)
	}

	// Non-existent pair
	text, found, _ = s.GetCachedTranslation(context.Background(), "Hello", "en", "es")
	if found {
		t.Error("en->es: expected not found")
	}
}

// --- Fuzzy matching tests ---

func TestStore_FuzzyGetCachedTranslation_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	s, _ := New(filepath.Join(tmpDir, "test.db"))
	defer s.Close()

	s.SaveToMemory(context.Background(), "Hello world", "en", "uk", "Привіт світ", "", "google")

	text, found, err := s.FuzzyGetCachedTranslation(context.Background(), "Hello world", "en", "uk", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected not found when threshold=0 (disabled)")
	}
	_ = text
}

func TestStore_FuzzyGetCachedTranslation_ExactMatch(t *testing.T) {
	tmpDir := t.TempDir()
	s, _ := New(filepath.Join(tmpDir, "test.db"))
	defer s.Close()

	s.SaveToMemory(context.Background(), "Hello world", "en", "uk", "Привіт світ", "", "google")

	text, found, err := s.FuzzyGetCachedTranslation(context.Background(), "Hello world", "en", "uk", 0.85)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected found for exact match via fuzzy")
	}
	if text != "Привіт світ" {
		t.Errorf("expected 'Привіт світ', got %q", text)
	}
}

func TestStore_FuzzyGetCachedTranslation_NearMatch(t *testing.T) {
	tmpDir := t.TempDir()
	s, _ := New(filepath.Join(tmpDir, "test.db"))
	defer s.Close()

	// "Hello world!" vs "Hello world" — edit distance 1, max_len 12, similarity ~0.917
	s.SaveToMemory(context.Background(), "Hello world!", "en", "uk", "Привіт, світ!", "", "google")

	// Same text with minor punctuation difference — should match at 0.85 threshold
	text, found, err := s.FuzzyGetCachedTranslation(context.Background(), "Hello world", "en", "uk", 0.85)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("expected fuzzy match for near-identical text")
	}
	if text != "Привіт, світ!" {
		t.Errorf("expected 'Привіт, світ!', got %q", text)
	}
}

func TestStore_FuzzyGetCachedTranslation_TooLowSimilarity(t *testing.T) {
	tmpDir := t.TempDir()
	s, _ := New(filepath.Join(tmpDir, "test.db"))
	defer s.Close()

	s.SaveToMemory(context.Background(), "The quick brown fox", "en", "uk", "Швидка руда лисиця", "", "google")

	// Completely different text — should not match
	_, found, err := s.FuzzyGetCachedTranslation(context.Background(), "Hello world", "en", "uk", 0.85)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected not found for text with low similarity")
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"Hello", "Hello, world!", 8},
	}

	for _, tt := range tests {
		result := levenshtein(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

func TestStringSimilarity(t *testing.T) {
	tests := []struct {
		a, b     string
		minScore float64
		maxScore float64
	}{
		{"", "", 1.0, 1.0},
		{"abc", "abc", 1.0, 1.0},
		{"Hello world", "Hello world", 1.0, 1.0},
		{"Hello world", "Hello, world!", 0.80, 0.99}, // minor punctuation difference
		{"abc", "xyz", 0.0, 0.1},                     // totally different
	}

	for _, tt := range tests {
		score := stringSimilarity(tt.a, tt.b)
		if score < tt.minScore || score > tt.maxScore {
			t.Errorf("stringSimilarity(%q, %q) = %f, want in [%f, %f]", tt.a, tt.b, score, tt.minScore, tt.maxScore)
		}
	}
}

// --- Glossary tests ---

func TestStore_AddAndGetGlossaryTerms(t *testing.T) {
	tmpDir := t.TempDir()
	s, _ := New(filepath.Join(tmpDir, "test.db"))
	defer s.Close()

	err := s.AddGlossaryTerm(context.Background(), "en", "uk", "Kyiv", "Київ")
	if err != nil {
		t.Fatalf("AddGlossaryTerm failed: %v", err)
	}
	err = s.AddGlossaryTerm(context.Background(), "en", "uk", "Ukraine", "Україна")
	if err != nil {
		t.Fatalf("AddGlossaryTerm failed: %v", err)
	}

	terms, err := s.GetGlossaryTerms(context.Background(), "en", "uk")
	if err != nil {
		t.Fatalf("GetGlossaryTerms failed: %v", err)
	}
	if len(terms) != 2 {
		t.Errorf("expected 2 terms, got %d", len(terms))
	}
	if terms["Kyiv"] != "Київ" {
		t.Errorf("expected 'Київ' for 'Kyiv', got %q", terms["Kyiv"])
	}
	if terms["Ukraine"] != "Україна" {
		t.Errorf("expected 'Україна' for 'Ukraine', got %q", terms["Ukraine"])
	}
}

func TestStore_GetGlossaryTerms_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	s, _ := New(filepath.Join(tmpDir, "test.db"))
	defer s.Close()

	terms, err := s.GetGlossaryTerms(context.Background(), "en", "uk")
	if err != nil {
		t.Fatalf("GetGlossaryTerms failed: %v", err)
	}
	if len(terms) != 0 {
		t.Errorf("expected 0 terms, got %d", len(terms))
	}
}

func TestStore_ListGlossaryTerms(t *testing.T) {
	tmpDir := t.TempDir()
	s, _ := New(filepath.Join(tmpDir, "test.db"))
	defer s.Close()

	s.AddGlossaryTerm(context.Background(), "en", "uk", "Kyiv", "Київ")
	s.AddGlossaryTerm(context.Background(), "en", "de", "Kyiv", "Kiew")

	// List all
	all, err := s.ListGlossaryTerms(context.Background(), "", "")
	if err != nil {
		t.Fatalf("ListGlossaryTerms failed: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 total entries, got %d", len(all))
	}

	// Filter by source language
	enOnly, err := s.ListGlossaryTerms(context.Background(), "en", "uk")
	if err != nil {
		t.Fatalf("ListGlossaryTerms failed: %v", err)
	}
	if len(enOnly) != 1 {
		t.Errorf("expected 1 entry for en->uk, got %d", len(enOnly))
	}
}

func TestStore_DeleteGlossaryTerm(t *testing.T) {
	tmpDir := t.TempDir()
	s, _ := New(filepath.Join(tmpDir, "test.db"))
	defer s.Close()

	s.AddGlossaryTerm(context.Background(), "en", "uk", "Kyiv", "Київ")

	entries, _ := s.ListGlossaryTerms(context.Background(), "en", "uk")
	if len(entries) != 1 {
		t.Fatal("expected 1 entry before delete")
	}

	err := s.DeleteGlossaryTerm(context.Background(), entries[0].ID)
	if err != nil {
		t.Fatalf("DeleteGlossaryTerm failed: %v", err)
	}

	entries, _ = s.ListGlossaryTerms(context.Background(), "en", "uk")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries after delete, got %d", len(entries))
	}
}

func TestStore_AddGlossaryTerm_Upsert(t *testing.T) {
	tmpDir := t.TempDir()
	s, _ := New(filepath.Join(tmpDir, "test.db"))
	defer s.Close()

	s.AddGlossaryTerm(context.Background(), "en", "uk", "Kyiv", "Київ")
	s.AddGlossaryTerm(context.Background(), "en", "uk", "Kyiv", "Кийів") // replace

	terms, _ := s.GetGlossaryTerms(context.Background(), "en", "uk")
	if len(terms) != 1 {
		t.Errorf("expected 1 term after upsert, got %d", len(terms))
	}
	if terms["Kyiv"] != "Кийів" {
		t.Errorf("expected updated value 'Кийів', got %q", terms["Kyiv"])
	}
}


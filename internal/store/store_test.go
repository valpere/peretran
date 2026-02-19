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


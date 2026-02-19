package chunker_test

import (
	"strings"
	"testing"

	"github.com/valpere/peretran/internal/chunker"
)

// --- Chunk tests ---

func TestChunk_ShortText(t *testing.T) {
	text := "Hello, world!"
	chunks := chunker.Chunk(text, 100)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if chunks[0] != text {
		t.Errorf("expected %q, got %q", text, chunks[0])
	}
}

func TestChunk_Unlimited(t *testing.T) {
	text := strings.Repeat("word ", 500)
	chunks := chunker.Chunk(text, 0)
	if len(chunks) != 1 {
		t.Errorf("expected 1 chunk when maxChars=0, got %d", len(chunks))
	}
}

func TestChunk_ParagraphBoundary(t *testing.T) {
	// Two paragraphs joined by blank line; each fits in 50 chars when alone.
	para1 := "First paragraph text here."
	para2 := "Second paragraph text here."
	text := para1 + "\n\n" + para2

	chunks := chunker.Chunk(text, 40)
	if len(chunks) < 2 {
		t.Fatalf("expected ≥2 chunks, got %d: %v", len(chunks), chunks)
	}
	if !strings.Contains(chunks[0], "First") {
		t.Errorf("first chunk should contain 'First': %q", chunks[0])
	}
	if !strings.Contains(chunks[len(chunks)-1], "Second") {
		t.Errorf("last chunk should contain 'Second': %q", chunks[len(chunks)-1])
	}
}

func TestChunk_SentenceBoundary(t *testing.T) {
	text := "First sentence ends here. Second sentence follows. Third sentence."
	// maxChars just enough to hold first two sentences
	chunks := chunker.Chunk(text, 40)
	if len(chunks) < 2 {
		t.Fatalf("expected ≥2 chunks, got %d", len(chunks))
	}
	// Each chunk should be non-empty and trimmed.
	for i, c := range chunks {
		if strings.TrimSpace(c) == "" {
			t.Errorf("chunk %d is empty", i)
		}
		if c != strings.TrimSpace(c) {
			t.Errorf("chunk %d has leading/trailing whitespace: %q", i, c)
		}
	}
}

func TestChunk_WordBoundary(t *testing.T) {
	// No sentence-ending punctuation, no paragraphs.
	text := "one two three four five six seven eight nine ten"
	chunks := chunker.Chunk(text, 20)
	if len(chunks) < 2 {
		t.Fatalf("expected ≥2 chunks, got %d", len(chunks))
	}
	// Reassemble and verify no words are lost.
	rejoined := strings.Join(chunks, " ")
	for _, word := range strings.Fields(text) {
		if !strings.Contains(rejoined, word) {
			t.Errorf("word %q lost after chunking", word)
		}
	}
}

func TestChunk_ReconstructsFullText(t *testing.T) {
	// After joining chunks the full text should be recovered (ignoring trim).
	original := "The quick brown fox jumps over the lazy dog. " +
		"Pack my box with five dozen liquor jugs. " +
		"How vexingly quick daft zebras jump!"
	chunks := chunker.Chunk(original, 50)
	rejoined := strings.Join(chunks, " ")
	// Every word in the original should appear in the rejoined text.
	for _, word := range strings.Fields(original) {
		clean := strings.Trim(word, ".,!?")
		if !strings.Contains(rejoined, clean) {
			t.Errorf("word %q missing after chunk+join", clean)
		}
	}
}

func TestChunk_EmptyText(t *testing.T) {
	chunks := chunker.Chunk("", 100)
	// Either 0 chunks or 1 empty chunk; caller trims anyway.
	for _, c := range chunks {
		if c != "" {
			t.Errorf("expected empty chunk, got %q", c)
		}
	}
}

// --- ExtractContext tests ---

func TestExtractContext_FewerWordsThanLimit(t *testing.T) {
	text := "short text"
	ctx := chunker.ExtractContext(text, 25)
	if ctx != text {
		t.Errorf("expected %q, got %q", text, ctx)
	}
}

func TestExtractContext_ExactWordCount(t *testing.T) {
	words := make([]string, 25)
	for i := range words {
		words[i] = "word"
	}
	text := strings.Join(words, " ")
	ctx := chunker.ExtractContext(text, 25)
	if ctx != text {
		t.Errorf("expected same text back, got %q", ctx)
	}
}

func TestExtractContext_MoreWordsThanLimit(t *testing.T) {
	words := make([]string, 50)
	for i := range words {
		words[i] = "word"
	}
	text := strings.Join(words, " ")
	ctx := chunker.ExtractContext(text, 25)
	got := len(strings.Fields(ctx))
	if got != 25 {
		t.Errorf("expected 25 words, got %d", got)
	}
}

func TestExtractContext_DefaultWordCount(t *testing.T) {
	// wordCount ≤ 0 should use DefaultContextWords (25).
	words := make([]string, 50)
	for i := range words {
		words[i] = "w"
	}
	text := strings.Join(words, " ")
	ctx := chunker.ExtractContext(text, 0)
	got := len(strings.Fields(ctx))
	if got != chunker.DefaultContextWords {
		t.Errorf("expected %d words, got %d", chunker.DefaultContextWords, got)
	}
}

func TestExtractContext_LastWordsCorrect(t *testing.T) {
	text := "alpha beta gamma delta epsilon"
	ctx := chunker.ExtractContext(text, 3)
	if ctx != "gamma delta epsilon" {
		t.Errorf("expected last 3 words, got %q", ctx)
	}
}

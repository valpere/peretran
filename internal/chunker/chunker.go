// Package chunker splits large texts into translatable chunks while
// preserving sentence and paragraph integrity. It also extracts a
// sliding-window context snippet (last N words) for use with LLM
// translators to maintain continuity across chunk boundaries.
package chunker

import (
	"strings"
	"unicode"
)

const (
	// DefaultContextWords is the default number of words extracted by
	// ExtractContext for use as a sliding-window context.
	DefaultContextWords = 25
)

// Chunk splits text into pieces each no longer than maxChars unicode
// code points. Splits are attempted (in order of preference) at:
//  1. Paragraph boundaries (\n\n or \r\n\r\n)
//  2. Sentence-ending punctuation (. ! ?)
//  3. Whitespace (word boundary)
//  4. Hard cut at maxChars if no suitable boundary is found
//
// If text fits entirely within maxChars, a single-element slice is returned.
// If maxChars ≤ 0 it is treated as unlimited (returns the whole text).
func Chunk(text string, maxChars int) []string {
	if maxChars <= 0 || len([]rune(text)) <= maxChars {
		return []string{text}
	}

	var chunks []string
	remaining := text

	for len([]rune(remaining)) > maxChars {
		split := findSplit(remaining, maxChars)
		chunk := strings.TrimSpace(remaining[:split])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
		remaining = strings.TrimSpace(remaining[split:])
	}

	if strings.TrimSpace(remaining) != "" {
		chunks = append(chunks, strings.TrimSpace(remaining))
	}

	return chunks
}

// findSplit returns the byte index within text at which to split, aiming for
// at most maxChars runes. It searches backwards from maxChars for the best
// split boundary.
func findSplit(text string, maxChars int) int {
	runes := []rune(text)
	if len(runes) <= maxChars {
		return len(text)
	}

	// Work with the candidate prefix (runes[:maxChars]).
	// Convert back to byte offset for the split boundary.
	candidate := string(runes[:maxChars])

	// 1. Paragraph boundary — search backwards in candidate.
	if idx := lastIndex(candidate, "\n\n"); idx > 0 {
		return idx + 2 // include the blank line in the consumed part
	}
	if idx := lastIndex(candidate, "\r\n\r\n"); idx > 0 {
		return idx + 4
	}

	// 2. Sentence-ending punctuation followed by a space.
	for i := len([]rune(candidate)) - 1; i > 0; i-- {
		r := []rune(candidate)[i]
		if (r == '.' || r == '!' || r == '?') && i+1 < len([]rune(candidate)) {
			next := []rune(candidate)[i+1]
			if unicode.IsSpace(next) {
				byteOffset := len(string([]rune(candidate)[:i+1]))
				return byteOffset
			}
		}
	}

	// 3. Whitespace word boundary.
	for i := len([]rune(candidate)) - 1; i > 0; i-- {
		if unicode.IsSpace([]rune(candidate)[i]) {
			byteOffset := len(string([]rune(candidate)[:i]))
			return byteOffset
		}
	}

	// 4. Hard cut.
	return len(candidate)
}

// lastIndex returns the last byte index of substr within s, or -1 if not found.
func lastIndex(s, substr string) int {
	idx := -1
	start := 0
	for {
		i := strings.Index(s[start:], substr)
		if i == -1 {
			break
		}
		idx = start + i
		start = idx + 1
	}
	return idx
}

// ExtractContext returns the last wordCount words of text, joined by a single
// space. It is intended for use as a sliding-window context snippet passed to
// LLM translators so they can maintain narrative continuity across chunks.
// If text has fewer words than wordCount, the entire text is returned.
// If wordCount ≤ 0, DefaultContextWords is used.
func ExtractContext(text string, wordCount int) string {
	if wordCount <= 0 {
		wordCount = DefaultContextWords
	}
	words := strings.Fields(text)
	if len(words) <= wordCount {
		return strings.TrimSpace(text)
	}
	return strings.Join(words[len(words)-wordCount:], " ")
}

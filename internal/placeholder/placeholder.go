// Package placeholder protects structured content (HTML tags, fenced code
// blocks, inline code spans) during translation by replacing them with
// numbered markers ([PH0], [PH1], …) that LLMs are instructed to preserve.
// After translation, Restore substitutes the markers back.
package placeholder

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// fenced code blocks: ```...``` (non-greedy, may span lines)
	reFencedCode = regexp.MustCompile("(?s)```.*?```")

	// inline code spans: `...`
	reInlineCode = regexp.MustCompile("`[^`]+`")

	// HTML/XML tags: opening, closing, and self-closing
	reHTMLTag = regexp.MustCompile(`<[^>]+>`)

	// placeholder reference in translated text
	rePlaceholder = regexp.MustCompile(`\[PH(\d+)\]`)
)

// Protect replaces structured markup (fenced code blocks, inline code,
// HTML tags) with numbered placeholders [PH0], [PH1], … in the order they
// appear in text. It returns the modified text and the slice of captured
// originals so Restore can put them back.
func Protect(text string) (string, []string) {
	var markers []string
	counter := 0

	replace := func(match string) string {
		id := fmt.Sprintf("[PH%d]", counter)
		markers = append(markers, match)
		counter++
		return id
	}

	// Order matters: fenced first (longest match), then inline, then HTML tags.
	text = reFencedCode.ReplaceAllStringFunc(text, replace)
	text = reInlineCode.ReplaceAllStringFunc(text, replace)
	text = reHTMLTag.ReplaceAllStringFunc(text, replace)

	return text, markers
}

// Restore substitutes [PHn] markers in text back with the originals captured
// by Protect. Markers missing from the translated text are silently ignored;
// unrecognised indices leave the placeholder as-is.
func Restore(text string, markers []string) string {
	return rePlaceholder.ReplaceAllStringFunc(text, func(match string) string {
		sub := rePlaceholder.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		idx := 0
		fmt.Sscanf(sub[1], "%d", &idx)
		if idx < 0 || idx >= len(markers) {
			return match
		}
		return markers[idx]
	})
}

// InstructionHint returns a short sentence to append to an LLM prompt so the
// model knows to leave placeholders intact.
func InstructionHint() string {
	return "Preserve all [PHn] markers exactly as they appear — do not translate, move, or remove them."
}

// Validate checks whether all markers that were created by Protect are still
// present in the translated text. It returns the list of missing indices.
func Validate(text string, markers []string) []int {
	var missing []int
	for i := range markers {
		if !strings.Contains(text, fmt.Sprintf("[PH%d]", i)) {
			missing = append(missing, i)
		}
	}
	return missing
}

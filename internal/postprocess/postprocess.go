// Package postprocess removes common LLM artifacts from translation output.
//
// It is applied to the raw text returned by any LLM-backed service (Ollama,
// OpenRouter, arbiter, refiner) before the result is used downstream.
package postprocess

import (
	"regexp"
	"strings"
)

// Clean removes LLM artifacts from text in three phases and returns the
// trimmed result:
//  1. Thinking / reasoning block removal
//  2. Instruction echo removal (prompt leakage)
//  3. Quote wrapping removal
func Clean(text string) string {
	text = removeThinkingBlocks(text)
	text = removeInstructionEchoes(text)
	text = removeQuoteWrapping(text)
	return strings.TrimSpace(text)
}

// --- Phase 1: thinking blocks ---

// thinkingBlockRe matches complete <thinking>…</thinking> style blocks.
// Each tag variant is listed explicitly because Go's RE2 engine does not
// support backreferences.
// Flags: i = case-insensitive, s = dot matches newline.
var thinkingBlockRe = regexp.MustCompile(
	`(?is)<thinking>.*?</thinking>|<think>.*?</think>|<reasoning>.*?</reasoning>|<reflection>.*?</reflection>`,
)

// truncatedThinkingRe matches an opened thinking tag whose closing tag is
// missing (the model was cut off mid-thought).
var truncatedThinkingRe = regexp.MustCompile(
	`(?is)(?:<thinking>|<think>|<reasoning>|<reflection>).*$`,
)

func removeThinkingBlocks(text string) string {
	text = thinkingBlockRe.ReplaceAllString(text, "")
	text = truncatedThinkingRe.ReplaceAllString(text, "")
	return strings.TrimSpace(text)
}

// --- Phase 2: instruction echoes ---

// echoPatterns match introductory phrases that LLMs sometimes prepend even
// when instructed not to.  Each pattern is anchored to the start of the string
// and requires a colon to reduce false positives on legitimate content.
var echoPatterns = []*regexp.Regexp{
	// "Here is / Here's [the] [refined|polished|translated] translation:"
	regexp.MustCompile(`(?i)^here(?:'s| is)(?: the)? (?:refined |polished |translated )?(?:translation|text)\s*:`),
	// "[The] [refined|polished] [translation|translated text]:"
	regexp.MustCompile(`(?i)^(?:the )?(?:refined |polished )?(?:translation|translated text)\s*:`),
	// "Certainly / Sure / Of course[,] here is [the] translation:"
	regexp.MustCompile(`(?i)^(?:certainly|sure|of course)[,.]? here(?:'s| is)(?: the)? (?:refined |polished |translated )?(?:translation|text)\s*:`),
}

func removeInstructionEchoes(text string) string {
	for _, re := range echoPatterns {
		if loc := re.FindStringIndex(text); loc != nil && loc[0] == 0 {
			text = strings.TrimSpace(text[loc[1]:])
		}
	}
	return text
}

// --- Phase 3: quote wrapping ---

// removeQuoteWrapping strips a matching pair of outer quotes when the entire
// text is wrapped in them (a common LLM artifact).  Supported pairs:
//
//	"…"  '…'  «…»  "…"  '…'
func removeQuoteWrapping(text string) string {
	runes := []rune(text)
	n := len(runes)
	if n < 2 {
		return text
	}
	first, last := runes[0], runes[n-1]
	if (first == '"' && last == '"') ||
		(first == '\'' && last == '\'') ||
		(first == '«' && last == '»') ||
		(first == '\u201C' && last == '\u201D') || // " "
		(first == '\u2018' && last == '\u2019') { //  ' '
		return strings.TrimSpace(string(runes[1 : n-1]))
	}
	return text
}

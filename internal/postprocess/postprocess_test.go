package postprocess

import "testing"

func TestRemoveThinkingBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no thinking blocks",
			input:    "Hello, this is a normal translation.",
			expected: "Hello, this is a normal translation.",
		},
		{
			name:     "simple thinking block",
			input:    "Some text<thinking>Let me translate this</thinking>More text",
			expected: "Some textMore text",
		},
		{
			name:     "reasoning block",
			input:    "Start<reasoning>Analyzing the grammar</reasoning>End",
			expected: "StartEnd",
		},
		{
			name:     "reflection block",
			input:    "Begin<reflection>Checking context</reflection>Finish",
			expected: "BeginFinish",
		},
		{
			name:     "multiple thinking blocks",
			input:    "<thinking>First</thinking>middle<thinking>Second</thinking>",
			expected: "middle",
		},
		{
			name:     "truncated thinking block (no closing)",
			input:    "<thinking>Translation in progress",
			expected: "",
		},
		{
			name:     "truncated reasoning block",
			input:    "<reasoning>This model was cut off",
			expected: "",
		},
		{
			name:     "truncated thinking in middle",
			input:    "Before<thinking>Incomplete",
			expected: "Before",
		},
		{
			name:     "nested thinking inside content",
			input:    "Text<thinking>Ignored</thinking> after",
			expected: "Text after",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeThinkingBlocks(tt.input)
			if result != tt.expected {
				t.Errorf("removeThinkingBlocks(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveInstructionEchoes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no echo",
			input:    "Just a normal translation.",
			expected: "Just a normal translation.",
		},
		{
			name:     "here's translation echo",
			input:    "Here's the translation: Actual translation text",
			expected: "Actual translation text",
		},
		{
			name:     "here is translation echo",
			input:    "Here is the refined translation: Done",
			expected: "Done",
		},
		{
			name:     "here is translation no the",
			input:    "Here's translation: Text",
			expected: "Text",
		},
		{
			name:     "the translation echo",
			input:    "The translation: Hello world",
			expected: "Hello world",
		},
		{
			name:     "the refined translation echo",
			input:    "The refined translation: Done",
			expected: "Done",
		},
		{
			name:     "certainly echo",
			input:    "Certainly, here's the translation: Text",
			expected: "Text",
		},
		{
			name:     "sure echo",
			input:    "Sure, here's the polished translation: Done",
			expected: "Done",
		},
		{
			name:     "of course echo",
			input:    "Of course here's the refined translation: Text",
			expected: "Text",
		},
		{
			name:     "echo not at start (should not match)",
			input:    "Before Here's the translation: After",
			expected: "Before Here's the translation: After",
		},
		{
			name:     "echo without colon (should not match)",
			input:    "Here's the translation text",
			expected: "Here's the translation text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeInstructionEchoes(tt.input)
			if result != tt.expected {
				t.Errorf("removeInstructionEchoes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveQuoteWrapping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single char",
			input:    "a",
			expected: "a",
		},
		{
			name:     "no quotes",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "double quotes",
			input:    "\"Hello world\"",
			expected: "Hello world",
		},
		{
			name:     "single quotes",
			input:    "'Hello world'",
			expected: "Hello world",
		},
		{
			name:     "guillemets",
			input:    "«Hello world»",
			expected: "Hello world",
		},
		{
			name:     "curly double quotes",
			input:    "\u201CHello world\u201D",
			expected: "Hello world",
		},
		{
			name:     "curly single quotes",
			input:    "\u2018Hello world\u2019",
			expected: "Hello world",
		},
		{
			name:     "unmatched quotes",
			input:    "\"Hello world'",
			expected: "\"Hello world'",
		},
		{
			name:     "only opening quote",
			input:    "\"Hello world",
			expected: "\"Hello world",
		},
		{
			name:     "only closing quote",
			input:    "Hello world\"",
			expected: "Hello world\"",
		},
		{
			name:     "quotes with leading/trailing whitespace",
			input:    "\"  Hello  \"",
			expected: "Hello",
		},
		{
			name:     "content with quotes inside",
			input:    "\"He said \"hello\"\"",
			expected: "He said \"hello\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeQuoteWrapping(tt.input)
			if result != tt.expected {
				t.Errorf("removeQuoteWrapping(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "clean text",
			input:    "Just a normal translation.",
			expected: "Just a normal translation.",
		},
		{
			name:     "full cleanup pipeline",
			input:    "<thinking>Thinking</thinking>Here's the translation:\n\"Translated text\"",
			expected: "Translated text",
		},
		{
			name:     "thinking + echo + quotes",
			input:    "<reasoning>Reasoning</reasoning>Here's the polished translation:\n\"Result\"",
			expected: "Result",
		},
		{
			name:     "truncated thinking at end",
			input:    "Text<thinking>Incomplete",
			expected: "Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clean(tt.input)
			if result != tt.expected {
				t.Errorf("Clean(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

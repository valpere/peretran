package placeholder_test

import (
	"testing"

	"github.com/valpere/peretran/internal/placeholder"
)

func TestProtect_NoMarkup(t *testing.T) {
	text := "Hello, world!"
	got, markers := placeholder.Protect(text)
	if got != text {
		t.Errorf("expected unchanged text, got %q", got)
	}
	if len(markers) != 0 {
		t.Errorf("expected 0 markers, got %d", len(markers))
	}
}

func TestProtect_HTMLTags(t *testing.T) {
	text := "<p>Hello <b>world</b></p>"
	got, markers := placeholder.Protect(text)

	if len(markers) != 4 {
		t.Fatalf("expected 4 markers (<p>, <b>, </b>, </p>), got %d: %v", len(markers), markers)
	}
	// Original tags should no longer appear in the protected text.
	for _, tag := range []string{"<p>", "<b>", "</b>", "</p>"} {
		if contains(got, tag) {
			t.Errorf("expected tag %q to be replaced, still present in %q", tag, got)
		}
	}
}

func TestProtect_FencedCode(t *testing.T) {
	text := "Before\n```go\nfmt.Println(\"hi\")\n```\nAfter"
	got, markers := placeholder.Protect(text)

	if len(markers) != 1 {
		t.Fatalf("expected 1 marker for fenced block, got %d", len(markers))
	}
	if contains(got, "```") {
		t.Errorf("fenced block still present in %q", got)
	}
	if !contains(got, "[PH0]") {
		t.Errorf("expected [PH0] in %q", got)
	}
}

func TestProtect_InlineCode(t *testing.T) {
	text := "Use `fmt.Println` to print."
	got, markers := placeholder.Protect(text)

	if len(markers) != 1 {
		t.Fatalf("expected 1 marker, got %d", len(markers))
	}
	if contains(got, "`fmt.Println`") {
		t.Error("inline code still present after Protect")
	}
	if !contains(got, "[PH0]") {
		t.Errorf("expected [PH0] in %q", got)
	}
}

func TestProtect_Mixed(t *testing.T) {
	text := "See <a href=\"#\">link</a> or use `code` here."
	got, markers := placeholder.Protect(text)

	// 2 HTML tags + 1 inline code = 3 markers
	if len(markers) != 3 {
		t.Fatalf("expected 3 markers, got %d: %v", len(markers), markers)
	}
	_ = got
}

func TestRestore_RoundTrip(t *testing.T) {
	original := "<p>Hello <b>world</b></p>"
	protected, markers := placeholder.Protect(original)

	restored := placeholder.Restore(protected, markers)
	if restored != original {
		t.Errorf("round-trip failed:\n  original:  %q\n  restored:  %q", original, restored)
	}
}

func TestRestore_FencedCodeRoundTrip(t *testing.T) {
	original := "Before\n```go\nfmt.Println(\"hi\")\n```\nAfter"
	protected, markers := placeholder.Protect(original)
	restored := placeholder.Restore(protected, markers)
	if restored != original {
		t.Errorf("round-trip failed:\n  original: %q\n  restored: %q", original, restored)
	}
}

func TestRestore_OutOfRangeIndexIgnored(t *testing.T) {
	// A translated text that invents a placeholder index that doesn't exist.
	text := "[PH99] some text"
	restored := placeholder.Restore(text, []string{"<p>"})
	// [PH99] should remain as-is since index 99 is out of range.
	if !contains(restored, "[PH99]") {
		t.Errorf("expected [PH99] to remain, got %q", restored)
	}
}

func TestRestore_MissingMarkerIgnored(t *testing.T) {
	// Simulates an LLM that dropped [PH1] from the translation.
	original := "<p>Hello</p> <b>world</b>"
	protected, markers := placeholder.Protect(original)

	// Manually remove [PH1] from the protected string.
	withoutPH1 := removeSubstring(protected, "[PH1]")

	// Restore should handle the missing marker gracefully.
	restored := placeholder.Restore(withoutPH1, markers)
	// [PH1] is gone — the corresponding tag won't appear; no panic or error.
	_ = restored
}

func TestValidate_AllPresent(t *testing.T) {
	text := "[PH0] some [PH1] text"
	markers := []string{"<p>", "</p>"}
	missing := placeholder.Validate(text, markers)
	if len(missing) != 0 {
		t.Errorf("expected no missing, got %v", missing)
	}
}

func TestValidate_SomeMissing(t *testing.T) {
	text := "[PH0] some text"
	markers := []string{"<p>", "</p>", "<b>"}
	missing := placeholder.Validate(text, markers)
	if len(missing) != 2 {
		t.Errorf("expected 2 missing (indices 1,2), got %v", missing)
	}
	if missing[0] != 1 || missing[1] != 2 {
		t.Errorf("expected missing [1 2], got %v", missing)
	}
}

func TestInstructionHint_NotEmpty(t *testing.T) {
	hint := placeholder.InstructionHint()
	if hint == "" {
		t.Error("InstructionHint should not return empty string")
	}
}

// helpers

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func removeSubstring(s, sub string) string {
	idx := -1
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			idx = i
			break
		}
	}
	if idx == -1 {
		return s
	}
	return s[:idx] + s[idx+len(sub):]
}

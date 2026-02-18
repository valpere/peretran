package refiner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/valpere/peretran/internal/postprocess"
)

// OllamaRefiner uses a local Ollama model as a literary editor for Stage 2.
type OllamaRefiner struct {
	model   string
	baseURL string
	client  *http.Client
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

// NewOllamaRefiner creates a refiner backed by a local Ollama model.
func NewOllamaRefiner(model, baseURL string) *OllamaRefiner {
	return &OllamaRefiner{
		model:   model,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// Refine sends the draft to the LLM with a literary-editor prompt and returns
// the polished translation.
func (r *OllamaRefiner) Refine(ctx context.Context, sourceLang, targetLang, sourceText, draftText string) (string, error) {
	prompt := buildRefinementPrompt(sourceLang, targetLang, sourceText, draftText)

	reqBody := ollamaRequest{
		Model:  r.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal refinement request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/generate", r.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create refinement request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("refinement request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("refiner returned status %d", resp.StatusCode)
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode refinement response: %w", err)
	}

	refined := postprocess.Clean(ollamaResp.Response)
	if refined == "" {
		return draftText, nil
	}
	return refined, nil
}

func buildRefinementPrompt(sourceLang, targetLang, sourceText, draftText string) string {
	return fmt.Sprintf(`You are an elite %s literary editor and prose stylist.

# YOUR TASK: REFINE AND POLISH

You will receive a DRAFT %s translation that needs improvement.
Your job is to REWRITE it with perfect literary %s style.

ORIGINAL (%s):
%s

DRAFT TRANSLATION (%s):
%s

# REFINEMENT PRINCIPLES

**Priority:**
1. Natural flow - Sentences should flow beautifully
2. Idiomatic expressions - Use natural %s idioms
3. Elegant word choice - Select refined vocabulary
4. Rhythm and cadence - Pleasant reading rhythm
5. Preserve meaning - Keep original meaning intact

**What to Fix:**
- Awkward literal translations → Natural expressions
- Repetitive vocabulary → Rich, varied word choices
- Unnatural word order → Proper syntax

**What to Preserve:**
- All factual content and meaning
- Character names and proper nouns
- Technical terms (if any)

CRITICAL: If the draft is already good, return it unchanged.

Output ONLY the refined translation in %s. Do not include any explanation.`,
		targetLang,
		targetLang, targetLang,
		sourceLang, sourceText,
		targetLang, draftText,
		targetLang,
		targetLang,
	)
}

package arbiter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/valpere/peretran/internal/translator"
)

type OllamaArbiter struct {
	model   string
	baseURL string
	client  *http.Client
}

type OllamaRequest struct {
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
	Stream   bool   `json:"stream"`
	Format   string `json:"format"`
	Template string `json:"template"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func NewOllamaArbiter(model, baseURL string) *OllamaArbiter {
	return &OllamaArbiter{
		model:   model,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (a *OllamaArbiter) Evaluate(ctx context.Context, source string, sourceLang, targetLang string, results []translator.ServiceResult) (*EvaluationResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to evaluate")
	}

	if len(results) == 1 {
		return &EvaluationResult{
			SelectedService: results[0].ServiceName,
			CompositeText:   results[0].TranslatedText,
			IsComposite:     false,
			Reasoning:       "Only one service available",
		}, nil
	}

	prompt := buildArbiterPrompt(source, sourceLang, targetLang, results)

	reqBody := OllamaRequest{
		Model:  a.model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, fmt.Sprintf("%s/api/generate", a.baseURL), "POST", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("arbiter request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arbiter returned status %d", resp.StatusCode)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return parseArbiterResponse(ollamaResp.Response)
}

func buildArbiterPrompt(source, sourceLang, targetLang string, results []translator.ServiceResult) string {
	var sb strings.Builder
	sb.WriteString("You are a professional translator evaluator.\n")
	sb.WriteString(fmt.Sprintf("Given the original text in %s:\n", sourceLang))
	sb.WriteString(fmt.Sprintf(`"%s"`, source))
	sb.WriteString(fmt.Sprintf("\n\nAnd these translations to %s:\n", targetLang))

	for i, r := range results {
		sb.WriteString(fmt.Sprintf("  %d. [%s]: \"%s\"\n", i+1, r.ServiceName, r.TranslatedText))
	}

	sb.WriteString(`Select the best translation or compose an improved one from the available options.
Respond ONLY in JSON:
{
  "selected_service": "google|systran|ollama|composite",
  "final_text": "...",
  "reasoning": "..."
}
`)

	return sb.String()
}

func parseArbiterResponse(response string) (*EvaluationResult, error) {
	response = strings.TrimSpace(response)

	var parsed struct {
		SelectedService string `json:"selected_service"`
		FinalText       string `json:"final_text"`
		Reasoning       string `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(response), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse arbiter response as JSON: %w", err)
	}

	isComposite := parsed.SelectedService == "composite"

	return &EvaluationResult{
		SelectedService: parsed.SelectedService,
		CompositeText:   parsed.FinalText,
		IsComposite:     isComposite,
		Reasoning:       parsed.Reasoning,
	}, nil
}

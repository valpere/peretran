/*
Copyright © 2025 Valentyn Solomko <valentyn.solomko@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/valpere/peretran/internal"
	"github.com/valpere/peretran/internal/arbiter"
	"github.com/valpere/peretran/internal/chunker"
	"github.com/valpere/peretran/internal/detector"
	"github.com/valpere/peretran/internal/orchestrator"
	"github.com/valpere/peretran/internal/placeholder"
	"github.com/valpere/peretran/internal/refiner"
	"github.com/valpere/peretran/internal/store"
	"github.com/valpere/peretran/internal/translator"
)

var (
	inputFile   string
	outputFile  string
	sourceLang  string
	targetLang  string
	credentials string
	projectID   string

	services     []string
	useArbiter   bool
	arbiterModel string
	arbiterURL   string

	ollamaURL        string
	ollamaModels     []string
	openrouterKey    string
	openrouterModels []string

	systranKey    string
	mymemoryEmail string

	dbPath     string
	noCache    bool
	maxRetries int

	useRefine    bool
	refinerModel string
	refinerURL   string

	// Phase 6 flags
	fuzzyThreshold float64
	usePlaceholder bool
	chunkSize      int
	useGlossary    bool
)

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "Translate text using multiple services",
	Long: `Translate text using multiple translation services in parallel
and select the best result using an LLM arbiter.

Available services:
  - google       Google Translate (requires credentials)
  - systran     Systran Translate (requires API key)
  - mymemory    MyMemory (free, 5000 chars/day)
  - ollama      Ollama LLM (self-hosted)
  - openrouter  OpenRouter LLM (requires API key)

Use multiple services: --services google,ollama,openrouter

Two-pass translation:
  --refine      Enable Stage 2 literary refinement pass

Phase 6 options:
  --fuzzy-threshold  Fuzzy cache matching (0 to disable, e.g. 0.85)
  --placeholder      Protect HTML/Markdown markup during translation
  --chunk-size       Split large texts into chunks of N characters
  --glossary         Load terminology glossary from database`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if inputFile == outputFile {
			return fmt.Errorf("input file and output file cannot be the same")
		}

		strInp, err := os.ReadFile(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}

		text := string(strInp)

		ctx := context.Background()

		// Auto-detect source language when not specified.
		if sourceLang == "auto" {
			det := detector.New()
			if detected, ok := det.DetectISO(text); ok {
				sourceLang = detected
				fmt.Fprintf(os.Stderr, "Detected source language: %s\n", sourceLang)
			}
		}

		var db *store.Store
		if !noCache && dbPath != "" {
			db, err = store.New(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			// Exact cache check.
			if cached, found, cacheErr := db.GetCachedTranslation(ctx, text, sourceLang, targetLang); cacheErr == nil && found {
				fmt.Fprintf(os.Stderr, "Using cached translation\n")
				return writeOutput(outputFile, cached, sourceLang, targetLang, true)
			}

			// Fuzzy cache check.
			if fuzzyThreshold > 0 {
				if cached, found, cacheErr := db.FuzzyGetCachedTranslation(ctx, text, sourceLang, targetLang, fuzzyThreshold); cacheErr == nil && found {
					fmt.Fprintf(os.Stderr, "Using fuzzy-matched cached translation\n")
					return writeOutput(outputFile, cached, sourceLang, targetLang, true)
				}
			}
		}

		// Load glossary from DB.
		var glossaryTerms map[string]string
		if useGlossary && db != nil {
			glossaryTerms, err = db.GetGlossaryTerms(ctx, sourceLang, targetLang)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load glossary: %v\n", err)
			} else if len(glossaryTerms) > 0 {
				fmt.Fprintf(os.Stderr, "Loaded %d glossary terms\n", len(glossaryTerms))
			}
		}

		// Placeholder protection.
		var phMarkers []string
		phHint := ""
		if usePlaceholder {
			text, phMarkers = placeholder.Protect(text)
			if len(phMarkers) > 0 {
				phHint = placeholder.InstructionHint()
				fmt.Fprintf(os.Stderr, "Placeholder protection: %d markers\n", len(phMarkers))
			}
		}

		// Split into chunks if requested.
		chunks := chunker.Chunk(text, chunkSize)
		if len(chunks) > 1 {
			fmt.Fprintf(os.Stderr, "Splitting into %d chunks (max %d chars each)\n", len(chunks), chunkSize)
		}

		cfg := translator.ServiceConfig{
			Credentials: credentials,
			ProjectID:   projectID,
		}

		serviceList, err := buildServices(services, ollamaURL, openrouterKey, systranKey, mymemoryEmail, ollamaModels, openrouterModels)
		if err != nil {
			return err
		}

		orch := orchestrator.New(serviceList, orchestrator.OrchestratorConfig{
			Timeout:     30 * time.Second,
			MinServices: 1,
			MaxAttempts: maxRetries,
		})

		// Translate all chunks sequentially with sliding context.
		var translatedChunks []string
		previousContext := ""

		for i, chunk := range chunks {
			if len(chunks) > 1 {
				fmt.Fprintf(os.Stderr, "Translating chunk %d/%d...\n", i+1, len(chunks))
			}

			req := translator.TranslateRequest{
				Text:            chunk,
				SourceLang:      sourceLang,
				TargetLang:      targetLang,
				PreviousContext: previousContext,
				GlossaryTerms:   glossaryTerms,
				Instructions:    phHint,
			}

			// Stage 1: parallel translation.
			result := orch.Execute(ctx, cfg, req)
			if result.Succeeded == 0 {
				return fmt.Errorf("all translation services failed (chunk %d)", i+1)
			}

			var draftText string
			var selectedService string
			var isComposite bool
			var arbiterReasoning string

			if useArbiter && len(result.Results) > 1 {
				arb := arbiter.NewOllamaArbiter(arbiterModel, arbiterURL)
				evalResult, evalErr := arb.Evaluate(ctx, chunk, sourceLang, targetLang, result.Results)
				if evalErr != nil {
					fmt.Fprintf(os.Stderr, "Arbiter failed: %v, using first result\n", evalErr)
					draftText = result.Results[0].TranslatedText
					selectedService = result.Results[0].ServiceName
				} else {
					draftText = evalResult.CompositeText
					selectedService = evalResult.SelectedService
					isComposite = evalResult.IsComposite
					arbiterReasoning = evalResult.Reasoning
					fmt.Fprintf(os.Stderr, "Arbiter selected: %s\n", evalResult.SelectedService)
				}
			} else {
				draftText = result.Results[0].TranslatedText
				selectedService = result.Results[0].ServiceName
			}

			// Stage 2: optional literary refinement.
			chunkTranslation := draftText
			if useRefine {
				fmt.Fprintf(os.Stderr, "Running Stage 2 refinement (chunk %d)...\n", i+1)
				ref := refiner.NewOllamaRefiner(refinerModel, refinerURL)
				refined, refErr := ref.Refine(ctx, sourceLang, targetLang, chunk, draftText)
				if refErr != nil {
					fmt.Fprintf(os.Stderr, "Refiner failed: %v, using draft\n", refErr)
				} else {
					chunkTranslation = refined
				}
			}

			// Update sliding context for the next chunk.
			previousContext = chunker.ExtractContext(chunkTranslation, chunker.DefaultContextWords)

			translatedChunks = append(translatedChunks, chunkTranslation)

			// Persist chunk result to cache and DB.
			if db != nil && !noCache && len(chunks) == 1 {
				// Only save single-chunk translations to full-text cache.
				reqID := uuid.New().String()
				memReq := internal.TranslationRequest{
					ID:         reqID,
					SourceText: string(strInp),
					SourceLang: sourceLang,
					TargetLang: targetLang,
					Timestamp:  time.Now(),
				}
				_ = db.SaveRequest(ctx, memReq)
				for _, r := range result.Results {
					_ = db.SaveResult(ctx, reqID, r.ServiceName, r.TranslatedText, r.Confidence, int(r.Latency.Milliseconds()), r.Error)
				}
				_ = db.SaveFinalTranslation(ctx, reqID, selectedService, chunkTranslation, isComposite, arbiterReasoning)
				_ = db.SaveToMemory(ctx, string(strInp), sourceLang, targetLang, chunkTranslation, draftText, selectedService)
				if useRefine {
					_ = db.SaveToStage1Cache(ctx, string(strInp), sourceLang, targetLang, draftText, selectedService)
				}
			}
		}

		// Join chunk translations.
		finalText := strings.Join(translatedChunks, "\n\n")

		// Restore placeholders.
		if usePlaceholder && len(phMarkers) > 0 {
			finalText = placeholder.Restore(finalText, phMarkers)
			if missing := placeholder.Validate(finalText, phMarkers); len(missing) > 0 {
				fmt.Fprintf(os.Stderr, "Warning: %d placeholder(s) missing after translation: %v\n", len(missing), missing)
			}
		}

		return writeOutput(outputFile, finalText, sourceLang, targetLang, false)
	},
}

// writeOutput writes the translated text to outputFile and prints a summary.
func writeOutput(outputFile, text, sourceLang, targetLang string, fromCache bool) error {
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	if err := os.WriteFile(outputFile, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	if fromCache {
		fmt.Printf("Successfully translated %s to %s (from cache)\n", sourceLang, targetLang)
	} else {
		fmt.Printf("Successfully translated %s to %s\n", sourceLang, targetLang)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(translateCmd)

	translateCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file to translate (required)")
	translateCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for translation (required)")
	translateCmd.Flags().StringVarP(&sourceLang, "source", "s", "auto", "Source language code")
	translateCmd.Flags().StringVarP(&targetLang, "target", "t", "", "Target language code (required)")
	translateCmd.Flags().StringVarP(&credentials, "credentials", "c", "", "Path to Google Cloud credentials")
	translateCmd.Flags().StringVarP(&projectID, "project", "p", "", "Google Cloud Project ID")

	translateCmd.Flags().StringSliceVar(&services, "services", []string{"google"}, "Translation services to use (comma-separated)")
	translateCmd.Flags().BoolVar(&useArbiter, "arbiter", false, "Use LLM arbiter to select best translation")
	translateCmd.Flags().StringVar(&arbiterModel, "arbiter-model", "llama3.2", "Arbiter model name")
	translateCmd.Flags().StringVar(&arbiterURL, "arbiter-url", "http://localhost:11434", "Arbiter Ollama URL")

	translateCmd.Flags().BoolVar(&useRefine, "refine", false, "Enable Stage 2 literary refinement (two-pass translation)")
	translateCmd.Flags().StringVar(&refinerModel, "refiner-model", "llama3.2", "Refiner model name")
	translateCmd.Flags().StringVar(&refinerURL, "refiner-url", "http://localhost:11434", "Refiner Ollama URL")

	translateCmd.Flags().StringVar(&ollamaURL, "ollama-url", "http://localhost:11434", "Ollama base URL")
	translateCmd.Flags().StringSliceVar(&ollamaModels, "ollama-models", nil, "Ollama models to rotate (default list used if empty)")
	translateCmd.Flags().StringVar(&openrouterKey, "openrouter-key", "", "OpenRouter API key")
	translateCmd.Flags().StringSliceVar(&openrouterModels, "openrouter-models", nil, "OpenRouter models to rotate (default list used if empty)")
	translateCmd.Flags().StringVar(&systranKey, "systran-key", "", "Systran API key")
	translateCmd.Flags().StringVar(&mymemoryEmail, "mymemory-email", "", "MyMemory email (for higher limits)")

	translateCmd.Flags().StringVar(&dbPath, "db", "./data/peretran.db", "Database path for translation memory")
	translateCmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable translation memory cache")
	translateCmd.Flags().IntVar(&maxRetries, "max-retries", 3, "Total attempts per service including the first (1 = no retries)")

	// Phase 6 flags
	translateCmd.Flags().Float64Var(&fuzzyThreshold, "fuzzy-threshold", 0, "Fuzzy cache similarity threshold (0 to disable, e.g. 0.85)")
	translateCmd.Flags().BoolVar(&usePlaceholder, "placeholder", false, "Protect HTML/Markdown markup with placeholders during translation")
	translateCmd.Flags().IntVar(&chunkSize, "chunk-size", 0, "Split input into chunks of N characters (0 = no chunking)")
	translateCmd.Flags().BoolVar(&useGlossary, "glossary", false, "Load terminology glossary from database for LLM services")

	translateCmd.MarkFlagRequired("input")
	translateCmd.MarkFlagRequired("output")
	translateCmd.MarkFlagRequired("target")
}

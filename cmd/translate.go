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
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/valpere/peretran/internal"
	"github.com/valpere/peretran/internal/arbiter"
	"github.com/valpere/peretran/internal/detector"
	"github.com/valpere/peretran/internal/orchestrator"
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

	ollamaURL      string
	ollamaModels   []string
	openrouterKey  string
	openrouterModels []string

	systranKey    string
	mymemoryEmail string

	dbPath     string
	noCache    bool
	maxRetries int

	useRefine    bool
	refinerModel string
	refinerURL   string
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
  --refine      Enable Stage 2 literary refinement pass`,
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

		// Auto-detect source language when not specified
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

			if cached, found, cacheErr := db.GetCachedTranslation(ctx, text, sourceLang, targetLang); cacheErr == nil && found {
				fmt.Fprintf(os.Stderr, "Using cached translation\n")
				if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}
				if err := os.WriteFile(outputFile, []byte(cached), 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Printf("Successfully translated %s to %s (from cache)\n", sourceLang, targetLang)
				return nil
			}
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

		req := translator.TranslateRequest{
			Text:       text,
			SourceLang: sourceLang,
			TargetLang: targetLang,
		}

		// Stage 1: parallel translation
		result := orch.Execute(ctx, cfg, req)

		if result.Succeeded == 0 {
			return fmt.Errorf("all translation services failed")
		}

		var draftText string
		var selectedService string
		var isComposite bool
		var arbiterReasoning string

		if useArbiter && len(result.Results) > 1 {
			arb := arbiter.NewOllamaArbiter(arbiterModel, arbiterURL)
			evalResult, err := arb.Evaluate(ctx, text, sourceLang, targetLang, result.Results)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Arbiter failed: %v, using first result\n", err)
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

		// Stage 2: optional literary refinement pass
		finalText := draftText
		if useRefine {
			fmt.Fprintf(os.Stderr, "Running Stage 2 refinement...\n")
			ref := refiner.NewOllamaRefiner(refinerModel, refinerURL)
			refined, err := ref.Refine(ctx, sourceLang, targetLang, text, draftText)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Refiner failed: %v, using draft\n", err)
			} else {
				finalText = refined
				fmt.Fprintf(os.Stderr, "Refinement complete\n")
			}
		}

		if db != nil && !noCache {
			reqID := uuid.New().String()
			memReq := internal.TranslationRequest{
				ID:         reqID,
				SourceText: text,
				SourceLang: sourceLang,
				TargetLang: targetLang,
				Timestamp:  time.Now(),
			}
			_ = db.SaveRequest(ctx, memReq)

			for _, r := range result.Results {
				_ = db.SaveResult(ctx, reqID, r.ServiceName, r.TranslatedText, r.Confidence, int(r.Latency.Milliseconds()), r.Error)
			}

			_ = db.SaveFinalTranslation(ctx, reqID, selectedService, finalText, isComposite, arbiterReasoning)
			// Store both the final and draft (stage1) text in translation memory
			_ = db.SaveToMemory(ctx, text, sourceLang, targetLang, finalText, draftText, selectedService)
			if useRefine {
				_ = db.SaveToStage1Cache(ctx, text, sourceLang, targetLang, draftText, selectedService)
			}
		}

		if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := os.WriteFile(outputFile, []byte(finalText), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Printf("Successfully translated %s to %s\n", sourceLang, targetLang)
		fmt.Printf("Services used: %d/%d\n", result.Succeeded, len(services))
		if useRefine {
			fmt.Printf("Stage 2 refinement applied\n")
		}
		return nil
	},
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

	translateCmd.MarkFlagRequired("input")
	translateCmd.MarkFlagRequired("output")
	translateCmd.MarkFlagRequired("target")
}

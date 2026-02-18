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
	"github.com/valpere/peretran/internal/orchestrator"
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

	dbPath  string
	noCache bool

	csvColumn    []string
	csvDelimiter string
	csvComment   string
)

var translateCmd = &cobra.Command{
	Use:   "translate",
	Short: "Translate text using multiple services",
	Long: `Translate text using multiple translation services in parallel
and select the best result using an LLM arbiter.`,
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

		var db *store.Store
		if !noCache && dbPath != "" {
			db, err = store.New(dbPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			if cached, found, err := db.GetCachedTranslation(ctx, text, sourceLang, targetLang); err == nil && found {
				fmt.Fprintf(os.Stderr, "Using cached translation\n")
				text = cached
			}
		}

		cfg := translator.ServiceConfig{
			Credentials: credentials,
			ProjectID:   projectID,
			Timeout:     30 * time.Second,
		}

		var serviceList []translator.TranslationService

		for _, svcName := range services {
			switch svcName {
			case "google":
				serviceList = append(serviceList, translator.NewGoogleService())
			default:
				fmt.Fprintf(os.Stderr, "Unknown service: %s, skipping\n", svcName)
			}
		}

		if len(serviceList) == 0 {
			return fmt.Errorf("no valid services configured")
		}

		orch := orchestrator.New(serviceList, orchestrator.OrchestratorConfig{
			Timeout:     60 * time.Second,
			MinServices: 1,
		})

		req := translator.TranslateRequest{
			Text:       text,
			SourceLang: sourceLang,
			TargetLang: targetLang,
		}

		result := orch.Execute(ctx, cfg, req)

		if result.Succeeded == 0 {
			return fmt.Errorf("all translation services failed")
		}

		var finalText string

		if useArbiter && len(result.Results) > 1 {
			arb := arbiter.NewOllamaArbiter(arbiterModel, arbiterURL)
			evalResult, err := arb.Evaluate(ctx, text, sourceLang, targetLang, result.Results)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Arbiter failed: %v, using first result\n", err)
				finalText = result.Results[0].TranslatedText
			} else {
				finalText = evalResult.CompositeText
				fmt.Fprintf(os.Stderr, "Arbiter selected: %s\n", evalResult.SelectedService)
			}
		} else {
			finalText = result.Results[0].TranslatedText
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

			_ = db.SaveFinalTranslation(ctx, reqID, result.Results[0].ServiceName, finalText, false, "")
			_ = db.SaveToMemory(ctx, text, sourceLang, targetLang, finalText, result.Results[0].ServiceName)
		}

		if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := os.WriteFile(outputFile, []byte(finalText), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}

		fmt.Printf("Successfully translated %s to %s\n", sourceLang, targetLang)
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
	translateCmd.Flags().StringVar(&dbPath, "db", "./data/peretran.db", "Database path for translation memory")
	translateCmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable translation memory cache")

	translateCmd.MarkFlagRequired("input")
	translateCmd.MarkFlagRequired("output")
	translateCmd.MarkFlagRequired("target")
}

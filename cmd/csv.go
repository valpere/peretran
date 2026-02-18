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
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/valpere/peretran/internal/arbiter"
	"github.com/valpere/peretran/internal/detector"
	"github.com/valpere/peretran/internal/orchestrator"
	"github.com/valpere/peretran/internal/refiner"
	"github.com/valpere/peretran/internal/store"
	"github.com/valpere/peretran/internal/translator"
)

var (
	csvInputFile  string
	csvOutputFile string
	csvSourceLang string
	csvTargetLang string
	csvColumns    []int

	csvServices     []string
	csvUseArbiter   bool
	csvArbiterModel string
	csvArbiterURL   string

	csvOllamaURL        string
	csvOllamaModels     []string
	csvOpenrouterKey    string
	csvOpenrouterModels []string
	csvSystranKey       string
	csvMymemoryEmail    string

	csvUseRefine    bool
	csvRefinerModel string
	csvRefinerURL   string

	csvMaxRetries int
	csvDBPath     string
	csvNoCache    bool
	csvResume     string
)

var csvCmd = &cobra.Command{
	Use:   "csv",
	Short: "Translate columns of a CSV file",
	Long: `Translate one or more columns in a CSV file.

By default all columns are translated. Use -l to select specific columns
(0-indexed). The flag may be repeated to select multiple columns.

A checkpoint ID is printed at the start of each run. If the job is interrupted,
use --resume with that ID to skip already-translated cells.

Example:
  peretran translate csv -i data.csv -o out.csv -t uk -l 1 -l 3
  peretran translate csv -i data.csv -o out.csv -t uk --resume cp_123456789`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if csvInputFile == csvOutputFile {
			return fmt.Errorf("input file and output file cannot be the same")
		}

		f, err := os.Open(csvInputFile)
		if err != nil {
			return fmt.Errorf("failed to open input CSV: %w", err)
		}
		defer f.Close()

		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		if err != nil {
			return fmt.Errorf("failed to read CSV: %w", err)
		}

		if len(records) == 0 {
			return fmt.Errorf("CSV file is empty")
		}

		ctx := context.Background()

		srcLang := csvSourceLang
		if srcLang == "auto" && len(records) > 1 && len(records[1]) > 0 {
			det := detector.New()
			sample := records[1][0]
			if detected, ok := det.DetectISO(sample); ok {
				srcLang = detected
				fmt.Fprintf(os.Stderr, "Detected source language: %s\n", srcLang)
			}
		}

		// Open store for cache and checkpoint support.
		var db *store.Store
		if !csvNoCache && csvDBPath != "" {
			db, err = store.New(csvDBPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()
		}

		// Load or create checkpoint.
		var checkpointID string
		completedCells := make(map[string]string)

		if csvResume != "" {
			if db == nil {
				return fmt.Errorf("--resume requires --db to be set and --no-cache to be disabled")
			}
			if _, cpErr := db.GetCSVCheckpoint(ctx, csvResume); cpErr != nil {
				return fmt.Errorf("failed to load checkpoint: %w", cpErr)
			}
			checkpointID = csvResume
			cells, cpErr := db.GetCSVCells(ctx, checkpointID)
			if cpErr != nil {
				return fmt.Errorf("failed to load checkpoint cells: %w", cpErr)
			}
			completedCells = cells
			fmt.Fprintf(os.Stderr, "Resuming checkpoint %s (%d cells already done)\n", checkpointID, len(completedCells))
		} else if db != nil {
			checkpointID, err = db.CreateCSVCheckpoint(ctx, csvInputFile, csvOutputFile, srcLang, csvTargetLang)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create checkpoint: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Checkpoint ID: %s (use --resume %s to resume if interrupted)\n", checkpointID, checkpointID)
			}
		}

		serviceList, err := buildServices(csvServices, csvOllamaURL, csvOpenrouterKey, csvSystranKey, csvMymemoryEmail, csvOllamaModels, csvOpenrouterModels)
		if err != nil {
			return err
		}

		cfg := translator.ServiceConfig{}

		orch := orchestrator.New(serviceList, orchestrator.OrchestratorConfig{
			Timeout:     30 * time.Second,
			MinServices: 1,
			MaxAttempts: csvMaxRetries,
		})

		// Determine which columns to translate.
		colSet := make(map[int]bool, len(csvColumns))
		for _, c := range csvColumns {
			colSet[c] = true
		}
		translateAll := len(csvColumns) == 0

		// Build output records.
		out := make([][]string, len(records))
		for rowIdx, row := range records {
			out[rowIdx] = make([]string, len(row))
			copy(out[rowIdx], row)

			for colIdx, cell := range row {
				if !translateAll && !colSet[colIdx] {
					continue
				}
				if cell == "" {
					continue
				}

				cellKey := fmt.Sprintf("%d:%d", rowIdx, colIdx)

				// Use checkpoint data when resuming.
				if translated, done := completedCells[cellKey]; done {
					out[rowIdx][colIdx] = translated
					continue
				}

				// Check translation memory cache.
				if db != nil {
					if cached, found, cacheErr := db.GetCachedTranslation(ctx, cell, srcLang, csvTargetLang); cacheErr == nil && found {
						out[rowIdx][colIdx] = cached
						if checkpointID != "" {
							_ = db.SaveCSVCell(ctx, checkpointID, rowIdx, colIdx, cached)
						}
						continue
					}
				}

				req := translator.TranslateRequest{
					Text:       cell,
					SourceLang: srcLang,
					TargetLang: csvTargetLang,
				}

				result := orch.Execute(ctx, cfg, req)
				if result.Succeeded == 0 {
					fmt.Fprintf(os.Stderr, "Row %d col %d: all services failed, keeping original\n", rowIdx, colIdx)
					continue
				}

				translated := result.Results[0].TranslatedText

				if csvUseArbiter && len(result.Results) > 1 {
					arb := arbiter.NewOllamaArbiter(csvArbiterModel, csvArbiterURL)
					eval, arbErr := arb.Evaluate(ctx, cell, srcLang, csvTargetLang, result.Results)
					if arbErr != nil {
						fmt.Fprintf(os.Stderr, "Arbiter failed row %d col %d: %v\n", rowIdx, colIdx, arbErr)
					} else {
						translated = eval.CompositeText
					}
				}

				if csvUseRefine {
					ref := refiner.NewOllamaRefiner(csvRefinerModel, csvRefinerURL)
					refined, refErr := ref.Refine(ctx, srcLang, csvTargetLang, cell, translated)
					if refErr != nil {
						fmt.Fprintf(os.Stderr, "Refiner failed row %d col %d: %v\n", rowIdx, colIdx, refErr)
					} else {
						translated = refined
					}
				}

				out[rowIdx][colIdx] = translated

				// Persist to cache and checkpoint.
				if db != nil {
					draftText := result.Results[0].TranslatedText
					serviceUsed := result.Results[0].ServiceName
					_ = db.SaveToMemory(ctx, cell, srcLang, csvTargetLang, translated, draftText, serviceUsed)
					if checkpointID != "" {
						_ = db.SaveCSVCell(ctx, checkpointID, rowIdx, colIdx, translated)
					}
				}
			}
		}

		outFile, err := os.Create(csvOutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output CSV: %w", err)
		}
		defer outFile.Close()

		writer := csv.NewWriter(outFile)
		if err := writer.WriteAll(out); err != nil {
			return fmt.Errorf("failed to write output CSV: %w", err)
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return fmt.Errorf("failed to flush output CSV: %w", err)
		}

		// Mark checkpoint complete.
		if db != nil && checkpointID != "" {
			_ = db.CompleteCSVCheckpoint(ctx, checkpointID)
		}

		fmt.Printf("CSV translated successfully: %s\n", csvOutputFile)
		return nil
	},
}

func init() {
	translateCmd.AddCommand(csvCmd)

	csvCmd.Flags().StringVarP(&csvInputFile, "input", "i", "", "Input CSV file (required)")
	csvCmd.Flags().StringVarP(&csvOutputFile, "output", "o", "", "Output CSV file (required)")
	csvCmd.Flags().StringVarP(&csvSourceLang, "source", "s", "auto", "Source language code")
	csvCmd.Flags().StringVarP(&csvTargetLang, "target", "t", "", "Target language code (required)")
	csvCmd.Flags().IntSliceVarP(&csvColumns, "column", "l", nil, "Column index to translate (0-indexed, repeatable; default: all columns)")

	csvCmd.Flags().StringSliceVar(&csvServices, "services", []string{"google"}, "Translation services to use (comma-separated)")
	csvCmd.Flags().BoolVar(&csvUseArbiter, "arbiter", false, "Use LLM arbiter to select best translation")
	csvCmd.Flags().StringVar(&csvArbiterModel, "arbiter-model", "llama3.2", "Arbiter model name")
	csvCmd.Flags().StringVar(&csvArbiterURL, "arbiter-url", "http://localhost:11434", "Arbiter Ollama URL")

	csvCmd.Flags().BoolVar(&csvUseRefine, "refine", false, "Enable Stage 2 literary refinement")
	csvCmd.Flags().StringVar(&csvRefinerModel, "refiner-model", "llama3.2", "Refiner model name")
	csvCmd.Flags().StringVar(&csvRefinerURL, "refiner-url", "http://localhost:11434", "Refiner Ollama URL")

	csvCmd.Flags().StringVar(&csvOllamaURL, "ollama-url", "http://localhost:11434", "Ollama base URL")
	csvCmd.Flags().StringSliceVar(&csvOllamaModels, "ollama-models", nil, "Ollama models to rotate (default list used if empty)")
	csvCmd.Flags().StringVar(&csvOpenrouterKey, "openrouter-key", "", "OpenRouter API key")
	csvCmd.Flags().StringSliceVar(&csvOpenrouterModels, "openrouter-models", nil, "OpenRouter models to rotate (default list used if empty)")
	csvCmd.Flags().StringVar(&csvSystranKey, "systran-key", "", "Systran API key")
	csvCmd.Flags().StringVar(&csvMymemoryEmail, "mymemory-email", "", "MyMemory email (for higher limits)")
	csvCmd.Flags().IntVar(&csvMaxRetries, "max-retries", 3, "Total attempts per service including the first (1 = no retries)")

	csvCmd.Flags().StringVar(&csvDBPath, "db", "./data/peretran.db", "Database path for translation memory and checkpoints")
	csvCmd.Flags().BoolVar(&csvNoCache, "no-cache", false, "Disable translation memory cache and checkpoints")
	csvCmd.Flags().StringVar(&csvResume, "resume", "", "Resume from checkpoint ID (printed at start of original run)")

	csvCmd.MarkFlagRequired("input")
	csvCmd.MarkFlagRequired("output")
	csvCmd.MarkFlagRequired("target")
}

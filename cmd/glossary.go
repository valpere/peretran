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
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/valpere/peretran/internal/store"
)

var glossaryDBPath string

var glossaryCmd = &cobra.Command{
	Use:   "glossary",
	Short: "Manage the terminology glossary",
	Long: `Add, list, and delete terminology glossary entries.

Glossary entries ensure that specific source terms are always translated
to the same target term — useful for proper nouns, brand names, and
domain-specific vocabulary.`,
}

var (
	glossaryListSource string
	glossaryListTarget string
)

var glossaryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all glossary entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.New(glossaryDBPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Pass empty strings to list everything; flags narrow the filter.
		entries, err := db.ListGlossaryTerms(context.Background(), glossaryListSource, glossaryListTarget)
		if err != nil {
			return fmt.Errorf("failed to list glossary: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("Glossary is empty.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSOURCE LANG\tTARGET LANG\tSOURCE TERM\tTARGET TERM")
		for _, e := range entries {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				e.ID, e.SourceLang, e.TargetLang, e.SourceTerm, e.TargetTerm)
		}
		return w.Flush()
	},
}

var (
	glossaryAddSource string
	glossaryAddTarget string
)

var glossaryAddCmd = &cobra.Command{
	Use:   "add <source-term> <target-term>",
	Short: "Add or update a glossary entry",
	Long: `Add a glossary entry mapping a source-language term to a target-language term.

Example:
  peretran glossary add "Kyiv" "Київ" --source en --target uk`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if glossaryAddSource == "" {
			return fmt.Errorf("--source language flag is required")
		}
		if glossaryAddTarget == "" {
			return fmt.Errorf("--target language flag is required")
		}

		db, err := store.New(glossaryDBPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		if err := db.AddGlossaryTerm(context.Background(), glossaryAddSource, glossaryAddTarget, args[0], args[1]); err != nil {
			return fmt.Errorf("failed to add glossary entry: %w", err)
		}
		fmt.Printf("Added: [%s→%s] %q → %q\n", glossaryAddSource, glossaryAddTarget, args[0], args[1])
		return nil
	},
}

var glossaryDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a glossary entry by ID",
	Long: `Delete a glossary entry by its ID (shown in "peretran glossary list").

Example:
  peretran glossary delete gl_1234567890123456789`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.New(glossaryDBPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		if err := db.DeleteGlossaryTerm(context.Background(), args[0]); err != nil {
			return fmt.Errorf("failed to delete glossary entry: %w", err)
		}
		fmt.Printf("Deleted glossary entry: %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(glossaryCmd)

	glossaryCmd.PersistentFlags().StringVar(&glossaryDBPath, "db", "./data/peretran.db", "Database path")

	// --source / --target flags on the list subcommand for optional filtering.
	glossaryListCmd.Flags().StringVarP(&glossaryListSource, "source", "s", "", "Filter by source language code (e.g. en)")
	glossaryListCmd.Flags().StringVarP(&glossaryListTarget, "target", "t", "", "Filter by target language code (e.g. uk)")

	// --source / --target are required for add.
	glossaryAddCmd.Flags().StringVarP(&glossaryAddSource, "source", "s", "", "Source language code (e.g. en)")
	glossaryAddCmd.Flags().StringVarP(&glossaryAddTarget, "target", "t", "", "Target language code (e.g. uk)")

	glossaryCmd.AddCommand(glossaryListCmd)
	glossaryCmd.AddCommand(glossaryAddCmd)
	glossaryCmd.AddCommand(glossaryDeleteCmd)
}

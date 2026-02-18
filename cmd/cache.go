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

var cacheDBPath string

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the translation memory cache",
	Long:  `List, inspect, and clear the SQLite translation memory cache.`,
}

var cacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all translation memory entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.New(cacheDBPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		entries, err := db.ListMemory(context.Background())
		if err != nil {
			return fmt.Errorf("failed to list entries: %w", err)
		}

		if len(entries) == 0 {
			fmt.Println("No entries in translation memory.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSOURCE\tTARGET\tSERVICE\tUSED\tLAST USED\tINVALID\tTEXT")
		for _, e := range entries {
			snippet := e.SourceText
			if len(snippet) > 40 {
				snippet = snippet[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t%v\t%s\n",
				e.ID, e.SourceLang, e.TargetLang, e.ServiceUsed,
				e.UsageCount, e.LastUsed.Format("2006-01-02 15:04"),
				e.Invalidated, snippet)
		}
		return w.Flush()
	},
}

var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show translation memory statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.New(cacheDBPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		stats, err := db.Stats(context.Background())
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}

		fmt.Printf("Total entries:   %d\n", stats.TotalEntries)
		fmt.Printf("Active entries:  %d\n", stats.ActiveEntries)
		fmt.Printf("Invalid entries: %d\n", stats.InvalidEntries)
		fmt.Printf("Total usage:     %d\n", stats.TotalUsage)
		return nil
	},
}

var cacheDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a translation memory entry by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.New(cacheDBPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		if err := db.DeleteMemory(context.Background(), args[0]); err != nil {
			return fmt.Errorf("failed to delete entry: %w", err)
		}
		fmt.Printf("Deleted entry: %s\n", args[0])
		return nil
	},
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove all entries from translation memory",
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := store.New(cacheDBPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		n, err := db.ClearMemory(context.Background())
		if err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		fmt.Printf("Cleared %d entries from translation memory.\n", n)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)

	cacheCmd.PersistentFlags().StringVar(&cacheDBPath, "db", "./data/peretran.db", "Database path")

	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheDeleteCmd)
	cacheCmd.AddCommand(cacheClearCmd)
}

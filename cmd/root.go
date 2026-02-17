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
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Global variables to store command-line flags and configuration
var (
	cfgFile      string   // Path to configuration file
	inputFile    string   // Path to input file for translation
	outputFile   string   // Path where translated text will be saved
	sourceLang   string   // Source language code (e.g., 'en' for English)
	targetLang   string   // Target language code (e.g., 'es' for Spanish)
	projectID    string   // Google Cloud Project ID (required for Advanced API)
	credentials  string   // Path to Google Cloud credentials JSON file
	useAdvanced  bool     // Flag to switch between Basic and Advanced APIs
	csvColumn    []string // Column number to translate (for CSV files)
	csvDelimiter string   // Delimiter for CSV files
	csvComment   string   // Comment character for CSV files
	version      bool     // Print version of the application
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "peretran",
	Short: "CLI Google Translator written on Golang",
	Long: `A CLI application that translates text files using Google Translate API.
It supports both Basic and Advanced Google Translate APIs and various language options.
The Basic API is simpler but has fewer features, while the Advanced API offers more control but requires a Google Cloud Project ID.`,

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	// RunE is used instead of Run to allow error handling
	RunE: func(cmd *cobra.Command, args []string) error {
		if version {
			fmt.Println("peretran v0.1.0")
			return nil
		}

		if inputFile == outputFile {
			return fmt.Errorf("input file and output file are the same: %v", inputFile)
		}

		strInp, err := readInp(inputFile)
		if err != nil {
			return fmt.Errorf("failed to read input file: %v", err)
		}

		strOut, err := translateEx([]string{strInp}, useAdvanced)
		if err != nil {
			return fmt.Errorf("failed to translate text: %v", err)
		}

		// Ensure the output directory exists
		if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		return writeOut(outputFile, strOut)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.peretran.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Local flags (only available to this command)
	rootCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "Input file to translate (required)")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file for translation (required)")
	rootCmd.PersistentFlags().StringVarP(&sourceLang, "source", "s", "auto", "Source language code (e.g., 'en' for English)")
	rootCmd.PersistentFlags().StringVarP(&targetLang, "target", "t", "", "Target language code (e.g., 'uk' for Ukrainian) (required)")
	rootCmd.PersistentFlags().StringVarP(&projectID, "project", "p", "", "Google Cloud Project ID (required for advanced API)")
	rootCmd.PersistentFlags().StringVarP(&credentials, "credentials", "c", "", "Path to Google Cloud credentials JSON file")
	rootCmd.PersistentFlags().BoolVarP(&useAdvanced, "advanced", "a", false, "Use Advanced Google Translate API")
	rootCmd.Flags().BoolVarP(&version, "version", "v", false, "Print the version of the application")

	// Mark required flags
	// These flags must be provided or the application will show an error
	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("output")
	rootCmd.MarkFlagRequired("target")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gootrago" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".peretran")
	}

	viper.AutomaticEnv() // read in environment variables that match

	err := viper.ReadInConfig() // Find and read the config file
	if err == nil {
		// If a config file is found, read it in.
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else { // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

}

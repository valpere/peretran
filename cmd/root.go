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
	"os"

	"github.com/spf13/cobra"
)

var version = "0.2.0"

var rootCmd = &cobra.Command{
	Use:   "peretran",
	Short: "CLI Multi-Service Translator",
	Long: `A CLI application that translates text using multiple translation services in parallel
and selects the best result using an LLM arbiter.
	
Supported services: Google Translate, Systran, Ollama (LLM)
	
Use "peretran translate --help" for translation options.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

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

	"github.com/valpere/peretran/internal/translator"
)

var (
	defaultOllamaModels = []string{
		"gemma2:27b", "aya:35b", "mixtral:8x7b", "qwen3:14b",
		"gemma3:12b-it-qat", "phi4:14b-q4_K_M", "llama3.1:8b", "mistral:7b",
	}
	defaultOpenRouterModels = []string{
		"google/gemini-2.5-flash-preview:free",
		"qwen/qwen2.5-72b-instruct:free",
		"mistralai/mistral-nemo:free",
		"meta-llama/llama-3.1-8b-instruct:free",
	}
)

// buildServices constructs the list of translation services from CLI parameters.
// ollamaModels and openrouterModels may be nil to use the defaults.
func buildServices(serviceNames []string, ollamaBaseURL, openrouterAPIKey, systranAPIKey, mymemoryEmailAddr string, ollamaModels, openrouterModels []string) ([]translator.TranslationService, error) {
	if len(ollamaModels) == 0 {
		ollamaModels = defaultOllamaModels
	}
	if len(openrouterModels) == 0 {
		openrouterModels = defaultOpenRouterModels
	}

	var list []translator.TranslationService

	for _, name := range serviceNames {
		switch name {
		case "google":
			list = append(list, translator.NewGoogleService())
		case "systran":
			list = append(list, translator.NewSystranService(systranAPIKey))
		case "mymemory":
			list = append(list, translator.NewMyMemoryService(mymemoryEmailAddr))
		case "ollama":
			list = append(list, translator.NewOllamaTranslator(ollamaBaseURL, ollamaModels))
		case "openrouter":
			list = append(list, translator.NewOpenRouterService(openrouterAPIKey, "", openrouterModels))
		default:
			fmt.Fprintf(os.Stderr, "Unknown service: %s, skipping\n", name)
		}
	}

	if len(list) == 0 {
		return nil, fmt.Errorf("no valid services configured")
	}
	return list, nil
}

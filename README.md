# peretran

A CLI tool for translating text and CSV files using multiple translation services in parallel. It aggregates results and selects the best translation using an LLM arbiter, with an optional two-pass literary refinement stage.

## Features

- **Multi-service parallel translation** — Google Translate, Systran, MyMemory, Ollama (local LLM), OpenRouter (cloud LLM)
- **LLM arbiter** — optional LLM-based evaluation selects or composes the best result from all services
- **Two-pass refinement** — optional Stage 2 literary editor pass for higher-quality output
- **Translation memory** — SQLite cache for instant retrieval of repeated translations
- **Auto language detection** — detects source language automatically via [lingua-go](https://github.com/pemistahl/lingua-go)
- **CSV support** — translate selected columns or all columns in CSV files

## Installation

### Prerequisites

- Go 1.24+

### Build from Source

```bash
git clone https://github.com/valpere/peretran.git
cd peretran
go build -o peretran
```

### Verify

```bash
./peretran --version
```

## Quick Start

```bash
# Translate with Google Translate (requires GOOGLE_APPLICATION_CREDENTIALS)
./peretran translate -i input.txt -o output.txt -t uk

# Translate with Ollama (local LLM, no API key required)
./peretran translate -i input.txt -o output.txt -t uk --services ollama

# Translate with multiple services and LLM arbiter
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,ollama,openrouter \
  --openrouter-key sk-or-... \
  --arbiter

# Two-pass translation (parallel services + arbiter + literary refinement)
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,ollama \
  --arbiter --refine

# Translate a CSV file (all columns)
./peretran translate csv -i data.csv -o translated.csv -t uk

# Translate specific columns (0-indexed)
./peretran translate csv -i data.csv -o translated.csv -t uk -l 1 -l 3

# Manage translation memory
./peretran cache stats
./peretran cache list
./peretran cache clear
```

## Commands

### `peretran translate`

Translate a text file using one or more services in parallel.

```
Usage:
  peretran translate -i <input> -o <output> -t <lang> [flags]

Flags:
  -i, --input string             Input file to translate (required)
  -o, --output string            Output file for translation (required)
  -t, --target string            Target language code, e.g. uk, es, fr (required)
  -s, --source string            Source language code (default "auto")
  -c, --credentials string       Path to Google Cloud credentials JSON
  -p, --project string           Google Cloud Project ID

  --services strings             Services to use, comma-separated (default [google])
                                 Available: google, systran, mymemory, ollama, openrouter

  --arbiter                      Use LLM arbiter to select/compose best translation
  --arbiter-model string         Arbiter Ollama model (default "llama3.2")
  --arbiter-url string           Arbiter Ollama URL (default "http://localhost:11434")

  --refine                       Enable Stage 2 literary refinement (two-pass)
  --refiner-model string         Refiner Ollama model (default "llama3.2")
  --refiner-url string           Refiner Ollama URL (default "http://localhost:11434")

  --ollama-url string            Ollama base URL (default "http://localhost:11434")
  --ollama-models strings        Ollama models to rotate (uses default list if empty)
  --openrouter-key string        OpenRouter API key
  --openrouter-models strings    OpenRouter models to rotate (uses default list if empty)
  --systran-key string           Systran API key
  --mymemory-email string        MyMemory email for higher daily limits

  --db string                    SQLite database path (default "./data/peretran.db")
  --no-cache                     Disable translation memory cache
```

### `peretran translate csv`

Translate columns of a CSV file.

```
Usage:
  peretran translate csv -i <input.csv> -o <output.csv> -t <lang> [flags]

Flags:
  -i, --input string    Input CSV file (required)
  -o, --output string   Output CSV file (required)
  -t, --target string   Target language code (required)
  -s, --source string   Source language code (default "auto")
  -l, --column int      Column index to translate, 0-indexed (repeatable; default: all columns)

  All --services, --arbiter, --refine, --ollama-*, --openrouter-* flags apply
```

### `peretran cache`

Manage the SQLite translation memory.

```
peretran cache stats               # Show entry counts and total hits
peretran cache list                # List all entries
peretran cache delete <id>         # Delete one entry by ID
peretran cache clear               # Remove all entries
```

## Translation Services

| Service | Free | Requires |
|---------|------|----------|
| `google` | Paid | `GOOGLE_APPLICATION_CREDENTIALS` or `-c` flag |
| `systran` | Free tier | `--systran-key` |
| `mymemory` | 5000 chars/day | Nothing (or `--mymemory-email` for higher limits) |
| `ollama` | Free | Local Ollama instance running |
| `openrouter` | Free models available | `--openrouter-key` |
| `amazon` | — | Not implemented yet |
| `ibm` | — | Not implemented yet |

## How It Works

```
Input text
    │
    ▼
Check translation memory (SQLite cache)
    │
    ├── Hit → return cached result
    │
    └── Miss:
        ▼
        Stage 1: Parallel translation
        ┌──────────┬──────────┬──────────┐
        │ Google   │ Systran  │ Ollama   │ ...
        └──────────┴──────────┴──────────┘
            │
            ▼
        Arbiter (--arbiter, optional)
        LLM selects or composes best result
            │
            ▼
        Stage 2: Refinement (--refine, optional)
        Literary editor pass for natural fluency
            │
            ▼
        Save to cache → Write output file
```

## Configuration File

Optional `~/.peretran.yaml`:

```yaml
services:
  google:
    enabled: true
  ollama:
    enabled: true
    base_url: "http://localhost:11434"
    models:
      - llama3.2
      - gemma2:27b
  openrouter:
    enabled: false
    api_key: "${OPENROUTER_API_KEY}"

arbiter:
  enabled: true
  model: "llama3.2"

storage:
  database: "./data/peretran.db"
```

## Project Structure

```
peretran/
├── main.go
├── cmd/
│   ├── root.go          # CLI entry, version
│   ├── translate.go     # translate subcommand
│   ├── csv.go           # translate csv subcommand
│   ├── cache.go         # cache subcommand
│   └── common.go        # shared service builder
├── internal/
│   ├── types.go         # common types
│   ├── translator/      # service implementations
│   │   ├── service.go   # TranslationService interface
│   │   ├── google.go
│   │   ├── systran.go
│   │   ├── ollama.go
│   │   ├── openrouter.go
│   │   ├── mymemory.go
│   │   ├── amazon.go    # stub
│   │   └── ibm.go       # stub
│   ├── orchestrator/    # parallel execution
│   ├── arbiter/         # LLM evaluation
│   ├── refiner/         # Stage 2 literary refinement
│   ├── store/           # SQLite cache
│   ├── detector/        # language detection
│   └── markdown/        # markdown utilities
├── docs/
└── go.mod
```

## Language Codes

Use ISO 639-1 codes: `en`, `uk`, `es`, `fr`, `de`, `zh`, `ja`, `ko`, `pl`, `pt`, ...

Use `auto` to let peretran detect the source language automatically.

## Documentation

- [Installation Guide](docs/installation.md)
- [Usage Examples](docs/usage.md)
- [Configuration](docs/configuration.md)
- [CSV Translation](docs/csv-translation.md)
- [Quality Principles](docs/quality-principles.md)

## License

Apache License 2.0. See [LICENSE](LICENSE).

## Author

Valentyn Solomko — [valentyn.solomko@gmail.com](mailto:valentyn.solomko@gmail.com)

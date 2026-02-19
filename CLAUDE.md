# CLAUDE.md - Project-Specific Instructions

## Project Overview

- **Name**: peretran
- **Type**: CLI Multi-Service Translation Tool written in Go
- **Module**: `github.com/valpere/peretran`
- **Go Version**: 1.24+
- **License**: Apache 2.0

## Goals: Quality Translation Principles

This project implements proven approaches from TranslateBooksWithLLMs and book-translator for high-quality LLM translations:

### Core Principles

1. **Quality over Speed - Two-Pass Translation**
   - **Stage 1 (Primary Translation):** Translate text from source to target language using multiple parallel services
   - **Stage 2 (Refinement):** LLM literary editor reviews and improves the draft for fluency and idiomatic quality
   - Separates concerns: accuracy in stage 1, fluency in stage 2

2. **Adaptive Systems - Text Chunking with Sliding Context** *(Phase 6)*
   - Split large texts into chunks at paragraph/sentence/word boundaries (`--chunk-size`)
   - Pass last ~25 words of each chunk's translation to the next as `PreviousContext`
   - Maintains narrative continuity across chunk boundaries

3. **Structure Preservation - Placeholder System** *(Phase 6)*
   - For structured content (HTML, Markdown): protect tags with `[PHn]` placeholders (`--placeholder`)
   - Validate and restore after translation; warn on missing markers

4. **Terminology Glossary** *(Phase 6)*
   - Manage source→target term mappings via `peretran glossary list/add/delete`
   - Inject glossary into LLM prompts for consistent translation (`--glossary`)

5. **Engineering Rigor**
   - **Translation memory:** SQLite cache; instant retrieval of repeated translations
   - **Validation:** Verify translation is in target language (lingua-go)
   - **Retry with backoff:** Handle transient LLM failures
   - **Post-processing:** Remove LLM artifacts (thinking tags, echoes) in refiner

## Project Structure

```
peretran/
├── main.go
├── cmd/
│   ├── root.go          # CLI entry, version
│   ├── translate.go     # translate subcommand
│   ├── csv.go           # translate csv subcommand
│   ├── cache.go         # cache subcommand
│   └── common.go        # shared buildServices() helper + default model lists
├── internal/
│   ├── types.go         # common types (TranslationRequest)
│   ├── translator/      # translation service implementations
│   │   ├── types.go     # ServiceConfig, TranslateRequest, ServiceResult
│   │   ├── service.go   # TranslationService interface
│   │   ├── google.go
│   │   ├── systran.go
│   │   ├── ollama.go
│   │   ├── openrouter.go
│   │   ├── mymemory.go
│   │   ├── amazon.go    # stub: "Not Implemented Yet"
│   │   ├── ibm.go       # stub: "Not Implemented Yet"
│   │   └── doclingo.go  # stub: "Not Implemented Yet"
│   ├── orchestrator/    # parallel execution coordinator
│   ├── arbiter/         # LLM evaluation and selection
│   │   ├── arbiter.go   # Arbiter interface
│   │   └── ollama.go    # OllamaArbiter implementation
│   ├── refiner/         # Stage 2 literary refinement
│   │   ├── refiner.go   # Refiner interface
│   │   └── ollama.go    # OllamaRefiner implementation
│   ├── store/           # SQLite storage and caching
│   │   └── store.go     # Store with translation_memory, stage1_cache, glossary tables
│   ├── placeholder/     # HTML/Markdown tag protection (Phase 6)
│   │   └── placeholder.go  # Protect/Restore/Validate/InstructionHint
│   ├── chunker/         # Text chunking with sliding context (Phase 6)
│   │   └── chunker.go   # Chunk/ExtractContext
│   ├── detector/        # language detection (lingua-go)
│   └── markdown/        # markdown utilities (gomarkdown)
├── docs/
├── go.mod
└── go.sum
```

## Commands

### Build

```bash
go build -o peretran
# or
make build
```

### Run

```bash
# Basic translation (Google, default)
./peretran translate -i input.txt -o output.txt -t uk

# Multiple services
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,ollama,openrouter --openrouter-key sk-or-...

# With LLM arbiter
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,ollama --arbiter

# With two-pass refinement
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,ollama --arbiter --refine

# CSV translation - all columns
./peretran translate csv -i data.csv -o translated.csv -t uk

# CSV translation - specific columns (0-indexed)
./peretran translate csv -i data.csv -o translated.csv -t uk -l 1 -l 3

# Phase 6: fuzzy cache + placeholder protection + chunking + glossary
./peretran translate -i input.txt -o output.txt -t uk \
  --fuzzy-threshold 0.85 --placeholder --chunk-size 2000 --glossary

# Cache management
./peretran cache stats
./peretran cache list
./peretran cache delete <id>
./peretran cache clear

# Glossary management
./peretran glossary list
./peretran glossary add "Kyiv" "Київ" --source en --target uk
./peretran glossary delete <id>
```

### Test

```bash
go test ./...
```

### Lint

```bash
go vet ./...
```

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration
- `cloud.google.com/go/translate` - Google Cloud Translation API
- `golang.org/x/text` - Unicode normalization (NFC)
- `google.golang.org/api` - Google API client
- `modernc.org/sqlite` - Pure Go SQLite (no CGo)
- `github.com/google/uuid` - UUID generation
- `github.com/pemistahl/lingua-go` - Language detection
- `github.com/gomarkdown/markdown` - Markdown parsing

## Key Files

| File | Purpose |
|------|---------|
| `cmd/translate.go` | translate subcommand — parallel services, arbiter, refiner, cache |
| `cmd/csv.go` | translate csv subcommand — per-cell translation with column selection |
| `cmd/cache.go` | cache subcommand — list, stats, delete, clear |
| `cmd/glossary.go` | glossary subcommand — list, add, delete |
| `cmd/common.go` | `buildServices()` helper; default Ollama and OpenRouter model lists |
| `internal/translator/service.go` | `TranslationService` interface |
| `internal/translator/types.go` | `TranslateRequest` (Text, PreviousContext, GlossaryTerms, Instructions) |
| `internal/orchestrator/orchestrator.go` | parallel service execution |
| `internal/arbiter/ollama.go` | LLM-based translation evaluation/composition |
| `internal/refiner/ollama.go` | Stage 2 literary editor prompt and cleanup |
| `internal/store/store.go` | SQLite: `translation_memory`, `stage1_cache`, `glossary` tables |
| `internal/placeholder/placeholder.go` | Protect/Restore HTML and Markdown tags with `[PHn]` markers |
| `internal/chunker/chunker.go` | Chunk large texts; ExtractContext for sliding-window continuity |
| `internal/detector/detector.go` | source language auto-detection |

## Translation Flow

```
Input text
    │
    ▼
Auto-detect source language (if --source auto)
    │
    ▼
Check translation_memory (exact-match cache, NFC-normalized)
    ├── Hit  → write output, return
    └── Miss:
        ▼
        Stage 1: Orchestrator runs all services in parallel
        │
        ▼ (if --arbiter and >1 service succeeded)
        Arbiter: LLM selects or composes best result
        │
        ▼ (if --refine)
        Stage 2: Refiner LLM polishes for literary quality
        │
        ▼
        Save to translation_memory (+ stage1_cache if --refine)
        │
        ▼
        Write output file
```

## Implementation Status

| Phase | Feature | Status |
|-------|---------|--------|
| 1 | Foundation: types, interface, orchestrator, store, cache | ✅ Done |
| 2 | Services: Google, Systran, Ollama, OpenRouter, MyMemory | ✅ Done |
| 3 | Arbiter: LLM evaluation/composition (`--arbiter`) | ✅ Done |
| 4 | Two-pass refinement (`--refine`), stage1_cache | ✅ Done |
| 5a | Cache management CLI (`peretran cache`) | ✅ Done |
| 5b | CSV translation (`peretran translate csv`) | ✅ Done |
| 6 | Chunking, placeholder system, fuzzy matching, glossary | ✅ Done |

## Language Version

- Go 1.24+ (required)

## Code Style

- Follow Go best practices
- Exported functions have doc comments
- Error handling with descriptive messages
- Modular design per global CLAUDE.md principles

## Configuration

- Config file: `~/.peretran.yaml` (YAML, optional)
- Google credentials: via `-c` flag or `GOOGLE_APPLICATION_CREDENTIALS` env var
- Other API keys: via `--systran-key`, `--openrouter-key` flags or environment

## Special allowance

You feel free to read the `../context` directory.

## Author

Valentyn Solomko

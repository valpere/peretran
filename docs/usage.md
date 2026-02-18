# Usage Guide

Comprehensive examples for using peretran.

---

## Basic Translation

### Single service (Google Translate)

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
./peretran translate -i input.txt -o output.txt -t uk
```

### Single service (Ollama — no API key needed)

```bash
./peretran translate -i input.txt -o output.txt -t uk --services ollama
```

### Specify source language explicitly

```bash
./peretran translate -i input.txt -o output.txt -s en -t fr
```

Source language defaults to `auto`, which detects it via lingua-go.

---

## Multi-service Translation

Run several services in parallel and take the first successful result:

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,systran,mymemory
```

### With LLM arbiter

When multiple services succeed, the arbiter LLM selects or composes the best result:

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,ollama,openrouter \
  --openrouter-key sk-or-... \
  --arbiter
```

Use a different arbiter model:

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,ollama \
  --arbiter --arbiter-model gemma2:27b
```

---

## Two-Pass Translation (Stage 2 Refinement)

After the parallel stage (and optional arbiter), `--refine` runs a literary editor pass to improve fluency, idioms, and word choice:

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,ollama \
  --arbiter --refine
```

Use a different model for refinement than for arbitration:

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services ollama,openrouter \
  --arbiter --arbiter-model llama3.2 \
  --refine --refiner-model phi4:14b-q4_K_M
```

---

## Service-Specific Configuration

### Google Translate

```bash
# Via environment variable (recommended)
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
./peretran translate -i input.txt -o output.txt -t uk --services google

# Or via flag
./peretran translate -i input.txt -o output.txt -t uk \
  --services google -c /path/to/credentials.json -p your-project-id
```

### Systran

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services systran --systran-key YOUR_KEY
```

### MyMemory (free, no key)

```bash
./peretran translate -i input.txt -o output.txt -t uk --services mymemory

# Optional: provide email for higher daily limit
./peretran translate -i input.txt -o output.txt -t uk \
  --services mymemory --mymemory-email you@example.com
```

### Ollama (local LLM)

```bash
./peretran translate -i input.txt -o output.txt -t uk --services ollama

# Custom URL
./peretran translate -i input.txt -o output.txt -t uk \
  --services ollama --ollama-url http://192.168.1.10:11434

# Custom model list (randomly rotated)
./peretran translate -i input.txt -o output.txt -t uk \
  --services ollama \
  --ollama-models gemma2:27b,phi4:14b-q4_K_M,qwen3:14b
```

### OpenRouter (cloud LLMs including free models)

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services openrouter --openrouter-key sk-or-...

# Custom model list
./peretran translate -i input.txt -o output.txt -t uk \
  --services openrouter --openrouter-key sk-or-... \
  --openrouter-models "google/gemini-2.5-flash-preview:free,qwen/qwen2.5-72b-instruct:free"
```

---

## Translation Memory (Cache)

Translations are stored in SQLite and reused automatically on repeated input.

```bash
# Disable cache for this run
./peretran translate -i input.txt -o output.txt -t uk --no-cache

# Use a custom database location
./peretran translate -i input.txt -o output.txt -t uk \
  --db /var/lib/peretran/translations.db
```

---

## CSV Translation

### All columns

```bash
./peretran translate csv -i data.csv -o translated.csv -t uk
```

### Selected columns (0-indexed, repeatable)

```bash
# Translate columns 1 and 3 only
./peretran translate csv -i data.csv -o translated.csv -t uk -l 1 -l 3
```

### With multiple services and refinement

```bash
./peretran translate csv -i data.csv -o translated.csv -t uk \
  --services google,ollama \
  --arbiter --refine
```

---

## Cache Management

```bash
./peretran cache stats                  # Entry counts and total hits
./peretran cache list                   # Full list of cached translations
./peretran cache delete mem_1234567890  # Delete one entry by ID
./peretran cache clear                  # Remove all entries
```

Use a custom database path:

```bash
./peretran cache stats --db /var/lib/peretran/translations.db
```

---

## Batch Processing

Translate to multiple target languages:

```bash
#!/bin/bash
LANGS=(uk es fr de pl)
for lang in "${LANGS[@]}"; do
  ./peretran translate -i input.txt -o "output_${lang}.txt" -t "$lang" \
    --services google,ollama --arbiter
done
```

Translate multiple files:

```bash
for f in texts/*.txt; do
  name=$(basename "$f" .txt)
  ./peretran translate -i "$f" -o "translated/${name}_uk.txt" -t uk \
    --services ollama --no-cache
done
```

---

## Error Handling

```bash
# Input and output cannot be the same file
./peretran translate -i file.txt -o file.txt -t es
# Error: input file and output file cannot be the same

# All services failed
./peretran translate -i input.txt -o output.txt -t uk --services ollama
# Error: all translation services failed   (if Ollama is not running)
```

When `--arbiter` is set but the arbiter fails, peretran falls back to the first successful service result automatically.

When `--refine` is set but the refiner fails, peretran uses the draft (stage 1) result.

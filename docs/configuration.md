# Configuration Guide

## Precedence

Configuration values are applied in this order (highest priority first):

1. Command-line flags
2. Environment variables
3. Configuration file (`~/.peretran.yaml`)
4. Built-in defaults

---

## Configuration File

Optional. By default peretran looks for `~/.peretran.yaml`.

```yaml
# ~/.peretran.yaml

services:
  google:
    enabled: true
    # credentials set via GOOGLE_APPLICATION_CREDENTIALS env or -c flag
  systran:
    enabled: false
    api_key: "${SYSTRAN_API_KEY}"
  mymemory:
    enabled: false
  ollama:
    enabled: true
    base_url: "http://localhost:11434"
    models:
      - llama3.2
      - gemma2:27b
      - qwen3:14b
      - phi4:14b-q4_K_M
  openrouter:
    enabled: false
    api_key: "${OPENROUTER_API_KEY}"
    models:
      - google/gemini-2.5-flash-preview:free
      - qwen/qwen2.5-72b-instruct:free
      - mistralai/mistral-nemo:free
      - meta-llama/llama-3.1-8b-instruct:free

arbiter:
  enabled: false
  model: "llama3.2"
  base_url: "http://localhost:11434"

refiner:
  enabled: false
  model: "llama3.2"
  base_url: "http://localhost:11434"

storage:
  database: "./data/peretran.db"

cache:
  enabled: true
```

---

## Environment Variables

| Variable | Used by |
|----------|---------|
| `GOOGLE_APPLICATION_CREDENTIALS` | Google Translate service |
| `SYSTRAN_API_KEY` | Systran service |
| `OPENROUTER_API_KEY` | OpenRouter service |
| `OLLAMA_BASE_URL` | Ollama service |

---

## All CLI Flags

### `peretran translate`

| Flag | Default | Description |
|------|---------|-------------|
| `-i, --input` | required | Input file path |
| `-o, --output` | required | Output file path |
| `-t, --target` | required | Target language code (ISO 639-1) |
| `-s, --source` | `auto` | Source language code (or `auto` to detect) |
| `-c, --credentials` | — | Path to Google Cloud credentials JSON |
| `-p, --project` | — | Google Cloud Project ID |
| `--services` | `google` | Comma-separated service list |
| `--arbiter` | `false` | Enable LLM arbiter |
| `--arbiter-model` | `llama3.2` | Arbiter Ollama model |
| `--arbiter-url` | `http://localhost:11434` | Arbiter Ollama URL |
| `--refine` | `false` | Enable Stage 2 literary refinement |
| `--refiner-model` | `llama3.2` | Refiner Ollama model |
| `--refiner-url` | `http://localhost:11434` | Refiner Ollama URL |
| `--ollama-url` | `http://localhost:11434` | Ollama base URL |
| `--ollama-models` | *(built-in list)* | Ollama models to rotate |
| `--openrouter-key` | — | OpenRouter API key |
| `--openrouter-models` | *(built-in list)* | OpenRouter models to rotate |
| `--systran-key` | — | Systran API key |
| `--mymemory-email` | — | MyMemory email for higher limits |
| `--db` | `./data/peretran.db` | SQLite database path |
| `--no-cache` | `false` | Disable translation memory |

### `peretran translate csv`

Inherits all flags above, plus:

| Flag | Default | Description |
|------|---------|-------------|
| `-l, --column` | *(all)* | Column index to translate, 0-indexed (repeatable) |

### `peretran cache`

| Flag | Default | Description |
|------|---------|-------------|
| `--db` | `./data/peretran.db` | SQLite database path |

---

## Configuration Examples

### Local-only (Ollama, no cloud accounts)

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services ollama \
  --ollama-models "llama3.2,gemma2:27b"
```

### High-quality two-pass (multiple services + arbiter + refine)

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services google,ollama,openrouter \
  --openrouter-key sk-or-... \
  --arbiter --arbiter-model llama3.2 \
  --refine --refiner-model phi4:14b-q4_K_M
```

### Free services only

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services mymemory,ollama,openrouter \
  --openrouter-key sk-or-... \
  --arbiter
```

### Disable caching (for testing)

```bash
./peretran translate -i input.txt -o output.txt -t uk --no-cache
```

---

## Tips

- Keep API keys in environment variables rather than command-line flags to avoid leaking them in shell history.
- The `--ollama-models` and `--openrouter-models` lists are randomly rotated per translation to distribute load and vary quality.
- Use `--no-cache` when experimenting with different service configurations to avoid returning stale cached results.

# Installation Guide

## Prerequisites

- **Go 1.24+** — [golang.org/dl](https://golang.org/dl/)

No other mandatory dependencies. Additional services require service-specific credentials (see below).

---

## Build from Source

```bash
git clone https://github.com/valpere/peretran.git
cd peretran
go build -o peretran
./peretran --version
```

Or use the Makefile:

```bash
make build
```

---

## Translation Services Setup

peretran supports multiple translation services. You only need to configure the ones you intend to use.

### Google Translate (paid)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create or select a project
3. Enable **Cloud Translation API** under APIs & Services
4. Create a Service Account with the **Cloud Translate API User** role
5. Download the JSON key file

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"

# Or pass via flag:
./peretran translate -i input.txt -o output.txt -t uk \
  -c /path/to/credentials.json
```

### Systran (free tier available)

Sign up at [systransoft.com](https://www.systransoft.com/translation-products/translate-api/) to get an API key.

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services systran --systran-key YOUR_API_KEY
```

### MyMemory (free, no registration)

No configuration needed. An optional email increases the daily character limit.

```bash
./peretran translate -i input.txt -o output.txt -t uk --services mymemory
```

### Ollama (local LLM, free)

1. Install Ollama from [ollama.com](https://ollama.com/)
2. Pull one or more models:

```bash
ollama pull llama3.2
ollama pull gemma2:27b
ollama pull qwen3:14b
```

3. Ollama runs on `http://localhost:11434` by default.

```bash
./peretran translate -i input.txt -o output.txt -t uk --services ollama
```

Models are selected randomly from the configured list on each request. Default model list:

```
gemma2:27b, aya:35b, mixtral:8x7b, qwen3:14b,
gemma3:12b-it-qat, phi4:14b-q4_K_M, llama3.1:8b, mistral:7b
```

Override with `--ollama-models`:

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services ollama --ollama-models llama3.2,gemma2:27b
```

### OpenRouter (cloud LLMs, free models available)

Sign up at [openrouter.ai](https://openrouter.ai/) to get an API key. Free models require no credits.

```bash
./peretran translate -i input.txt -o output.txt -t uk \
  --services openrouter --openrouter-key sk-or-...
```

Default free model list:

```
google/gemini-2.5-flash-preview:free
qwen/qwen2.5-72b-instruct:free
mistralai/mistral-nemo:free
meta-llama/llama-3.1-8b-instruct:free
```

---

## Recommended First Translation

No cloud accounts needed — just Ollama:

```bash
# Install and start Ollama, pull a model
ollama pull llama3.2

# Run first translation
./peretran translate -i input.txt -o output.txt -t uk --services ollama
```

With Google Translate:

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
./peretran translate -i input.txt -o output.txt -t uk
```

---

## Troubleshooting

### "all translation services failed"

- **Google**: Check `GOOGLE_APPLICATION_CREDENTIALS` is set and the Translation API is enabled.
- **Ollama**: Verify Ollama is running (`ollama serve`) and the model is pulled.
- **OpenRouter**: Verify the API key is valid and you have selected models that exist.

### Language detection issues

Source language auto-detection runs on the full input text. For very short texts it may fail; use `-s en` (or the appropriate code) to specify explicitly.

### "failed to open database"

The default DB path is `./data/peretran.db`. Ensure the directory is writable, or specify a different path with `--db`:

```bash
./peretran translate -i input.txt -o output.txt -t uk --db /tmp/peretran.db
```

### Language code errors

Use ISO 639-1 codes: `en`, `uk`, `es`, `fr`, `de`, `zh`, `ja`, `ko`, `pl`, `pt`, `it`, `nl`, `ru`, ...

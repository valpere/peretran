# Usage Guide

This guide provides comprehensive examples for using peretran for various translation tasks.

## Basic Text Translation

### Simple Translation

Translate a text file from English to Spanish:

```bash
./peretran -i input.txt -o output.txt -t es
```

### Auto-detect Source Language

Let Google detect the source language automatically:

```bash
./peretran -i input.txt -o output.txt -t uk
```

### Specify Source Language

Explicitly specify the source language:

```bash
./peretran -i input.txt -o output.txt -s en -t fr
```

## Advanced API Usage

### Using Advanced API

The Advanced API (v3) provides more features but requires a Google Cloud Project ID:

```bash
./peretran -i input.txt -o output.txt -t es -a -p your-project-id -c /path/to/credentials.json
```

### Advanced API with All Options

```bash
./peretran \
  --input input.txt \
  --output output.txt \
  --source en \
  --target uk \
  --project your-project-id \
  --credentials /path/to/credentials.json \
  --advanced
```

## Common Examples

### Translate Multiple Languages

```bash
# To Spanish
./peretran -i article.txt -o article_es.txt -t es -c creds.json

# To French
./peretran -i article.txt -o article_fr.txt -t fr -c creds.json

# To German
./peretran -i article.txt -o article_de.txt -t de -c creds.json
```

### Batch Processing with Scripts

```bash
#!/bin/bash
# translate.sh

LANGS=("es" "fr" "de" "uk")
INPUT="input.txt"

for lang in "${LANGS[@]}"; do
  output="${INPUT%.*}_${lang}.txt"
  ./peretran -i "$INPUT" -o "$output" -t "$lang" -c creds.json
  echo "Translated to $lang: $output"
done
```

### Using with Environment Variable

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"

./peretran -i input.txt -o output.txt -t es
```

## Error Handling

### Input/Output Errors

```bash
# Error: input and output files cannot be the same
./peretran -i same.txt -o same.txt -t es
# Error: input file and output file are the same: same.txt
```

### Missing Required Flags

```bash
# Error: required flag(s) not set
./peretran -i input.txt -t es
# Error: required flag(s) not set: output
```

## API Differences

| Feature | Basic API | Advanced API |
|---------|-----------|---------------|
| Project ID Required | No | Yes |
| Authentication | Optional | Required |
| Use Case | Simple translations | Enterprise features |
| Cost | Standard pricing | Premium pricing |

## Tips

1. **Credentials**: Store credentials securely and never commit to version control
2. **Batch Processing**: For many files, use scripts to process them sequentially
3. **Language Detection**: Use `auto` for source language when input language is unknown
4. **CSV Files**: Use the `csv` subcommand for spreadsheet translations

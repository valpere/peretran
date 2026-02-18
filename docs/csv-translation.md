# CSV Translation Guide

## Basic CSV Translation

### Translate all columns

```bash
./peretran translate csv -i data.csv -o translated.csv -t uk
```

Input (`data.csv`):
```csv
Name,Description,Category
Apple,A sweet fruit,Fruit
Car,A vehicle,Vehicle
```

Output (`translated.csv`) — all columns translated to Ukrainian:
```csv
Назва,Опис,Категорія
Яблуко,Солодкий фрукт,Фрукт
Машина,Транспортний засіб,Транспортний засіб
```

---

## Column Selection

Use `-l` (repeatable) to select specific columns by 0-based index.

### Translate only columns 1 and 3

```bash
./peretran translate csv -i data.csv -o translated.csv -t uk -l 1 -l 3
```

Input:
```csv
ID,Name,Code,Description
1,Apple,FRU,A sweet fruit
2,Car,VEH,A vehicle
```

Output — only columns 1 and 3 translated, ID and Code preserved:
```csv
ID,Ім'я,Code,Опис
1,Яблуко,FRU,Солодкий фрукт
2,Машина,VEH,Транспортний засіб
```

---

## Multi-service and Refinement

All translation flags from `peretran translate` apply to `translate csv`:

```bash
# With arbiter
./peretran translate csv -i data.csv -o out.csv -t uk \
  --services google,ollama --arbiter -l 1

# With two-pass refinement
./peretran translate csv -i data.csv -o out.csv -t uk \
  --services ollama --refine -l 1 -l 2
```

---

## All Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-i, --input` | required | Input CSV file |
| `-o, --output` | required | Output CSV file |
| `-t, --target` | required | Target language code |
| `-s, --source` | `auto` | Source language code |
| `-l, --column` | *(all)* | Column index to translate, 0-indexed (repeatable) |
| `--services` | `google` | Translation services, comma-separated |
| `--arbiter` | `false` | Use LLM arbiter |
| `--refine` | `false` | Enable Stage 2 refinement |
| `--ollama-url` | `http://localhost:11434` | Ollama base URL |
| `--ollama-models` | *(built-in list)* | Ollama model rotation list |
| `--openrouter-key` | — | OpenRouter API key |
| `--openrouter-models` | *(built-in list)* | OpenRouter model rotation list |
| `--systran-key` | — | Systran API key |
| `--mymemory-email` | — | MyMemory email for higher limits |

---

## Notes

- **Column indices are 0-based**: first column = `0`, second = `1`, and so on.
- **Empty cells are skipped** and preserved as-is in the output.
- **All rows are translated** — there is no automatic header row detection. If your first row is a header you may either skip column 0 or post-process the output.
- **Large files** are loaded entirely into memory before translation begins.
- Each cell is translated independently; there is no cross-cell context.

---

## Use Cases

### Translate product descriptions (keep SKU and category)

```bash
./peretran translate csv -i products.csv -o products_uk.csv -t uk -l 2 -l 3
```

### Generate multi-language content

```bash
for lang in uk es fr de; do
  ./peretran translate csv -i content.csv -o "content_${lang}.csv" -t "$lang" -l 1
done
```

### Large files (split-translate-join)

```bash
# Split into chunks of 1000 lines
split -l 1000 large.csv chunk_

# Translate each chunk
for f in chunk_*; do
  ./peretran translate csv -i "$f" -o "tr_$f" -t uk --services ollama
done

# Recombine
cat tr_chunk_* > translated_large.csv
rm chunk_* tr_chunk_*
```

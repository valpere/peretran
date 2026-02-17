# CSV Translation Guide

This guide covers translating CSV (Comma-Separated Values) files using peretran.

## Basic CSV Translation

### Translate Entire CSV File

```bash
./peretran csv -i data.csv -o translated.csv -t uk
```

### Simple Example

Input (`data.csv`):
```csv
Name,Description,Category
Apple,A sweet fruit,Fruit
Car,A vehicle,Vehicle
```

Command:
```bash
./peretran csv -i data.csv -o translated.csv -t es
```

Output (`translated.csv`):
```csv
Name,Description,Category
Apple,Una fruta dulce,Fruit
Car,Un vehiculo,Vehicle
```

## Column Selection

### Translate Specific Columns

Translate only columns 1 and 3:

```bash
./peretran csv -i data.csv -o translated.csv -t uk -l 1 -l 3
```

### Using Letter Column Notation

You can also use Excel-style column letters:

```bash
./peretran csv -i data.csv -o translated.csv -t uk -l A -l C
```

## Delimiter Options

### Tab-Delimited Files (TSV)

```bash
./peretran csv -i data.tsv -o translated.tsv -t uk --csv-delimiter=$'\t'
```

### Pipe-Delimited Files

```bash
./peretran csv -i data.txt -o translated.txt -t uk --csv-delimiter="|"
```

### Semicolon-Delimited Files (European CSV)

```bash
./peretran csv -i data.csv -o translated.csv -t uk --csv-delimiter=";"
```

## Comment Handling

### Skip Comment Lines

```bash
./peretran csv -i data.csv -o translated.csv -t uk --csv-comment="#"
```

Input with comments:
```csv
# This is a comment
Name,Description
Apple,A fruit
```

## Advanced CSV Examples

### Multiple Columns with Tab Delimiter

```bash
./peretran csv \
  -i data.csv \
  -o translated.csv \
  -t uk \
  -l 2 -l 4 -l 5 \
  --csv-delimiter=$'\t'
```

### Complete Example

```bash
./peretran csv \
  -i spreadsheet.tsv \
  -o translated.tsv \
  -t uk \
  -l 1,3,5 \
  --csv-delimiter=$'\t' \
  --csv-comment="#" \
  -c /path/to/credentials.json
```

## CSV Command Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--column` | `-l` | Column numbers to translate | All columns |
| `--csv-delimiter` | | Delimiter character | `,` (comma) |
| `--csv-comment` | | Comment character | None |

## Column Numbering

- **Numeric**: `1`, `2`, `3` (1-indexed)
- **Letter**: `A`, `B`, `C` (Excel-style)
- **Multiple**: `-l 1 -l 3` or `-l 1,3`

## Use Cases

### Spreadsheet Translation

Translate product descriptions while keeping IDs and categories:

```bash
./peretran csv -i products.csv -o products_translated.csv -t es -l 2
```

### Multi-language Content

Maintain original language side-by-side:

```bash
# First translation
./peretran csv -i content.csv -o content_uk.csv -t uk -l 2

# Then edit to add original next to translation
```

### Large CSV Files

For very large files, consider processing in chunks:

```bash
# Split CSV
split -l 1000 large.csv chunk_

# Translate each chunk
for f in chunk_*; do
  ./peretran csv -i "$f" -o "translated_$f" -t uk
done

# Combine results
cat translated_chunk_* > final.csv
```

## Tips

1. **Headers**: The first row is treated as data, not headers
2. **Preservation**: Non-translated columns remain unchanged
3. **Memory**: Large files are loaded entirely into memory
4. **Delimiter**: Use `--csv-delimiter` for non-comma delimiters

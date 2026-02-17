# peretran

A CLI application for translating text files using Google Translate API. Written in Go, it supports both Basic and Advanced Google Cloud Translation APIs with configurable options for various file formats.

## Features

- **Dual API Support**: Use either Basic or Advanced Google Translate API
- **File Format Support**: Plain text files and CSV files with column selection
- **Language Detection**: Automatic source language detection or explicit specification
- **Flexible Configuration**: Command-line flags and YAML configuration file support
- **Custom Credentials**: Support for Google Cloud service account credentials

## Installation

### Prerequisites

- Go 1.24 or higher
- Google Cloud Platform account with Translation API enabled
- Google Cloud service account credentials (JSON key file)

### Build from Source

```bash
git clone https://github.com/valpere/peretran.git
cd peretran
go build -o peretran
```

## Quick Start

### Basic Translation

```bash
./peretran -i input.txt -o output.txt -t es
```

### CSV Translation

```bash
./peretran csv -i data.csv -o translated.csv -t uk
```

### With Advanced API

```bash
./peretran -i input.txt -o output.txt -t uk -c /path/to/credentials.json -a -p your-project-id
```

## Command-Line Options

| Flag | Short | Description | Required |
|------|-------|-------------|----------|
| `--input` | `-i` | Input file path | Yes |
| `--output` | `-o` | Output file path | Yes |
| `--target` | `-t` | Target language code (e.g., `uk`, `es`, `en`) | Yes |
| `--source` | `-s` | Source language code (default: `auto` for detection) | No |
| `--credentials` | `-c` | Path to Google Cloud credentials JSON file | No |
| `--advanced` | `-a` | Use Advanced Google Translate API | No |
| `--project` | `-p` | Google Cloud Project ID (required for Advanced API) | Conditional |
| `--config` | | Custom config file path | No |
| `--version` | `-v` | Print version information | No |

pecific Options

|### CSV-S Flag | Short | Description |
|------|-------|-------------|
| `--column` | `-l` | Column numbers to translate (e.g., `-l 1 -l 3` or `-l A`) |
| `--csv-delimiter` | | CSV delimiter character |
| `--csv-comment` | | CSV comment character |

## Configuration File

Create a `.peretran.yaml` file in your home directory:

```yaml
input: default-input.txt
output: default-output.txt
target: uk
source: auto
advanced: false
project: your-project-id
credentials: /path/to/credentials.json
```

## Project Structure

```
peretran/
├── main.go              # Application entry point
├── cmd/
│   ├── root.go          # Main command and flags
│   ├── csv.go           # CSV translation subcommand
│   ├── common.go        # File I/O utilities
│   └── translateEx.go   # Translation API implementations
├── docs/                # Documentation
├── go.mod               # Go module file
├── go.sum               # Dependencies
├── LICENSE              # Apache 2.0 License
└── README.md            # This file
```

## Documentation

- [Installation Guide](docs/installation.md)
- [Usage Examples](docs/usage.md)
- [Configuration](docs/configuration.md)
- [CSV Translation](docs/csv-translation.md)

## Language Codes

Use ISO 639-1 language codes. Examples:
- `en` - English
- `es` - Spanish
- `uk` - Ukrainian
- `fr` - French
- `de` - German
- `auto` - Auto-detect

## License

Licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.

## Author

Valentyn Solomko - [valentyn.solomko@gmail.com](mailto:valentyn.solomko@gmail.com)

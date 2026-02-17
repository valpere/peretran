# AGENTS.md - Project-Specific Instructions

## Project Overview

- **Name**: peretran
- **Type**: CLI Google Translator written in Go
- **Module**: `github.com/valpere/peretran`
- **Go Version**: 1.23.5+
- **License**: Apache 2.0

## Project Structure

```
peretran/
├── main.go              # Entry point - calls cmd.Execute()
├── cmd/
│   ├── root.go          # Main command with flags and translation logic
│   ├── csv.go           # CSV translation subcommand
│   ├── common.go        # File I/O utilities (read/write files, CSV)
│   └── translateEx.go  # Basic and Advanced Google Translate API
├── docs/                # Documentation
├── go.mod               # Go module
├── go.sum               # Dependencies
└── LICENSE
```

## Commands

### Build

```bash
go build -o peretran
```

### Run

```bash
# Basic translation
./peretran -i input.txt -o output.txt -t es

# CSV translation - translate specific columns
./peretran csv -i data.csv -o translated.csv -t uk -l 1 -l 3

# CSV translation - all columns
./peretran csv -i data.csv -o translated.csv -t uk

# Advanced API
./peretran -i input.txt -o output.txt -t uk -a -p project-id -c creds.json
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
- `cloud.google.com/go/translate` - Basic Translation API
- `cloud.google.com/go/translate/apiv3` - Advanced Translation API v3
- `golang.org/x/text/language` - Language parsing

## Key Files

| File | Purpose |
|------|---------|
| `cmd/root.go` | Main CLI entry, flags, file translation |
| `cmd/csv.go` | CSV subcommand with column selection |
| `cmd/common.go` | File I/O, CSV read/write utilities |
| `cmd/translateEx.go` | Basic/Advanced API implementations |

## Language Version

- Minimum: Go 1.23.5+ (from go.mod)
- Recommended: Go 1.24+ (from CLAUDE.md)

## Code Style

- Follow Go best practices
- Comprehensive comments on exported functions
- Error handling with descriptive messages
- Modular design per CLAUDE.md principles

## Configuration

- Config file: `~/.peretran.yaml` (YAML)
- Credentials: Via `-c` flag or `GOOGLE_APPLICATION_CREDENTIALS` env var

## Author

Valentyn Solomko - valentyn.solomko@gmail.com

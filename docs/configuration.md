# Configuration Guide

This guide explains how to configure peretran for your translation needs.

## Configuration File

peretran can read configuration from a YAML file. By default, it looks for `.peretran.yaml` in your home directory.

### Default Configuration File Location

- Linux/macOS: `$HOME/.peretran.yaml`
- Windows: `%USERPROFILE%\.peretran.yaml`

### Custom Configuration File

You can specify a custom config file using the `--config` flag:

```bash
./peretran --config /path/to/config.yaml -i input.txt -o output.txt -t es
```

## Configuration File Format

### Basic Configuration

```yaml
input: default-input.txt
output: default-output.txt
target: uk
```

### Full Configuration

```yaml
# File paths
input: input.txt
output: output.txt

# Language settings
source: auto
target: uk

# API settings
advanced: false
project: your-project-id
credentials: /path/to/credentials.json
```

## Configuration Options

| Option | Description | Required |
|--------|-------------|----------|
| `input` | Default input file path | No |
| `output` | Default output file path | No |
| `source` | Source language code (default: `auto`) | No |
| `target` | Default target language code | No |
| `advanced` | Use Advanced API (default: `false`) | No |
| `project` | Google Cloud Project ID | Conditional* |
| `credentials` | Path to credentials JSON file | No |

*Required when `advanced: true`

## Environment Variables

### Google Application Credentials

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
```

### Using Environment Variables

When credentials are provided via:
- Configuration file: `./peretran -c /path/to/creds.json ...`
- Environment variable: `export GOOGLE_APPLICATION_CREDENTIALS=/path/to/creds.json`

The application will use those credentials for authentication.

## Precedence

Configuration values are resolved in the following order (highest to lowest):

1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

Example:

```bash
# config.yaml has: target: fr
# This will override to Spanish
./peretran -i input.txt -o output.txt -t es
```

## Configuration Examples

### Development Setup

```yaml
input: test-input.txt
output: test-output.txt
target: es
source: en
advanced: false
credentials: ~/credentials/dev-creds.json
```

### Production Setup

```yaml
input: production-input.txt
output: production-output.txt
target: uk
source: auto
advanced: true
project: production-project-id
credentials: /secure/credentials/prod-creds.json
```

## CSV-Specific Configuration

For CSV translations, you can also specify CSV options in the config:

```yaml
input: data.csv
output: translated.csv
target: uk

# CSV options
csvColumn: "1,3"
csvDelimiter: ","
csvComment: "#"
```

Note: CSV-specific options are typically better specified via command-line flags when using the `csv` subcommand.

## Tips

1. **Security**: Never commit credentials files to version control
2. **Organization**: Use separate config files for different environments
3. **Defaults**: Set common options in the config file for convenience
4. **Overrides**: Use command-line flags to override config values as needed

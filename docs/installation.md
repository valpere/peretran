# Installation Guide

This guide covers setting up peretran for translation tasks.

## Prerequisites

- **Go 1.24+** - Download from [golang.org](https://golang.org/dl/)
- **Google Cloud Platform Account** - Sign up at [cloud.google.com](https://cloud.google.com/)
- **Translation API Enabled** - Enable Cloud Translation API in Google Cloud Console
- **Service Account Credentials** - Create a service account with Translation API permissions

## Google Cloud Setup

### 1. Enable Translation API

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select or create a new project
3. Navigate to APIs & Services > Library
4. Search for "Cloud Translation API" and enable it

### 2. Create Service Account

1. Go to IAM & Admin > Service Accounts
2. Click "Create Service Account"
3. Enter a name and description
4. Grant the "Cloud Translate API User" role
5. Create a JSON key file and save it securely

## Install from Source

### Clone the Repository

```bash
git clone https://github.com/valpere/peretran.git
cd peretran
```

### Build the Application

```bash
go build -o peretran
```

### Verify Installation

```bash
./peretran --version
```

Output:
```
peretran v0.1.0
```

## First Translation

### Basic Translation

```bash
./peretran -i input.txt -o output.txt -t es -c /path/to/credentials.json
```

### Translation with Auto-detection

```bash
./peretran -i input.txt -o output.txt -t uk -c /path/to/credentials.json
```

## Environment Variables

You can also set credentials via environment variable:

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
./peretran -i input.txt -o output.txt -t es
```

## Troubleshooting

### Authentication Errors

- Verify credentials file path is correct
- Ensure the service account has Cloud Translate API User role
- Check that the Translation API is enabled in your project

### Permission Denied Errors

If using Advanced API:
```bash
Error: failed to translate text: rpc error: code = PermissionDenied desc = Cloud IAM permission 'cloudtranslate.generalModels.predict' denied.
```

Ensure your service account has the correct permissions for the Advanced API.

### Language Code Errors

Use valid ISO 639-1 language codes:
- `en` - English
- `es` - Spanish
- `uk` - Ukrainian
- `fr` - French
- `de` - German
- `zh` - Chinese
- `ja` - Japanese
- `ko` - Korean

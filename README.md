# LogLion

LogLion is a Go-based CLI tool that analyzes logcat files to validate analytics event funnels for automated testing.

## Overview

LogLion helps you track user conversion funnels by parsing logcat files and checking if users complete expected
sequences of analytics events. This is particularly useful for automated testing of mobile applications where you need
to validate that analytics events are being fired correctly throughout user journeys.

## Features

- **Logcat parsing**: Parse logcat files with support for plain text and JSON formats
- **Funnel analysis**: Track multi-step user conversion funnels
- **Flexible configuration**: YAML-based configuration for defining funnels and steps
- **Multiple output formats**: Text and JSON output formats
- **Pattern matching**: Regex-based event pattern matching with property validation

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd loglion

# Build the binary
go build -o loglion

# Or install directly
go install
```

## Quick Start

1. **Create a funnel configuration file** (`funnel.yaml`):

```yaml
version: "1.0"
format: "logcat-plain"

funnel:
  name: "Purchase Flow"

  steps:
    - name: "Product View"
      event_pattern: "analytics.*page_view"
      required_properties:
        page: "/product"

    - name: "Add to Cart"
      event_pattern: "analytics.*add_to_cart"
      required_properties:
        product_id: ".*"

    - name: "Purchase"
      event_pattern: "analytics.*purchase"
      required_properties:
        transaction_id: ".*"
```

2. **Analyze your log file**:

```bash
loglion analyze --config funnel.yaml --log logcat.txt
```

## Commands

### `analyze`

Analyze log files for funnel validation.

```bash
loglion analyze --config funnel.yaml --log logcat.txt [flags]
```

**Flags:**

- `--config, -c`: Path to funnel configuration file (required)
- `--log, -l`: Path to log file (required)
- `--format, -f`: Log format preset (default: "logcat-plain")
- `--output, -o`: Output format (json, text) (default: "text")
- `--timeout, -t`: Session timeout in minutes (default: 30)

### `validate`

Validate funnel configuration file.

```bash
loglion validate --config funnel.yaml
```

### `version`

Show version information.

```bash
loglion version
```

## Configuration

The configuration file defines how LogLion should parse logs and what constitutes a successful funnel completion.

### Basic Structure

```yaml
version: "1.0"           # Configuration version
format: "logcat-plain"        # Log format preset

funnel:
  name: "My Funnel"                    # Descriptive name

  steps: # Funnel steps (in order)
    - name: "Step 1"
      event_pattern: "regex_pattern"   # Regex to match events
      required_properties: # Optional property validation
        key: "value_pattern"
```

### Log Parser Configuration

```yaml
log_parser:
  timestamp_format: "01-02 15:04:05.000"    # Timestamp parsing format
  event_regex: ".*Analytics.*: (.*)"         # Regex to extract event data
  json_extraction: true                      # Parse JSON from log lines
```

## Examples

See the `examples/` directory for:

- `funnel.yaml`: Example funnel configuration
- `sample_logcat.txt`: Sample Android logcat file

## License

Copyright 2025 Vladimir Parfenov

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
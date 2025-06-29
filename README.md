# LogLion

LogLion is a Go-based CLI tool that analyzes log files to validate analytics event funnels for automated testing.

## Overview

LogLion helps you track user conversion funnels by parsing various log file formats and checking if users complete expected
sequences of analytics events. This is particularly useful for automated testing of applications where you need
to validate that analytics events are being fired correctly throughout user journeys.

## Features

- **Flexible log parsing**: Parse plain text log files with configurable formats
- **Funnel analysis**: Track multi-step user conversion funnels
- **Flexible configuration**: YAML-based configuration for defining funnels and steps
- **Multiple output formats**: Text and JSON output formats
- **Pattern matching**: Regex-based event pattern matching with property validation
- **Configurable parsing**: Support for custom log formats through configuration without code changes

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
funnel:
  name: "Purchase Flow"

  steps:
    - name: "Product View"
      event_pattern: "page_view"

    - name: "Add to Cart"
      event_pattern: "add_to_cart"

    - name: "Purchase"
      event_pattern: "purchase"

# Configure parser for logcat-style logs
log_parser:
  timestamp_format: "01-02 15:04:05.000"
  event_regex: ".*Analytics: (.*)"
  json_extraction: true
  log_line_regex: "^(\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}\\.\\d{3})\\s+(\\d+)\\s+(\\d+)\\s+([VDIWEFS])\\s+([^:]+):\\s*(.*)$"
```

2. **Analyze your log file**:

```bash
loglion funnel --config funnel.yaml --log logcat.txt
```

## Commands

### `funnel`

Analyze log files for funnel validation.

```bash
loglion funnel --config funnel.yaml --log logcat.txt [flags]
```

**Flags:**

- `--config, -c`: Path to funnel configuration file (required)
- `--log, -l`: Path to log file (required)
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
funnel:
  name: "My Funnel"                    # Descriptive name

  steps: # Funnel steps (in order)
    - name: "Step 1"
      event_pattern: "regex_pattern"   # Regex to match events
      required_properties: # Optional property validation
        key: "value_pattern"
```

### Log Parser Configuration

LogLion supports flexible configuration for different log types:

```yaml
log_parser:
  timestamp_format: "01-02 15:04:05.000"     # Go time format (empty = no timestamp parsing)
  event_regex: ".*Analytics: (.*)"           # Regex to extract event data from message
  json_extraction: true                      # Parse JSON from extracted event data
  log_line_regex: "^(.*)$"                   # Regex to parse the entire log line (default: match everything)

# Example configurations:

# Simple text logs (just event names per line):
log_parser:
  event_regex: "^(.*)$"

# Logcat format:
log_parser:
  timestamp_format: "01-02 15:04:05.000"
  event_regex: ".*Analytics: (.*)"
  json_extraction: true
  log_line_regex: "^(\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}\\.\\d{3})\\s+(\\d+)\\s+(\\d+)\\s+([VDIWEFS])\\s+([^:]+):\\s*(.*)$"

# OSLog format:
log_parser:
  timestamp_format: "2006-01-02 15:04:05.000000-0700"
  event_regex: "Analytics: (.*)"
  json_extraction: true
  log_line_regex: "^(\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}\\.\\d{6}-\\d{4})\\s+(.*)$"
```

## Examples

See the `examples/` directory for:

- `simple_funnel.yaml`: Simple text log configuration
- `plain_funnel.yaml`: Plain text log configuration
- `oslog_funnel.yaml`: macOS oslog format configuration
- `sample_simple.txt`: Simple text log sample
- `sample_logcat_plain.txt`: Sample logcat file

## License

```
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
```
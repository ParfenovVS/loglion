# LogLion

LogLion is a Go-based CLI tool that analyzes log files to validate analytics event funnels for automated testing.

## Overview

LogLion helps you track user conversion funnels by parsing various log file formats and checking if users complete expected
sequences of analytics events. This is particularly useful for automated testing of applications where you need
to validate that analytics events are being fired correctly throughout user journeys.

## Features

- **Flexible log parsing**: Parse plain text log files with configurable formats
- **Funnel analysis**: Track multi-step user conversion funnels
- **Separate configurations**: Independent parser and funnel configurations for reusability
- **Multiple output formats**: Text and JSON output formats
- **Pattern matching**: Regex-based event pattern matching with property validation
- **JSON Schema validation**: Auto-completion and validation in IDEs through JSON schemas

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

1. **Create separate configuration files**:

**Parser Configuration** (`parser.yaml`):
```yaml
timestamp_format: "01-02 15:04:05.000"
event_regex: ".*Analytics: (.*)"
json_extraction: true
log_line_regex: "^(\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}\\.\\d{3})\\s+(\\d+)\\s+(\\d+)\\s+([VDIWEFS])\\s+([^:]+):\\s*(.*)$"
```

**Funnel Configuration** (`funnel.yaml`):
```yaml
name: "Purchase Flow"
steps:
  - name: "Product View"
    event_pattern: "page_view"
  - name: "Add to Cart"
    event_pattern: "add_to_cart"
  - name: "Purchase"
    event_pattern: "purchase"
```

2. **Analyze your log file**:

```bash
loglion funnel --parser-config parser.yaml --funnel-config funnel.yaml --log log.txt
```

## Commands

### `funnel`

Analyze log files for funnel validation.

```bash
loglion funnel --parser-config parser.yaml --funnel-config funnel.yaml --log log.txt [flags]
```

**Flags:**

- `--parser-config, -p`: Path to parser configuration file (required)
- `--funnel-config, -f`: Path to funnel configuration file (required)
- `--log, -l`: Path to log file (required)
- `--output, -o`: Output format (json, text) (default: "text")
- `--max`: Limit analysis (stop after N complete sequences) (default: 0 = analyze all)

### `validate`

Validate configuration files.

```bash
# Validate parser configuration
loglion validate --parser-config parser.yaml

# Validate funnel configuration  
loglion validate --funnel-config funnel.yaml

# Validate both configurations
loglion validate --parser-config parser.yaml --funnel-config funnel.yaml
```

### `version`

Show version information.

```bash
loglion version
```

## Configuration

LogLion uses separate configuration files for parser and funnel settings, allowing for better reusability and separation of concerns.

### Parser Configuration

Defines how to parse log files. Reusable across multiple funnels.

```yaml
timestamp_format: "01-02 15:04:05.000"     # Go time format (empty = no timestamp parsing)
event_regex: ".*Analytics: (.*)"           # Regex to extract event data from message
json_extraction: true                      # Parse JSON from extracted event data
log_line_regex: "^(.*)$"                   # Regex to parse the entire log line (default: match everything)
```

### Funnel Configuration

Defines the sequence of steps to track. Independent of log format.

```yaml
name: "My Funnel"                    # Descriptive name
steps: # Funnel steps (in order)
  - name: "Step 1"
    event_pattern: "regex_pattern"   # Regex to match events
    required_properties: # Optional property validation
      key: "value_pattern"
```

### Example Configurations

**Simple text logs** (`simple-parser.yaml`):
```yaml
event_regex: "^(.*)$"
json_extraction: false
```

**Logcat format** (`logcat-parser.yaml`):
```yaml
timestamp_format: "01-02 15:04:05.000"
event_regex: ".*Analytics: (.*)"
json_extraction: true
log_line_regex: "^(\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}\\.\\d{3})\\s+(\\d+)\\s+(\\d+)\\s+([VDIWEFS])\\s+([^:]+):\\s*(.*)$"
```

**OSLog format** (`oslog-parser.yaml`):
```yaml
timestamp_format: "2006-01-02 15:04:05.000000-0700"
event_regex: "Analytics: (.*)"
json_extraction: true
log_line_regex: "^(\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}\\.\\d{6}-\\d{4})\\s+(.*)$"
```

## Usage Examples

```bash
# Simple text logs
loglion funnel -p examples/simple-parser.yaml -f examples/simple-funnel.yaml -l examples/sample_simple.txt

# Logcat-style logs
loglion funnel -p examples/logcat-parser.yaml -f examples/purchase-funnel.yaml -l examples/sample_logcat_plain.txt

# OSLog-style logs  
loglion funnel -p examples/oslog-parser.yaml -f examples/simple-funnel.yaml -l oslog.txt

# JSON output for automation
loglion funnel -p examples/logcat-parser.yaml -f examples/purchase-funnel.yaml -l app.log --output json

# Limit analysis (stop after 5 complete sequences)
loglion funnel -p examples/simple-parser.yaml -f examples/simple-funnel.yaml -l app.log --max 5
```

## Examples

See the `examples/` directory for:

- `simple-parser.yaml`: Simple text log parser configuration
- `logcat-parser.yaml`: Android logcat parser configuration  
- `oslog-parser.yaml`: macOS oslog parser configuration
- `simple-funnel.yaml`: Basic funnel configuration
- `purchase-funnel.yaml`: E-commerce purchase funnel
- `sample_simple.txt`: Simple text log sample
- `sample_logcat_plain.txt`: Sample logcat file

## JSON Schema Support

LogLion includes JSON schemas that provide:
- IDE auto-completion for YAML configuration files
- Real-time validation while editing
- Documentation of all configuration options

Schemas are located in the `schema/` directory:
- `parser-config.schema.json`: Parser configuration schema
- `funnel-config.schema.json`: Funnel configuration schema

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
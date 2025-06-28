# LogLion

LogLion is a Go-based CLI tool that analyzes ADB logcat logs to validate analytics event funnels for automated testing.

## Overview

LogLion helps you track user conversion funnels by parsing Android log files and checking if users complete expected sequences of analytics events. This is particularly useful for automated testing of mobile applications where you need to validate that analytics events are being fired correctly throughout user journeys.

## Features

- **Android logcat parsing**: Parse ADB logcat logs with built-in Android format support
- **Funnel analysis**: Track multi-step user conversion funnels
- **Session management**: Group events by user/session with configurable timeouts
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
format: "android"

funnel:
  name: "Purchase Flow"
  session_key: "user_id"
  timeout_minutes: 30
  
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

3. **View results**:

```
✅ Funnel Analysis Complete

Funnel: Purchase Flow
Total Sessions: 5
Completed Funnels: 3 (60%)

Step Breakdown:
1. Product View: 5/5 (100%)
2. Add to Cart: 4/5 (80%)
3. Purchase: 3/5 (60%)

Drop-off Analysis:
- Product View → Add to Cart: 1 session lost
- Add to Cart → Purchase: 1 session lost
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
- `--format, -f`: Log format preset (default: "android")
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
format: "android"        # Log format preset

funnel:
  name: "My Funnel"                    # Descriptive name
  session_key: "user_id"              # Field to group events by user
  timeout_minutes: 30                 # Session timeout
  
  steps:                               # Funnel steps (in order)
    - name: "Step 1"
      event_pattern: "regex_pattern"   # Regex to match events
      required_properties:             # Optional property validation
        key: "value_pattern"
```

### Android Parser Configuration

```yaml
android_parser:
  timestamp_format: "01-02 15:04:05.000"    # Timestamp parsing format
  event_regex: ".*Analytics.*: (.*)"         # Regex to extract event data
  json_extraction: true                      # Parse JSON from log lines
```

## Examples

See the `examples/` directory for:
- `funnel.yaml`: Example funnel configuration
- `sample_logcat.txt`: Sample Android logcat file

## Development Status

This is an MVP (Minimum Viable Product) focused on Android logcat analysis. The project is structured to allow easy extension and enhancement.

### Current Implementation Status

- [x] Project structure and CLI framework
- [x] Configuration file parsing and validation
- [x] Basic Android logcat parser
- [x] Funnel analysis engine
- [x] Text and JSON output formatters
- [ ] Complete integration and testing
- [ ] Error handling improvements
- [ ] Performance optimizations

### Planned Enhancements

- iOS log format support
- Custom log format definitions
- Multiple funnel analysis in single run
- Visual HTML reports
- Real-time log monitoring
- CI/CD integration helpers

## Contributing

This project is designed to be developed iteratively by AI agents. Each component is modular and can be enhanced independently.

## License

[Add your license here]
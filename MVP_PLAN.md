# LogLion - MVP Plan

## Project Overview
LogLion is a Go-based CLI tool that analyzes ADB logcat logs to validate analytics event funnels for automated testing.

## MVP Scope
**Focus**: Android ADB logcat logs only
**Goal**: Validate single conversion funnel from log files
**Timeline**: 2-3 weeks for MVP

## Technical Stack
- **Language**: Go
- **CLI Framework**: Cobra CLI
- **Config Format**: YAML
- **Log Format**: ADB logcat (Android preset)

## MVP Features

### 1. Core Commands Structure
```
loglion
├── analyze          # Main analysis command
├── validate         # Validate config file
└── version          # Show version info
```

### 2. Essential Functionality

#### Command: `analyze`
```bash
loglion analyze --config funnel.yaml --log logcat.txt --format android
```

**Flags:**
- `--config, -c`: Path to funnel configuration file
- `--log, -l`: Path to log file
- `--format, -f`: Log format preset (default: "android")
- `--output, -o`: Output format (json, text) (default: "text")
- `--timeout, -t`: Session timeout in minutes (default: 30)

#### Command: `validate`
```bash
loglion validate --config funnel.yaml
```

### 3. Configuration File Format

```yaml
# funnel.yaml
version: "1.0"
format: "android"  # Log format preset

funnel:
  name: "Purchase Flow"
  session_key: "user_id"  # How to group events by user
  timeout_minutes: 30

  steps:
    - name: "Product View"
      event_pattern: "analytics.*page_view"
      required_properties:
        page: "/product"

    - name: "Add to Cart"
      event_pattern: "analytics.*add_to_cart"
      required_properties:
        product_id: ".*"  # regex pattern

    - name: "Purchase"
      event_pattern: "analytics.*purchase"
      required_properties:
        transaction_id: ".*"

# Optional: Define how to extract data from Android logs
android_parser:
  timestamp_format: "01-02 15:04:05.000"
  event_regex: ".*Analytics.*: (.*)"
  json_extraction: true  # Try to parse JSON from log line
```

### 4. Output Format

#### Success Case (Text):
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

#### JSON Output:
```json
{
  "funnel_name": "Purchase Flow",
  "total_sessions": 5,
  "completed_funnels": 3,
  "completion_rate": 0.6,
  "steps": [
    {
      "name": "Product View",
      "completed": 5,
      "completion_rate": 1.0
    }
  ],
  "sessions": [
    {
      "session_id": "user_123",
      "completed": true,
      "steps_completed": 3,
      "duration_minutes": 15
    }
  ]
}
```

## Implementation Plan

### Phase 1: Project Setup (2-3 days)
1. Initialize Go module with Cobra
2. Set up basic command structure
3. Create configuration file parsing
4. Set up basic testing framework

### Phase 2: Android Log Parser (3-4 days)
1. Implement ADB logcat parser
2. Extract timestamps, process info, and log messages
3. JSON extraction from log lines
4. Basic event pattern matching

### Phase 3: Funnel Analysis Engine (4-5 days)
1. Session grouping logic
2. Step-by-step funnel tracking
3. Timeout handling
4. Completion rate calculation

### Phase 4: Output & CLI Polish (2-3 days)
1. Text and JSON output formatting
2. Error handling and validation
3. Help text and documentation
4. Basic integration tests

## File Structure
```
loglion/
├── cmd/
│   ├── root.go
│   ├── analyze.go
│   ├── validate.go
│   └── version.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── parser/
│   │   ├── android.go
│   │   └── parser.go
│   ├── analyzer/
│   │   ├── funnel.go
│   │   └── session.go
│   └── output/
│       └── formatter.go
├── examples/
│   ├── funnel.yaml
│   └── sample_logcat.txt
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## Success Criteria for MVP
- [ ] Parse Android logcat files successfully
- [ ] Extract analytics events using regex patterns
- [ ] Group events by session/user
- [ ] Calculate funnel completion rates
- [ ] Output results in text and JSON formats
- [ ] Handle basic error cases gracefully
- [ ] Include example configuration and sample logs

## Future Enhancements (Post-MVP)
- iOS log format support
- Custom log format definitions
- Multiple funnel analysis in single run
- Visual HTML reports
- Real-time log monitoring
- CI/CD integration helpers
- Performance optimizations for large files

## Sample Test Case
```yaml
# Test config for e-commerce app
funnel:
  name: "Checkout Flow Test"
  session_key: "user_id"
  timeout_minutes: 10

  steps:
    - name: "App Launch"
      event_pattern: ".*app_launch.*"
    - name: "Product View"
      event_pattern: ".*screen_view.*"
      required_properties:
        screen_name: "product_detail"
    - name: "Purchase"
      event_pattern: ".*purchase.*"
```

This MVP focuses on core functionality while keeping the scope manageable for initial development and testing.
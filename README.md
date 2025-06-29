# LogLion

LogLion is a CLI tool for analyzing log files to track event funnels and count event patterns. It's designed to help with automation testing of client analytics and telemetry, as well as debugging and monitoring logs.

## Installation

### Install via go install

```bash
go install github.com/parfenovvs/loglion@latest
```

## Quick Start

### Funnel Analysis

Track multi-step user flows through your logs.

1. Create a parser config (`parser.yaml`):
```yaml
event_regex: "^(.*)$"
json_extraction: false
```

2. Create a funnel config (`funnel.yaml`):
```yaml
name: "User Flow"
steps:
  - name: "Login"
    event_pattern: "login"
  - name: "Action"
    event_pattern: "user_action"
  - name: "Logout"
    event_pattern: "logout"
```

3. Run analysis:
```bash
loglion funnel -p parser.yaml -f funnel.yaml -l log.txt
```

### Event Counting

Count how many times specific events occur in your logs.

```bash
# Count multiple patterns
loglion count -p parser.yaml -l log.txt "login" "logout" "error"

# Use regex patterns
loglion count -p parser.yaml -l log.txt "user_\\d+" "purchase"

# JSON output
loglion count -p parser.yaml -l log.txt --output json "login"
```

## Configuration Examples

**Simple text logs:**
```yaml
# parser.yaml
event_regex: "^(.*)$"
json_extraction: false
```

**Android logcat:**
```yaml
# parser.yaml
timestamp_format: "01-02 15:04:05.000"
event_regex: ".*Analytics: (.*)"
json_extraction: true
log_line_regex: "^(\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}\\.\\d{3})\\s+(\\d+)\\s+(\\d+)\\s+([VDIWEFS])\\s+([^:]+):\\s*(.*)$"
```

See `examples/` directory for more configurations and sample log files.

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
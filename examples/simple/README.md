# Simple Text Log Examples

This directory contains examples for parsing simple plain text logs.

## Files

- `simple-parser.yaml` - Parser configuration for simple text format
- `simple-funnel.yaml` - Example funnel configuration for basic workflow
- `sample_simple.txt` - Sample simple text log file for testing

## Usage

```bash
# Analyze simple text logs
loglion funnel --parser-config examples/simple/simple-parser.yaml --funnel-config examples/simple/simple-funnel.yaml --log examples/simple/sample_simple.txt

# Count events in simple logs
loglion count --parser-config examples/simple/simple-parser.yaml --log examples/simple/sample_simple.txt "login" "logout" "error"
```

## Format Details

Simple text format is the most basic log format:
- Each line is treated as a single log entry
- No structured parsing of timestamps or metadata
- Suitable for custom application logs or basic text files
- Event matching is done via regex patterns on the full line content
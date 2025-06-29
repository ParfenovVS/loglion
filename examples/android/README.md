# Android LogCat Examples

This directory contains examples for parsing Android logcat format logs.

## Files

- `logcat-parser.yaml` - Parser configuration for Android logcat format
- `sample_logcat_plain.txt` - Sample Android logcat file for testing
- `purchase-funnel.yaml` - Example funnel configuration

## Usage

```bash
# Analyze Android logcat logs
loglion funnel --parser-config examples/android/logcat-parser.yaml --funnel-config examples/android/purchase-funnel.yaml --log examples/android/sample_logcat_plain.txt

# Count events in Android logs
loglion count --parser-config examples/android/logcat-parser.yaml --log examples/android/sample_logcat_plain.txt "ActivityManager" "System"
```

## Format Details

Android logcat format typically includes:
- Timestamp
- Process ID (PID)
- Thread ID (TID)
- Log level (V/D/I/W/E/F)
- Tag
- Message content
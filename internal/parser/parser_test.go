package parser

import (
	"reflect"
	"testing"
	"time"
)

func TestLogFormat_String(t *testing.T) {
	tests := []struct {
		name   string
		format LogFormat
		want   string
	}{
		{
			name:   "logcat plain format",
			format: LogcatPlainFormat,
			want:   "logcat-plain",
		},
		{
			name:   "logcat json format",
			format: LogcatJSONFormat,
			want:   "logcat-json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(tt.format); got != tt.want {
				t.Errorf("LogFormat = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogEntry_Fields(t *testing.T) {
	timestamp := time.Now()
	eventData := map[string]interface{}{
		"event":   "test_event",
		"user_id": "123",
	}

	entry := &LogEntry{
		Timestamp: timestamp,
		Level:     "I",
		Tag:       "TestTag",
		PID:       1234,
		TID:       5678,
		Message:   "Test message",
		EventData: eventData,
	}

	// Test all fields are set correctly
	if !entry.Timestamp.Equal(timestamp) {
		t.Errorf("LogEntry.Timestamp = %v, want %v", entry.Timestamp, timestamp)
	}
	if entry.Level != "I" {
		t.Errorf("LogEntry.Level = %v, want %v", entry.Level, "I")
	}
	if entry.Tag != "TestTag" {
		t.Errorf("LogEntry.Tag = %v, want %v", entry.Tag, "TestTag")
	}
	if entry.PID != 1234 {
		t.Errorf("LogEntry.PID = %v, want %v", entry.PID, 1234)
	}
	if entry.TID != 5678 {
		t.Errorf("LogEntry.TID = %v, want %v", entry.TID, 5678)
	}
	if entry.Message != "Test message" {
		t.Errorf("LogEntry.Message = %v, want %v", entry.Message, "Test message")
	}
	if !reflect.DeepEqual(entry.EventData, eventData) {
		t.Errorf("LogEntry.EventData = %v, want %v", entry.EventData, eventData)
	}
}

func TestNewParser(t *testing.T) {
	tests := []struct {
		name   string
		format LogFormat
		want   string // We'll check the type name as string since we can't directly compare interface types
	}{
		{
			name:   "logcat plain format",
			format: LogcatPlainFormat,
			want:   "*parser.LogcatPlainParser",
		},
		{
			name:   "logcat json format",
			format: LogcatJSONFormat,
			want:   "*parser.LogcatJSONParser",
		},
		{
			name:   "unknown format defaults to plain",
			format: LogFormat("unknown"),
			want:   "*parser.LogcatPlainParser",
		},
		{
			name:   "empty format defaults to plain",
			format: LogFormat(""),
			want:   "*parser.LogcatPlainParser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.format)
			if parser == nil {
				t.Errorf("NewParser() returned nil")
				return
			}

			got := reflect.TypeOf(parser).String()
			if got != tt.want {
				t.Errorf("NewParser() type = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewParserWithConfig(t *testing.T) {
	tests := []struct {
		name            string
		format          LogFormat
		timestampFormat string
		eventRegex      string
		jsonExtraction  bool
		want            string
	}{
		{
			name:            "logcat plain with custom config",
			format:          LogcatPlainFormat,
			timestampFormat: "01-02 15:04:05.000",
			eventRegex:      `.*Analytics.*: (.*)`,
			jsonExtraction:  true,
			want:            "*parser.LogcatPlainParser",
		},
		{
			name:            "logcat json with custom config",
			format:          LogcatJSONFormat,
			timestampFormat: "01-02 15:04:05.000",
			eventRegex:      `.*Analytics.*: (.*)`,
			jsonExtraction:  false,
			want:            "*parser.LogcatJSONParser",
		},
		{
			name:            "unknown format defaults to plain with config",
			format:          LogFormat("invalid"),
			timestampFormat: "01-02 15:04:05.000",
			eventRegex:      `.*Test.*: (.*)`,
			jsonExtraction:  true,
			want:            "*parser.LogcatPlainParser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParserWithConfig(tt.format, tt.timestampFormat, tt.eventRegex, tt.jsonExtraction)
			if parser == nil {
				t.Errorf("NewParserWithConfig() returned nil")
				return
			}

			got := reflect.TypeOf(parser).String()
			if got != tt.want {
				t.Errorf("NewParserWithConfig() type = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParser_Interface(t *testing.T) {
	// Test that returned parsers implement the Parser interface
	formats := []LogFormat{
		LogcatPlainFormat,
		LogcatJSONFormat,
	}

	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			parser := NewParser(format)

			// Check that parser implements Parser interface by calling interface methods
			// This is a basic compile-time check
			var _ Parser = parser

			// Test that methods exist and don't panic (basic smoke test)
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Parser methods should not panic on basic calls")
				}
			}()

			// Test Parse with empty string (should return error, not panic)
			_, err := parser.Parse("")
			if err == nil {
				t.Errorf("Parse(\"\") should return error for empty input")
			}

			// Test ParseFile with invalid path (should return error, not panic)
			_, err = parser.ParseFile("/nonexistent/file.txt")
			if err == nil {
				t.Errorf("ParseFile() should return error for nonexistent file")
			}
		})
	}
}

func TestLogFormat_Constants(t *testing.T) {
	// Test that constants have expected values
	if LogcatPlainFormat != "logcat-plain" {
		t.Errorf("LogcatPlainFormat = %v, want %v", LogcatPlainFormat, "logcat-plain")
	}
	if LogcatJSONFormat != "logcat-json" {
		t.Errorf("LogcatJSONFormat = %v, want %v", LogcatJSONFormat, "logcat-json")
	}
}

func TestLogEntry_EmptyEventData(t *testing.T) {
	entry := &LogEntry{
		Timestamp: time.Now(),
		Level:     "D",
		Tag:       "Test",
		PID:       1,
		TID:       1,
		Message:   "test message",
		EventData: nil,
	}

	if entry.EventData != nil {
		t.Errorf("LogEntry.EventData should be nil when not set")
	}
}

func TestLogEntry_ZeroValues(t *testing.T) {
	entry := &LogEntry{}

	// Test zero values
	if !entry.Timestamp.IsZero() {
		t.Errorf("LogEntry.Timestamp should be zero value when not set")
	}
	if entry.Level != "" {
		t.Errorf("LogEntry.Level should be empty when not set")
	}
	if entry.Tag != "" {
		t.Errorf("LogEntry.Tag should be empty when not set")
	}
	if entry.PID != 0 {
		t.Errorf("LogEntry.PID should be 0 when not set")
	}
	if entry.TID != 0 {
		t.Errorf("LogEntry.TID should be 0 when not set")
	}
	if entry.Message != "" {
		t.Errorf("LogEntry.Message should be empty when not set")
	}
	if entry.EventData != nil {
		t.Errorf("LogEntry.EventData should be nil when not set")
	}
}

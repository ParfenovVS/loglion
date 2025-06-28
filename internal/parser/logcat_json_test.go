package parser

import (
	"os"
	"testing"
	"time"
)

func TestLogcatJSONParser_Parse(t *testing.T) {
	parser := NewLogcatJSONParser()

	// LogcatJSON parser should not support line-by-line parsing
	_, err := parser.Parse("some log line")
	if err == nil {
		t.Errorf("Parse() should return an error for LogcatJSON parser")
	}
}

func TestLogcatJSONParser_ParseFile_InvalidJSON(t *testing.T) {
	parser := NewLogcatJSONParser()

	// Create a temporary file with invalid JSON
	tmpFile, err := os.CreateTemp("", "invalid.logcat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write invalid JSON
	if _, err := tmpFile.WriteString("invalid json content"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test parsing
	_, err = parser.ParseFile(tmpFile.Name())
	if err == nil {
		t.Errorf("ParseFile() should return an error for invalid JSON")
	}
}

func TestLogcatJSONParser_ParseFile_ValidStructure(t *testing.T) {
	parser := NewLogcatJSONParser()

	// Create a temporary file with valid .logcat structure
	tmpFile, err := os.CreateTemp("", "valid.logcat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write valid .logcat JSON
	logcatContent := `{
  "metadata": {
    "device": {
      "deviceId": "test123",
      "name": "Test Device"
    },
    "filter": "test filter",
    "projectApplicationIds": ["com.test.app"]
  },
  "logcatMessages": [
    {
      "header": {
        "logLevel": "INFO",
        "pid": 1234,
        "tid": 5678,
        "applicationId": "com.test.app",
        "processName": "test_process",
        "tag": "TestTag",
        "timestamp": {
          "seconds": 1642248615,
          "nanos": 123000000
        }
      },
      "message": "Test log message"
    },
    {
      "header": {
        "logLevel": "DEBUG",
        "pid": 1234,
        "tid": 5678,
        "applicationId": "com.test.app",
        "processName": "test_process",
        "tag": "Analytics",
        "timestamp": {
          "seconds": 1642248616,
          "nanos": 456000000
        }
      },
      "message": "Analytics: {\"event\": \"test_event\", \"user_id\": \"user123\"}"
    }
  ]
}`

	if _, err := tmpFile.WriteString(logcatContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test parsing
	entries, err := parser.ParseFile(tmpFile.Name())
	if err != nil {
		t.Errorf("ParseFile() unexpected error: %v", err)
		return
	}

	if len(entries) != 2 {
		t.Errorf("ParseFile() expected 2 entries, got %d", len(entries))
		return
	}

	// Test first entry
	entry1 := entries[0]
	if entry1.Level != "INFO" {
		t.Errorf("Entry1 Level = %v, want INFO", entry1.Level)
	}
	if entry1.Tag != "TestTag" {
		t.Errorf("Entry1 Tag = %v, want TestTag", entry1.Tag)
	}
	if entry1.PID != 1234 {
		t.Errorf("Entry1 PID = %v, want 1234", entry1.PID)
	}
	if entry1.TID != 5678 {
		t.Errorf("Entry1 TID = %v, want 5678", entry1.TID)
	}
	if entry1.Message != "Test log message" {
		t.Errorf("Entry1 Message = %v, want 'Test log message'", entry1.Message)
	}

	// Test timestamp conversion
	expectedTime := time.Unix(1642248615, 123000000)
	if !entry1.Timestamp.Equal(expectedTime) {
		t.Errorf("Entry1 Timestamp = %v, want %v", entry1.Timestamp, expectedTime)
	}

	// Test second entry with analytics data
	entry2 := entries[1]
	if entry2.Level != "DEBUG" {
		t.Errorf("Entry2 Level = %v, want DEBUG", entry2.Level)
	}
	if entry2.Tag != "Analytics" {
		t.Errorf("Entry2 Tag = %v, want Analytics", entry2.Tag)
	}

	// Check if event data was extracted
	if entry2.EventData == nil {
		t.Errorf("Entry2 EventData should not be nil for analytics message")
		return
	}

	if event, exists := entry2.EventData["event"]; !exists || event != "test_event" {
		t.Errorf("Entry2 EventData[event] = %v, want 'test_event'", event)
	}

	if userID, exists := entry2.EventData["user_id"]; !exists || userID != "user123" {
		t.Errorf("Entry2 EventData[user_id] = %v, want 'user123'", userID)
	}
}

func TestLogcatJSONParser_JSONExtractionDisabled(t *testing.T) {
	// Test with JSON extraction disabled
	parser := NewLogcatJSONParserWithConfig("", `.*Analytics: (.*)`, false)

	// Create a temporary file with analytics message
	tmpFile, err := os.CreateTemp("", "test.logcat")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	logcatContent := `{
  "metadata": {"device": {}, "filter": "", "projectApplicationIds": []},
  "logcatMessages": [
    {
      "header": {
        "logLevel": "INFO",
        "pid": 1234,
        "tid": 5678,
        "applicationId": "com.test.app",
        "processName": "test_process",
        "tag": "Analytics",
        "timestamp": {"seconds": 1642248615, "nanos": 0}
      },
      "message": "Analytics: {\"event\": \"test\"}"
    }
  ]
}`

	if _, err := tmpFile.WriteString(logcatContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	entries, err := parser.ParseFile(tmpFile.Name())
	if err != nil {
		t.Errorf("ParseFile() unexpected error: %v", err)
		return
	}

	if len(entries) != 1 {
		t.Errorf("ParseFile() expected 1 entry, got %d", len(entries))
		return
	}

	if entries[0].EventData != nil {
		t.Errorf("EventData should be nil when JSON extraction is disabled")
	}
}
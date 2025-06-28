package parser

import (
	"testing"
	"time"
)

func TestPlainParser_Parse_SimpleFormat(t *testing.T) {
	parser := NewPlainParser()

	tests := []struct {
		name     string
		logLine  string
		wantErr  bool
		expected *LogEntry
	}{
		{
			name:    "simple event line",
			logLine: "event_1",
			wantErr: false,
			expected: &LogEntry{
				Timestamp: time.Time{}, // Zero time
				Level:     "",
				Tag:       "",
				PID:       0,
				TID:       0,
				Message:   "event_1",
				EventData: nil,
			},
		},
		{
			name:    "another simple event",
			logLine: "user_action_completed",
			wantErr: false,
			expected: &LogEntry{
				Timestamp: time.Time{},
				Level:     "",
				Tag:       "",
				PID:       0,
				TID:       0,
				Message:   "user_action_completed",
				EventData: nil,
			},
		},
		{
			name:    "empty line",
			logLine: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parser.Parse(tt.logLine)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if entry == nil {
				t.Errorf("Parse() returned nil entry")
				return
			}

			if entry.Message != tt.expected.Message {
				t.Errorf("Parse() Message = %v, want %v", entry.Message, tt.expected.Message)
			}
			if entry.Level != tt.expected.Level {
				t.Errorf("Parse() Level = %v, want %v", entry.Level, tt.expected.Level)
			}
			if entry.Tag != tt.expected.Tag {
				t.Errorf("Parse() Tag = %v, want %v", entry.Tag, tt.expected.Tag)
			}
			if entry.PID != tt.expected.PID {
				t.Errorf("Parse() PID = %v, want %v", entry.PID, tt.expected.PID)
			}
			if entry.TID != tt.expected.TID {
				t.Errorf("Parse() TID = %v, want %v", entry.TID, tt.expected.TID)
			}
		})
	}
}

func TestPlainParser_Parse_LogcatFormat(t *testing.T) {
	// Test logcat-style format with custom config
	parser := NewPlainParserWithConfig(
		"01-02 15:04:05.000",
		".*Analytics: (.*)",
		true,
		`^(\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3})\s+(\d+)\s+(\d+)\s+([VDIWEFS])\s+([^:]+):\s*(.*)$`,
	)

	tests := []struct {
		name     string
		logLine  string
		wantErr  bool
		expected *LogEntry
	}{
		{
			name:    "valid logcat analytics log with JSON",
			logLine: "01-15 10:30:15.123  1234  5678 I Analytics: {\"event\": \"app_launch\", \"user_id\": \"user_123\", \"timestamp\": 1642248615123}",
			wantErr: false,
			expected: &LogEntry{
				Level:   "I",
				Tag:     "Analytics",
				PID:     1234,
				TID:     5678,
				Message: "{\"event\": \"app_launch\", \"user_id\": \"user_123\", \"timestamp\": 1642248615123}",
				EventData: map[string]interface{}{
					"event":     "app_launch",
					"user_id":   "user_123",
					"timestamp": float64(1642248615123),
				},
			},
		},
		{
			name:    "valid logcat log without analytics",
			logLine: "01-15 10:30:15.123  1234  5678 D SystemServer: Starting service",
			wantErr: false,
			expected: &LogEntry{
				Level:     "D",
				Tag:       "SystemServer",
				PID:       1234,
				TID:       5678,
				Message:   "Starting service",
				EventData: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parser.Parse(tt.logLine)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if entry == nil {
				t.Errorf("Parse() returned nil entry")
				return
			}

			// Check basic fields
			if entry.Level != tt.expected.Level {
				t.Errorf("Parse() Level = %v, want %v", entry.Level, tt.expected.Level)
			}
			if entry.Tag != tt.expected.Tag {
				t.Errorf("Parse() Tag = %v, want %v", entry.Tag, tt.expected.Tag)
			}
			if entry.PID != tt.expected.PID {
				t.Errorf("Parse() PID = %v, want %v", entry.PID, tt.expected.PID)
			}
			if entry.TID != tt.expected.TID {
				t.Errorf("Parse() TID = %v, want %v", entry.TID, tt.expected.TID)
			}
			if entry.Message != tt.expected.Message {
				t.Errorf("Parse() Message = %v, want %v", entry.Message, tt.expected.Message)
			}

			// Check timestamp
			expectedTime, _ := time.Parse("01-02 15:04:05.000", "01-15 10:30:15.123")
			if !entry.Timestamp.Equal(expectedTime) {
				t.Errorf("Parse() Timestamp = %v, want %v", entry.Timestamp, expectedTime)
			}

			// Check EventData
			if tt.expected.EventData == nil {
				if entry.EventData != nil {
					t.Errorf("Parse() EventData = %v, want nil", entry.EventData)
				}
			} else {
				if entry.EventData == nil {
					t.Errorf("Parse() EventData = nil, want %v", tt.expected.EventData)
					return
				}

				for key, expectedVal := range tt.expected.EventData {
					if actualVal, exists := entry.EventData[key]; !exists {
						t.Errorf("Parse() EventData missing key %s", key)
					} else if actualVal != expectedVal {
						t.Errorf("Parse() EventData[%s] = %v, want %v", key, actualVal, expectedVal)
					}
				}
			}
		})
	}
}

func TestPlainParser_Parse_OSLogFormat(t *testing.T) {
	// Test with very simple format that just captures timestamp and message
	parser := NewPlainParserWithConfig(
		"2006-01-02 15:04:05.000000-0700",
		"Analytics: (.*)",
		true,
		`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{6}-\d{4})\s+(.*)$`,
	)

	tests := []struct {
		name     string
		logLine  string
		wantErr  bool
		expected *LogEntry
	}{
		{
			name:    "valid oslog line with analytics",
			logLine: "2023-01-28 13:13:28.923080-0500 0x17f616 Default 0x0 39182 0 myapp: Analytics: {\"event\": \"app_launch\"}",
			wantErr: false,
			expected: &LogEntry{
				Level:   "",
				Tag:     "",
				PID:     0,
				TID:     0,
				Message: "0x17f616 Default 0x0 39182 0 myapp: Analytics: {\"event\": \"app_launch\"}",
				EventData: map[string]interface{}{
					"event": "app_launch",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := parser.Parse(tt.logLine)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if entry == nil {
				t.Errorf("Parse() returned nil entry")
				return
			}

			// Check basic fields
			if entry.Message != tt.expected.Message {
				t.Errorf("Parse() Message = %v, want %v", entry.Message, tt.expected.Message)
			}

			// Check timestamp
			expectedTime, _ := time.Parse("2006-01-02 15:04:05.000000-0700", "2023-01-28 13:13:28.923080-0500")
			if !entry.Timestamp.Equal(expectedTime) {
				t.Errorf("Parse() Timestamp = %v, want %v", entry.Timestamp, expectedTime)
			}

			// Check EventData
			if tt.expected.EventData != nil {
				if entry.EventData == nil {
					t.Errorf("Parse() EventData = nil, want %v", tt.expected.EventData)
					return
				}

				for key, expectedVal := range tt.expected.EventData {
					if actualVal, exists := entry.EventData[key]; !exists {
						t.Errorf("Parse() EventData missing key %s", key)
					} else if actualVal != expectedVal {
						t.Errorf("Parse() EventData[%s] = %v, want %v", key, actualVal, expectedVal)
					}
				}
			}
		})
	}
}

func TestPlainParser_Parse_TimestampWithoutFormat(t *testing.T) {
	// Test with timestamp in log but no timestamp_format specified
	parser := NewPlainParserWithConfig(
		"", // No timestamp format
		"^(.*)$",
		false,
		`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) (.*)$`,
	)

	logLine := "2023-01-28 13:13:28 some event data"
	entry, err := parser.Parse(logLine)

	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
		return
	}

	// When timestamp_format is empty, timestamp group should be treated as message
	if entry.Timestamp != (time.Time{}) {
		t.Errorf("Parse() Timestamp should be zero when no format specified")
	}

	if entry.Message != "some event data" {
		t.Errorf("Parse() Message = %v, want 'some event data'", entry.Message)
	}
}

func TestPlainParser_JSONExtractionDisabled(t *testing.T) {
	// Test with JSON extraction disabled
	parser := NewPlainParserWithConfig(
		"",
		"Analytics: (.*)",
		false, // JSON extraction disabled
		"^(.*)$",
	)

	logLine := "Analytics: {\"event\": \"test\"}"
	entry, err := parser.Parse(logLine)

	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
		return
	}

	if entry.EventData != nil {
		t.Errorf("Parse() EventData should be nil when JSON extraction is disabled")
	}
}

func TestPlainParser_CustomEventRegex(t *testing.T) {
	// Test custom event extraction
	parser := NewPlainParserWithConfig(
		"",
		`Event: (.*)`,
		true,
		"^(.*)$",
	)

	logLine := "Event: {\"action\": \"click\", \"element\": \"button\"}"
	entry, err := parser.Parse(logLine)

	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
		return
	}

	if entry.EventData == nil {
		t.Errorf("Parse() EventData should not be nil for custom event regex")
		return
	}

	if action, exists := entry.EventData["action"]; !exists || action != "click" {
		t.Errorf("Parse() EventData[action] = %v, want 'click'", action)
	}
}
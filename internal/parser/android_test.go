package parser

import (
	"testing"
	"time"
)

func TestAndroidParser_Parse(t *testing.T) {
	parser := NewAndroidParser()

	tests := []struct {
		name        string
		logLine     string
		wantErr     bool
		expected    *LogEntry
	}{
		{
			name:    "valid analytics log with JSON",
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
			name:    "valid log without JSON",
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
		{
			name:    "analytics log with nested JSON",
			logLine: "01-15 10:30:15.123  1234  5678 I Analytics: {\"event\": \"page_view\", \"user_id\": \"user_123\", \"page\": \"/product\", \"product_id\": \"prod_456\"}",
			wantErr: false,
			expected: &LogEntry{
				Level:   "I",
				Tag:     "Analytics",
				PID:     1234,
				TID:     5678,
				Message: "{\"event\": \"page_view\", \"user_id\": \"user_123\", \"page\": \"/product\", \"product_id\": \"prod_456\"}",
				EventData: map[string]interface{}{
					"event":      "page_view",
					"user_id":    "user_123",
					"page":       "/product",
					"product_id": "prod_456",
				},
			},
		},
		{
			name:    "invalid log line format",
			logLine: "invalid log line",
			wantErr: true,
		},
		{
			name:    "empty log line",
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

			// Check timestamp (just verify it's parsed correctly)
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

func TestAndroidParser_ParseWithConfig(t *testing.T) {
	// Test custom configuration
	customParser := NewAndroidParserWithConfig(
		"01-02 15:04:05.000",
		`.*CustomTag.*: (.*)`,
		true,
	)

	logLine := "01-15 10:30:15.123  1234  5678 I CustomTag: {\"event\": \"custom_event\"}"
	entry, err := customParser.Parse(logLine)
	
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
		return
	}

	if entry.EventData == nil {
		t.Errorf("Parse() EventData should not be nil for custom parser")
		return
	}

	if event, exists := entry.EventData["event"]; !exists || event != "custom_event" {
		t.Errorf("Parse() EventData[event] = %v, want 'custom_event'", event)
	}
}

func TestAndroidParser_JSONExtractionDisabled(t *testing.T) {
	// Test with JSON extraction disabled
	parser := NewAndroidParserWithConfig(
		"01-02 15:04:05.000",
		`.*Analytics.*: (.*)`,
		false,
	)

	logLine := "01-15 10:30:15.123  1234  5678 I Analytics: {\"event\": \"test\"}"
	entry, err := parser.Parse(logLine)
	
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
		return
	}

	if entry.EventData != nil {
		t.Errorf("Parse() EventData should be nil when JSON extraction is disabled")
	}
}

func TestAndroidParser_ExtractEventData(t *testing.T) {
	parser := NewAndroidParser()

	tests := []struct {
		name        string
		logLine     string
		message     string
		expectData  bool
		expectedKey string
		expectedVal interface{}
	}{
		{
			name:        "direct JSON in message",
			logLine:     "01-15 10:30:15.123  1234  5678 I Test: {\"key\": \"value\"}",
			message:     "{\"key\": \"value\"}",
			expectData:  true,
			expectedKey: "key",
			expectedVal: "value",
		},
		{
			name:        "analytics pattern match",
			logLine:     "01-15 10:30:15.123  1234  5678 I Analytics: {\"event\": \"test\"}",
			message:     "{\"event\": \"test\"}",
			expectData:  true,
			expectedKey: "event",
			expectedVal: "test",
		},
		{
			name:       "invalid JSON",
			logLine:    "01-15 10:30:15.123  1234  5678 I Test: invalid json",
			message:    "invalid json",
			expectData: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &LogEntry{Message: tt.message}
			parser.extractEventData(entry, tt.logLine)

			if tt.expectData {
				if entry.EventData == nil {
					t.Errorf("extractEventData() EventData should not be nil")
					return
				}
				if val, exists := entry.EventData[tt.expectedKey]; !exists || val != tt.expectedVal {
					t.Errorf("extractEventData() EventData[%s] = %v, want %v", tt.expectedKey, val, tt.expectedVal)
				}
			} else {
				if entry.EventData != nil {
					t.Errorf("extractEventData() EventData should be nil for invalid JSON")
				}
			}
		})
	}
}
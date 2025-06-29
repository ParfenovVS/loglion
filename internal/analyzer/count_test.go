package analyzer

import (
	"github.com/parfenovvs/loglion/internal/parser"
	"testing"
	"time"
)

func TestNewCountAnalyzer(t *testing.T) {
	tests := []struct {
		name         string
		patterns     []string
		wantError    bool
		wantPatterns int
	}{
		{
			name:         "valid_single_pattern",
			patterns:     []string{"login"},
			wantError:    false,
			wantPatterns: 1,
		},
		{
			name:         "valid_multiple_patterns",
			patterns:     []string{"login", "logout", "error"},
			wantError:    false,
			wantPatterns: 3,
		},
		{
			name:         "valid_regex_patterns",
			patterns:     []string{"user_\\d+", "event_[a-z]+"},
			wantError:    false,
			wantPatterns: 2,
		},
		{
			name:         "empty_patterns",
			patterns:     []string{},
			wantError:    false,
			wantPatterns: 0,
		},
		{
			name:         "invalid_regex_pattern",
			patterns:     []string{"[invalid_regex"},
			wantError:    true,
			wantPatterns: 0,
		},
		{
			name:         "mixed_valid_invalid_patterns",
			patterns:     []string{"login", "[invalid_regex", "logout"},
			wantError:    true,
			wantPatterns: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := NewCountAnalyzer(tt.patterns)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewCountAnalyzer() expected error but got none")
				}
				if analyzer != nil {
					t.Errorf("NewCountAnalyzer() expected nil analyzer on error but got %v", analyzer)
				}
				return
			}

			if err != nil {
				t.Errorf("NewCountAnalyzer() unexpected error: %v", err)
				return
			}

			if analyzer == nil {
				t.Fatal("NewCountAnalyzer() returned nil without error")
			}

			if len(analyzer.patterns) != tt.wantPatterns {
				t.Errorf("NewCountAnalyzer() patterns count = %v, want %v", len(analyzer.patterns), tt.wantPatterns)
			}

			// Verify patterns are compiled correctly
			for i, pattern := range analyzer.patterns {
				if pattern.Name != tt.patterns[i] {
					t.Errorf("Pattern[%d].Name = %v, want %v", i, pattern.Name, tt.patterns[i])
				}
				if pattern.Pattern != tt.patterns[i] {
					t.Errorf("Pattern[%d].Pattern = %v, want %v", i, pattern.Pattern, tt.patterns[i])
				}
				if pattern.Regex == nil {
					t.Errorf("Pattern[%d].Regex is nil", i)
				}
			}
		})
	}
}

func TestCountAnalyzer_AnalyzeCount(t *testing.T) {
	tests := []struct {
		name              string
		patterns          []string
		entries           []*parser.LogEntry
		wantTotalEvents   int
		wantPatternCounts map[string]int
	}{
		{
			name:              "empty_entries",
			patterns:          []string{"login", "logout"},
			entries:           []*parser.LogEntry{},
			wantTotalEvents:   0,
			wantPatternCounts: map[string]int{},
		},
		{
			name:     "single_pattern_single_match",
			patterns: []string{"login"},
			entries: []*parser.LogEntry{
				{Message: "user login successful", Timestamp: time.Now()},
				{Message: "other event", Timestamp: time.Now()},
			},
			wantTotalEvents: 2,
			wantPatternCounts: map[string]int{
				"login": 1,
			},
		},
		{
			name:     "single_pattern_multiple_matches",
			patterns: []string{"login"},
			entries: []*parser.LogEntry{
				{Message: "user login successful", Timestamp: time.Now()},
				{Message: "admin login successful", Timestamp: time.Now()},
				{Message: "other event", Timestamp: time.Now()},
				{Message: "guest login failed", Timestamp: time.Now()},
			},
			wantTotalEvents: 4,
			wantPatternCounts: map[string]int{
				"login": 3,
			},
		},
		{
			name:     "multiple_patterns_multiple_matches",
			patterns: []string{"login", "logout", "error"},
			entries: []*parser.LogEntry{
				{Message: "user login successful", Timestamp: time.Now()},
				{Message: "user logout", Timestamp: time.Now()},
				{Message: "error occurred", Timestamp: time.Now()},
				{Message: "admin login successful", Timestamp: time.Now()},
				{Message: "other event", Timestamp: time.Now()},
				{Message: "database error", Timestamp: time.Now()},
				{Message: "user logout", Timestamp: time.Now()},
			},
			wantTotalEvents: 7,
			wantPatternCounts: map[string]int{
				"login":  2,
				"logout": 2,
				"error":  2,
			},
		},
		{
			name:     "no_matches",
			patterns: []string{"login", "logout"},
			entries: []*parser.LogEntry{
				{Message: "other event", Timestamp: time.Now()},
				{Message: "some data", Timestamp: time.Now()},
			},
			wantTotalEvents: 2,
			wantPatternCounts: map[string]int{
				"login":  0,
				"logout": 0,
			},
		},
		{
			name:     "regex_patterns",
			patterns: []string{"user_\\d+", "event_[a-z]+"},
			entries: []*parser.LogEntry{
				{Message: "user_123 logged in", Timestamp: time.Now()},
				{Message: "user_456 logged out", Timestamp: time.Now()},
				{Message: "event_start occurred", Timestamp: time.Now()},
				{Message: "event_end occurred", Timestamp: time.Now()},
				{Message: "user_abc invalid", Timestamp: time.Now()},  // Should not match user_\d+
				{Message: "event_123 invalid", Timestamp: time.Now()}, // Should not match event_[a-z]+
			},
			wantTotalEvents: 6,
			wantPatternCounts: map[string]int{
				"user_\\d+":    2,
				"event_[a-z]+": 2,
			},
		},
		{
			name:     "case_sensitive_patterns",
			patterns: []string{"Login", "login"},
			entries: []*parser.LogEntry{
				{Message: "Login successful", Timestamp: time.Now()},
				{Message: "login successful", Timestamp: time.Now()},
				{Message: "LOGIN successful", Timestamp: time.Now()}, // Should not match either
			},
			wantTotalEvents: 3,
			wantPatternCounts: map[string]int{
				"Login": 1,
				"login": 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := NewCountAnalyzer(tt.patterns)
			if err != nil {
				t.Fatalf("NewCountAnalyzer() unexpected error: %v", err)
			}

			result := analyzer.AnalyzeCount(tt.entries)

			if result.TotalEventsAnalyzed != tt.wantTotalEvents {
				t.Errorf("AnalyzeCount() TotalEventsAnalyzed = %v, want %v",
					result.TotalEventsAnalyzed, tt.wantTotalEvents)
			}

			if len(tt.wantPatternCounts) == 0 {
				// For empty results, check that we get empty pattern counts
				if len(result.PatternCounts) != 0 {
					t.Errorf("AnalyzeCount() expected empty PatternCounts but got %v", result.PatternCounts)
				}
				return
			}

			if len(result.PatternCounts) != len(tt.patterns) {
				t.Errorf("AnalyzeCount() PatternCounts length = %v, want %v",
					len(result.PatternCounts), len(tt.patterns))
			}

			// Check each pattern count
			for _, patternCount := range result.PatternCounts {
				expectedCount, exists := tt.wantPatternCounts[patternCount.Pattern]
				if !exists {
					t.Errorf("AnalyzeCount() unexpected pattern in results: %s", patternCount.Pattern)
					continue
				}

				if patternCount.Count != expectedCount {
					t.Errorf("AnalyzeCount() pattern %s count = %v, want %v",
						patternCount.Pattern, patternCount.Count, expectedCount)
				}
			}
		})
	}
}

func TestCountAnalyzer_EventMatchesPattern_RawMessage(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		entry     *parser.LogEntry
		wantMatch bool
	}{
		{
			name:    "simple_match",
			pattern: "login",
			entry: &parser.LogEntry{
				Message:   "user login successful",
				Timestamp: time.Now(),
			},
			wantMatch: true,
		},
		{
			name:    "no_match",
			pattern: "login",
			entry: &parser.LogEntry{
				Message:   "user logout",
				Timestamp: time.Now(),
			},
			wantMatch: false,
		},
		{
			name:    "regex_match",
			pattern: "user_\\d+",
			entry: &parser.LogEntry{
				Message:   "user_123 logged in",
				Timestamp: time.Now(),
			},
			wantMatch: true,
		},
		{
			name:    "regex_no_match",
			pattern: "user_\\d+",
			entry: &parser.LogEntry{
				Message:   "user_abc logged in",
				Timestamp: time.Now(),
			},
			wantMatch: false,
		},
		{
			name:    "case_sensitive",
			pattern: "Login",
			entry: &parser.LogEntry{
				Message:   "login successful",
				Timestamp: time.Now(),
			},
			wantMatch: false,
		},
		{
			name:    "partial_match",
			pattern: "log",
			entry: &parser.LogEntry{
				Message:   "user login successful",
				Timestamp: time.Now(),
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := NewCountAnalyzer([]string{tt.pattern})
			if err != nil {
				t.Fatalf("NewCountAnalyzer() unexpected error: %v", err)
			}

			pattern := analyzer.patterns[0]
			result := analyzer.eventMatchesPattern(tt.entry, pattern)

			if result != tt.wantMatch {
				t.Errorf("eventMatchesPattern() = %v, want %v", result, tt.wantMatch)
			}
		})
	}
}

func TestCountAnalyzer_EventMatchesPattern_StructuredData(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		entry     *parser.LogEntry
		wantMatch bool
	}{
		{
			name:    "structured_event_field_match",
			pattern: "user_signup",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event": "user_signup",
				},
			},
			wantMatch: true,
		},
		{
			name:    "structured_event_field_no_match",
			pattern: "user_signup",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event": "user_login",
				},
			},
			wantMatch: false,
		},
		{
			name:    "structured_no_event_field_fallback_to_message",
			pattern: "analytics",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"other_field": "value",
				},
			},
			wantMatch: true,
		},
		{
			name:    "structured_event_field_not_string",
			pattern: "123",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event": 123,
				},
			},
			wantMatch: false,
		},
		{
			name:    "structured_event_field_regex_match",
			pattern: "user_\\w+",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event": "user_signup",
				},
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := NewCountAnalyzer([]string{tt.pattern})
			if err != nil {
				t.Fatalf("NewCountAnalyzer() unexpected error: %v", err)
			}

			pattern := analyzer.patterns[0]
			result := analyzer.eventMatchesPattern(tt.entry, pattern)

			if result != tt.wantMatch {
				t.Errorf("eventMatchesPattern() = %v, want %v", result, tt.wantMatch)
			}
		})
	}
}

func TestCountAnalyzer_AnalyzeCount_EdgeCases(t *testing.T) {
	t.Run("nil_entries", func(t *testing.T) {
		analyzer, err := NewCountAnalyzer([]string{"test"})
		if err != nil {
			t.Fatalf("NewCountAnalyzer() unexpected error: %v", err)
		}

		result := analyzer.AnalyzeCount(nil)

		if result.TotalEventsAnalyzed != 0 {
			t.Errorf("AnalyzeCount(nil) TotalEventsAnalyzed = %v, want 0", result.TotalEventsAnalyzed)
		}
		if len(result.PatternCounts) != 0 {
			t.Errorf("AnalyzeCount(nil) should return empty PatternCounts")
		}
	})

	t.Run("empty_pattern_list", func(t *testing.T) {
		analyzer, err := NewCountAnalyzer([]string{})
		if err != nil {
			t.Fatalf("NewCountAnalyzer() unexpected error: %v", err)
		}

		entries := []*parser.LogEntry{
			{Message: "test message", Timestamp: time.Now()},
		}

		result := analyzer.AnalyzeCount(entries)

		if result.TotalEventsAnalyzed != 1 {
			t.Errorf("AnalyzeCount() TotalEventsAnalyzed = %v, want 1", result.TotalEventsAnalyzed)
		}
		if len(result.PatternCounts) != 0 {
			t.Errorf("AnalyzeCount() with empty patterns should return empty PatternCounts")
		}
	})

	t.Run("large_entry_set", func(t *testing.T) {
		analyzer, err := NewCountAnalyzer([]string{"test"})
		if err != nil {
			t.Fatalf("NewCountAnalyzer() unexpected error: %v", err)
		}

		// Create 1000 entries, half matching the pattern
		entries := make([]*parser.LogEntry, 1000)
		for i := 0; i < 1000; i++ {
			message := "other message"
			if i%2 == 0 {
				message = "test message"
			}
			entries[i] = &parser.LogEntry{
				Message:   message,
				Timestamp: time.Now(),
			}
		}

		result := analyzer.AnalyzeCount(entries)

		if result.TotalEventsAnalyzed != 1000 {
			t.Errorf("AnalyzeCount() TotalEventsAnalyzed = %v, want 1000", result.TotalEventsAnalyzed)
		}
		if len(result.PatternCounts) != 1 {
			t.Errorf("AnalyzeCount() PatternCounts length = %v, want 1", len(result.PatternCounts))
		}
		if result.PatternCounts[0].Count != 500 {
			t.Errorf("AnalyzeCount() pattern count = %v, want 500", result.PatternCounts[0].Count)
		}
	})
}

func TestCountAnalyzer_ComplexRegexPatterns(t *testing.T) {
	tests := []struct {
		name         string
		pattern      string
		testMessages []string
		wantMatches  int
	}{
		{
			name:    "email_pattern",
			pattern: "\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}\\b",
			testMessages: []string{
				"User test@example.com logged in",
				"Invalid email test@",
				"Another user admin@company.org signed up",
				"No email in this message",
			},
			wantMatches: 2,
		},
		{
			name:    "ip_address_pattern",
			pattern: "\\b(?:[0-9]{1,3}\\.){3}[0-9]{1,3}\\b",
			testMessages: []string{
				"Request from 192.168.1.1",
				"Server 10.0.0.1 responded",
				"Invalid IP 999.999.999.999",
				"Another request from 127.0.0.1",
				"No IP in this message",
			},
			wantMatches: 4,
		},
		{
			name:    "timestamp_pattern",
			pattern: "\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}",
			testMessages: []string{
				"Event at 2024-01-15 14:30:25",
				"Another event at 2024-12-31 23:59:59",
				"Invalid timestamp 24-01-15 14:30:25",
				"No timestamp here",
			},
			wantMatches: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer, err := NewCountAnalyzer([]string{tt.pattern})
			if err != nil {
				t.Fatalf("NewCountAnalyzer() unexpected error: %v", err)
			}

			entries := make([]*parser.LogEntry, len(tt.testMessages))
			for i, msg := range tt.testMessages {
				entries[i] = &parser.LogEntry{
					Message:   msg,
					Timestamp: time.Now(),
				}
			}

			result := analyzer.AnalyzeCount(entries)

			if result.TotalEventsAnalyzed != len(tt.testMessages) {
				t.Errorf("AnalyzeCount() TotalEventsAnalyzed = %v, want %v",
					result.TotalEventsAnalyzed, len(tt.testMessages))
			}

			if len(result.PatternCounts) != 1 {
				t.Errorf("AnalyzeCount() PatternCounts length = %v, want 1", len(result.PatternCounts))
			}

			if result.PatternCounts[0].Count != tt.wantMatches {
				t.Errorf("AnalyzeCount() pattern count = %v, want %v",
					result.PatternCounts[0].Count, tt.wantMatches)
			}
		})
	}
}

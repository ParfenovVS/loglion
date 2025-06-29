package analyzer

import (
	"github.com/parfenovvs/loglion/internal/config"
	"github.com/parfenovvs/loglion/internal/parser"
	"testing"
	"time"
)

func TestNewFunnelAnalyzer(t *testing.T) {
	cfg := &config.FunnelConfig{
		Name: "test_funnel",
		Steps: []config.Step{
			{Name: "step1", EventPattern: "event1"},
			{Name: "step2", EventPattern: "event2"},
		},
	}

	analyzer := NewFunnelAnalyzer(cfg)

	if analyzer == nil {
		t.Fatal("NewFunnelAnalyzer() returned nil")
	}
	if analyzer.config != cfg {
		t.Error("NewFunnelAnalyzer() did not store config correctly")
	}
}

func TestAnalyzeFunnel(t *testing.T) {
	tests := []struct {
		name              string
		config            *config.FunnelConfig
		entries           []*parser.LogEntry
		limit             int
		wantCompleted     bool
		wantStepCounts    []int
		wantTotalEvents   int
		wantDropOffsCount int
	}{
		{
			name: "empty_entries",
			config: &config.FunnelConfig{
				Name: "test",
				Steps: []config.Step{
					{Name: "step1", EventPattern: "event1"},
				},
			},
			entries:           []*parser.LogEntry{},
			limit:             0,
			wantCompleted:     false,
			wantStepCounts:    []int{},
			wantTotalEvents:   0,
			wantDropOffsCount: 0,
		},
		{
			name: "single_complete_funnel_mode1",
			config: &config.FunnelConfig{
				Name: "test",
				Steps: []config.Step{
					{Name: "step1", EventPattern: "event1"},
					{Name: "step2", EventPattern: "event2"},
				},
			},
			entries: []*parser.LogEntry{
				{Message: "event1", Timestamp: time.Now()},
				{Message: "event2", Timestamp: time.Now()},
			},
			limit:             0,
			wantCompleted:     true,
			wantStepCounts:    []int{1, 1},
			wantTotalEvents:   2,
			wantDropOffsCount: 1,
		},
		{
			name: "incomplete_funnel",
			config: &config.FunnelConfig{
				Name: "test",
				Steps: []config.Step{
					{Name: "step1", EventPattern: "event1"},
					{Name: "step2", EventPattern: "event2"},
				},
			},
			entries: []*parser.LogEntry{
				{Message: "event1", Timestamp: time.Now()},
				{Message: "other_event", Timestamp: time.Now()},
			},
			limit:             0,
			wantCompleted:     false,
			wantStepCounts:    []int{1, 0},
			wantTotalEvents:   2,
			wantDropOffsCount: 1,
		},
		{
			name: "multiple_complete_funnels_mode2",
			config: &config.FunnelConfig{
				Name: "test",
				Steps: []config.Step{
					{Name: "step1", EventPattern: "event1"},
					{Name: "step2", EventPattern: "event2"},
				},
			},
			entries: []*parser.LogEntry{
				{Message: "event1", Timestamp: time.Now()},
				{Message: "event2", Timestamp: time.Now()},
				{Message: "event1", Timestamp: time.Now()},
				{Message: "event2", Timestamp: time.Now()},
			},
			limit:             1,
			wantCompleted:     true,
			wantStepCounts:    []int{1, 1},
			wantTotalEvents:   4,
			wantDropOffsCount: 1,
		},
		{
			name: "out_of_order_events",
			config: &config.FunnelConfig{
				Name: "test",
				Steps: []config.Step{
					{Name: "step1", EventPattern: "event1"},
					{Name: "step2", EventPattern: "event2"},
				},
			},
			entries: []*parser.LogEntry{
				{Message: "event2", Timestamp: time.Now()},
				{Message: "event1", Timestamp: time.Now()},
				{Message: "event2", Timestamp: time.Now()},
			},
			limit:             0,
			wantCompleted:     true,
			wantStepCounts:    []int{1, 1},
			wantTotalEvents:   3,
			wantDropOffsCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewFunnelAnalyzer(tt.config)
			result := analyzer.AnalyzeFunnel(tt.entries, tt.limit)

			if result.FunnelCompleted != tt.wantCompleted {
				t.Errorf("AnalyzeFunnel() FunnelCompleted = %v, want %v", result.FunnelCompleted, tt.wantCompleted)
			}

			if result.TotalEventsAnalyzed != tt.wantTotalEvents {
				t.Errorf("AnalyzeFunnel() TotalEventsAnalyzed = %v, want %v", result.TotalEventsAnalyzed, tt.wantTotalEvents)
			}

			if len(tt.wantStepCounts) > 0 {
				if len(result.Steps) != len(tt.wantStepCounts) {
					t.Errorf("AnalyzeFunnel() Steps count = %v, want %v", len(result.Steps), len(tt.wantStepCounts))
				}

				for i, expectedCount := range tt.wantStepCounts {
					if i < len(result.Steps) && result.Steps[i].EventCount != expectedCount {
						t.Errorf("AnalyzeFunnel() Step[%d].EventCount = %v, want %v", i, result.Steps[i].EventCount, expectedCount)
					}
				}
			}

			if len(result.DropOffs) != tt.wantDropOffsCount {
				t.Errorf("AnalyzeFunnel() DropOffs count = %v, want %v", len(result.DropOffs), tt.wantDropOffsCount)
			}

			if result.FunnelName != tt.config.Name {
				t.Errorf("AnalyzeFunnel() FunnelName = %v, want %v", result.FunnelName, tt.config.Name)
			}
		})
	}
}

func TestEventMatchesStep(t *testing.T) {
	tests := []struct {
		name      string
		entry     *parser.LogEntry
		step      config.Step
		wantMatch bool
	}{
		{
			name: "simple_message_match",
			entry: &parser.LogEntry{
				Message:   "user_login_success",
				Timestamp: time.Now(),
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "user_login",
			},
			wantMatch: true,
		},
		{
			name: "message_no_match",
			entry: &parser.LogEntry{
				Message:   "user_logout",
				Timestamp: time.Now(),
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "user_login",
			},
			wantMatch: false,
		},
		{
			name: "structured_event_data_match",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event": "user_signup",
				},
			},
			step: config.Step{
				Name:         "signup",
				EventPattern: "user_signup",
			},
			wantMatch: true,
		},
		{
			name: "structured_event_data_no_match",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event": "user_login",
				},
			},
			step: config.Step{
				Name:         "signup",
				EventPattern: "user_signup",
			},
			wantMatch: false,
		},
		{
			name: "event_data_without_event_field",
			entry: &parser.LogEntry{
				Message:   "user_purchase",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"other_field": "value",
				},
			},
			step: config.Step{
				Name:         "purchase",
				EventPattern: "user_purchase",
			},
			wantMatch: true,
		},
		{
			name: "regex_pattern_match",
			entry: &parser.LogEntry{
				Message:   "user_123_login",
				Timestamp: time.Now(),
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "user_\\d+_login",
			},
			wantMatch: true,
		},
		{
			name: "invalid_regex_pattern",
			entry: &parser.LogEntry{
				Message:   "user_login",
				Timestamp: time.Now(),
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "[invalid_regex",
			},
			wantMatch: false,
		},
		{
			name: "event_field_not_string",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event": 123,
				},
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "user_login",
			},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &FunnelAnalyzer{
				config: &config.FunnelConfig{},
			}

			result := analyzer.eventMatchesStep(tt.entry, tt.step)
			if result != tt.wantMatch {
				t.Errorf("eventMatchesStep() = %v, want %v", result, tt.wantMatch)
			}
		})
	}
}

func TestEventMatchesStepWithRequiredProperties(t *testing.T) {
	tests := []struct {
		name      string
		entry     *parser.LogEntry
		step      config.Step
		wantMatch bool
	}{
		{
			name: "required_properties_match",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event":   "user_login",
					"user_id": "123",
					"source":  "mobile",
				},
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "user_login",
				RequiredProperties: map[string]string{
					"user_id": "\\d+",
					"source":  "mobile",
				},
			},
			wantMatch: true,
		},
		{
			name: "required_properties_no_match",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event":   "user_login",
					"user_id": "abc",
					"source":  "mobile",
				},
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "user_login",
				RequiredProperties: map[string]string{
					"user_id": "\\d+",
					"source":  "mobile",
				},
			},
			wantMatch: false,
		},
		{
			name: "missing_required_property",
			entry: &parser.LogEntry{
				Message:   "analytics event",
				Timestamp: time.Now(),
				EventData: map[string]interface{}{
					"event":  "user_login",
					"source": "mobile",
				},
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "user_login",
				RequiredProperties: map[string]string{
					"user_id": "\\d+",
					"source":  "mobile",
				},
			},
			wantMatch: false,
		},
		{
			name: "no_structured_data_with_required_props",
			entry: &parser.LogEntry{
				Message:   "user_login",
				Timestamp: time.Now(),
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "user_login",
				RequiredProperties: map[string]string{
					"user_id": "\\d+",
				},
			},
			wantMatch: false,
		},
		{
			name: "no_required_properties",
			entry: &parser.LogEntry{
				Message:   "user_login",
				Timestamp: time.Now(),
			},
			step: config.Step{
				Name:         "login",
				EventPattern: "user_login",
			},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &FunnelAnalyzer{
				config: &config.FunnelConfig{},
			}

			result := analyzer.eventMatchesStep(tt.entry, tt.step)
			if result != tt.wantMatch {
				t.Errorf("eventMatchesStep() = %v, want %v", result, tt.wantMatch)
			}
		})
	}
}

func TestCheckRequiredProperties(t *testing.T) {
	tests := []struct {
		name          string
		eventData     map[string]interface{}
		requiredProps map[string]string
		wantMatch     bool
	}{
		{
			name: "all_properties_match",
			eventData: map[string]interface{}{
				"user_id": "123",
				"source":  "mobile",
				"version": "1.0.0",
			},
			requiredProps: map[string]string{
				"user_id": "\\d+",
				"source":  "mobile",
			},
			wantMatch: true,
		},
		{
			name: "property_pattern_no_match",
			eventData: map[string]interface{}{
				"user_id": "abc",
				"source":  "mobile",
			},
			requiredProps: map[string]string{
				"user_id": "\\d+",
				"source":  "mobile",
			},
			wantMatch: false,
		},
		{
			name: "missing_property",
			eventData: map[string]interface{}{
				"source": "mobile",
			},
			requiredProps: map[string]string{
				"user_id": "\\d+",
				"source":  "mobile",
			},
			wantMatch: false,
		},
		{
			name: "property_not_string",
			eventData: map[string]interface{}{
				"user_id": 123,
				"source":  "mobile",
			},
			requiredProps: map[string]string{
				"user_id": "\\d+",
				"source":  "mobile",
			},
			wantMatch: false,
		},
		{
			name: "invalid_regex_pattern",
			eventData: map[string]interface{}{
				"user_id": "123",
			},
			requiredProps: map[string]string{
				"user_id": "[invalid_regex",
			},
			wantMatch: false,
		},
		{
			name:          "no_required_properties",
			eventData:     map[string]interface{}{"key": "value"},
			requiredProps: map[string]string{},
			wantMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := &FunnelAnalyzer{
				config: &config.FunnelConfig{},
			}

			result := analyzer.checkRequiredProperties(tt.eventData, tt.requiredProps)
			if result != tt.wantMatch {
				t.Errorf("checkRequiredProperties() = %v, want %v", result, tt.wantMatch)
			}
		})
	}
}

func TestDropOffCalculation(t *testing.T) {
	cfg := &config.FunnelConfig{
		Name: "test",
		Steps: []config.Step{
			{Name: "step1", EventPattern: "event1"},
			{Name: "step2", EventPattern: "event2"},
			{Name: "step3", EventPattern: "event3"},
		},
	}

	// Create a sequential funnel completion
	entries := []*parser.LogEntry{
		{Message: "event1", Timestamp: time.Now()},
		{Message: "event2", Timestamp: time.Now()},
		{Message: "event1", Timestamp: time.Now()},
		{Message: "event2", Timestamp: time.Now()},
		{Message: "event3", Timestamp: time.Now()},
	}

	analyzer := NewFunnelAnalyzer(cfg)
	result := analyzer.AnalyzeFunnel(entries, 0)

	if len(result.DropOffs) != 2 {
		t.Errorf("Expected 2 drop-offs, got %d", len(result.DropOffs))
	}

	// First drop-off: step1 to step2
	if len(result.DropOffs) > 0 && (result.DropOffs[0].From != "step1" || result.DropOffs[0].To != "step2") {
		t.Errorf("Expected drop-off from step1 to step2, got %s to %s", result.DropOffs[0].From, result.DropOffs[0].To)
	}

	// Second drop-off: step2 to step3
	if len(result.DropOffs) > 1 && (result.DropOffs[1].From != "step2" || result.DropOffs[1].To != "step3") {
		t.Errorf("Expected drop-off from step2 to step3, got %s to %s", result.DropOffs[1].From, result.DropOffs[1].To)
	}
}

func TestPercentageCalculation(t *testing.T) {
	cfg := &config.FunnelConfig{
		Name: "test",
		Steps: []config.Step{
			{Name: "step1", EventPattern: "event1"},
			{Name: "step2", EventPattern: "event2"},
		},
	}

	// Create entries where we have 2 event1s followed by 1 event2 in sequence
	entries := []*parser.LogEntry{
		{Message: "event1", Timestamp: time.Now()},
		{Message: "event2", Timestamp: time.Now()},
		{Message: "event1", Timestamp: time.Now()},
		{Message: "other", Timestamp: time.Now()}, // This should cause step2 to have lower count
	}

	analyzer := NewFunnelAnalyzer(cfg)
	result := analyzer.AnalyzeFunnel(entries, 0)

	// Step 1 should have 100% (base)
	if len(result.Steps) > 0 && result.Steps[0].Percentage != 100.0 {
		t.Errorf("Expected step1 percentage to be 100.0, got %f", result.Steps[0].Percentage)
	}

	// Step 2 should have lower percentage than step 1
	if len(result.Steps) > 1 && result.Steps[1].Percentage >= result.Steps[0].Percentage {
		t.Errorf("Expected step2 percentage to be less than step1, got step1=%f step2=%f", result.Steps[0].Percentage, result.Steps[1].Percentage)
	}
}
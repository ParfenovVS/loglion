package output

import (
	"encoding/json"
	"loglion/internal/analyzer"
	"reflect"
	"strings"
	"testing"
)

func TestOutputFormat_Constants(t *testing.T) {
	if TextFormat != "text" {
		t.Errorf("TextFormat = %v, want %v", TextFormat, "text")
	}
	if JSONFormat != "json" {
		t.Errorf("JSONFormat = %v, want %v", JSONFormat, "json")
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
		want   string
	}{
		{
			name:   "text format",
			format: TextFormat,
			want:   "*output.TextFormatter",
		},
		{
			name:   "json format",
			format: JSONFormat,
			want:   "*output.JSONFormatter",
		},
		{
			name:   "unknown format defaults to text",
			format: OutputFormat("unknown"),
			want:   "*output.TextFormatter",
		},
		{
			name:   "empty format defaults to text",
			format: OutputFormat(""),
			want:   "*output.TextFormatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format)
			if formatter == nil {
				t.Errorf("NewFormatter() returned nil")
				return
			}

			got := reflect.TypeOf(formatter).String()
			if got != tt.want {
				t.Errorf("NewFormatter() type = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextFormatter_Format_EmptyResult(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "Test Funnel",
		TotalEventsAnalyzed: 0,
		FunnelCompleted:     false,
		Steps:               []analyzer.StepResult{},
		DropOffs:            []analyzer.DropOff{},
	}

	output, err := formatter.Format(result)
	if err != nil {
		t.Errorf("Format() unexpected error: %v", err)
		return
	}

	expected := "❌ No events found\n"
	if output != expected {
		t.Errorf("Format() = %q, want %q", output, expected)
	}
}

func TestTextFormatter_Format_CompletedFunnel(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "User Registration",
		TotalEventsAnalyzed: 100,
		FunnelCompleted:     true,
		Steps: []analyzer.StepResult{
			{Name: "App Launch", EventCount: 100, Percentage: 100.0},
			{Name: "Sign Up Click", EventCount: 50, Percentage: 50.0},
			{Name: "Form Submit", EventCount: 30, Percentage: 30.0},
		},
		DropOffs: []analyzer.DropOff{
			{From: "App Launch", To: "Sign Up Click", EventsLost: 50, DropOffRate: 50.0},
			{From: "Sign Up Click", To: "Form Submit", EventsLost: 20, DropOffRate: 40.0},
		},
	}

	output, err := formatter.Format(result)
	if err != nil {
		t.Errorf("Format() unexpected error: %v", err)
		return
	}

	// Check key components of the output
	if !strings.Contains(output, "✅ Funnel Analysis Complete") {
		t.Errorf("Format() should contain success icon and title")
	}
	if !strings.Contains(output, "Funnel: User Registration") {
		t.Errorf("Format() should contain funnel name")
	}
	if !strings.Contains(output, "Total Events Analyzed: 100") {
		t.Errorf("Format() should contain total events")
	}
	if !strings.Contains(output, "Funnel Completed: Yes") {
		t.Errorf("Format() should indicate funnel completion")
	}
	if !strings.Contains(output, "1. App Launch: 100 events (100.0%)") {
		t.Errorf("Format() should contain step breakdown")
	}
	if !strings.Contains(output, "Drop-off Analysis:") {
		t.Errorf("Format() should contain drop-off analysis section")
	}
	if !strings.Contains(output, "App Launch → Sign Up Click: 50 events lost (50.0% drop-off)") {
		t.Errorf("Format() should contain drop-off details")
	}
}

func TestTextFormatter_Format_IncompleteFunnel(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "Purchase Flow",
		TotalEventsAnalyzed: 50,
		FunnelCompleted:     false,
		Steps: []analyzer.StepResult{
			{Name: "Product View", EventCount: 50, Percentage: 100.0},
			{Name: "Add to Cart", EventCount: 20, Percentage: 40.0},
		},
		DropOffs: []analyzer.DropOff{
			{From: "Product View", To: "Add to Cart", EventsLost: 30, DropOffRate: 60.0},
		},
	}

	output, err := formatter.Format(result)
	if err != nil {
		t.Errorf("Format() unexpected error: %v", err)
		return
	}

	if !strings.Contains(output, "❌ Funnel Analysis Complete") {
		t.Errorf("Format() should contain failure icon for incomplete funnel")
	}
	if !strings.Contains(output, "Funnel Completed: No") {
		t.Errorf("Format() should indicate funnel is not completed")
	}
}

func TestTextFormatter_Format_NoDropOffs(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "Simple Flow",
		TotalEventsAnalyzed: 10,
		FunnelCompleted:     true,
		Steps: []analyzer.StepResult{
			{Name: "Start", EventCount: 10, Percentage: 100.0},
		},
		DropOffs: []analyzer.DropOff{},
	}

	output, err := formatter.Format(result)
	if err != nil {
		t.Errorf("Format() unexpected error: %v", err)
		return
	}

	if strings.Contains(output, "Drop-off Analysis:") {
		t.Errorf("Format() should not contain drop-off section when no drop-offs exist")
	}
}

func TestJSONFormatter_Format_ValidResult(t *testing.T) {
	formatter := &JSONFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "Test Funnel",
		TotalEventsAnalyzed: 100,
		FunnelCompleted:     true,
		Steps: []analyzer.StepResult{
			{Name: "Step 1", EventCount: 100, Percentage: 100.0},
			{Name: "Step 2", EventCount: 80, Percentage: 80.0},
		},
		DropOffs: []analyzer.DropOff{
			{From: "Step 1", To: "Step 2", EventsLost: 20, DropOffRate: 20.0},
		},
	}

	output, err := formatter.Format(result)
	if err != nil {
		t.Errorf("Format() unexpected error: %v", err)
		return
	}

	// Verify it's valid JSON
	var parsed analyzer.FunnelResult
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("Format() output is not valid JSON: %v", err)
		return
	}

	// Verify content matches original
	if parsed.FunnelName != result.FunnelName {
		t.Errorf("JSON FunnelName = %v, want %v", parsed.FunnelName, result.FunnelName)
	}
	if parsed.TotalEventsAnalyzed != result.TotalEventsAnalyzed {
		t.Errorf("JSON TotalEventsAnalyzed = %v, want %v", parsed.TotalEventsAnalyzed, result.TotalEventsAnalyzed)
	}
	if parsed.FunnelCompleted != result.FunnelCompleted {
		t.Errorf("JSON FunnelCompleted = %v, want %v", parsed.FunnelCompleted, result.FunnelCompleted)
	}
	if len(parsed.Steps) != len(result.Steps) {
		t.Errorf("JSON Steps length = %v, want %v", len(parsed.Steps), len(result.Steps))
	}
	if len(parsed.DropOffs) != len(result.DropOffs) {
		t.Errorf("JSON DropOffs length = %v, want %v", len(parsed.DropOffs), len(result.DropOffs))
	}
}

func TestJSONFormatter_Format_EmptyResult(t *testing.T) {
	formatter := &JSONFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "Empty Funnel",
		TotalEventsAnalyzed: 0,
		FunnelCompleted:     false,
		Steps:               []analyzer.StepResult{},
		DropOffs:            []analyzer.DropOff{},
	}

	output, err := formatter.Format(result)
	if err != nil {
		t.Errorf("Format() unexpected error: %v", err)
		return
	}

	// Verify it's valid JSON
	var parsed analyzer.FunnelResult
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("Format() output is not valid JSON: %v", err)
		return
	}

	if parsed.TotalEventsAnalyzed != 0 {
		t.Errorf("JSON TotalEventsAnalyzed = %v, want 0", parsed.TotalEventsAnalyzed)
	}
	if len(parsed.Steps) != 0 {
		t.Errorf("JSON Steps should be empty array")
	}
	if len(parsed.DropOffs) != 0 {
		t.Errorf("JSON DropOffs should be empty array")
	}
}

func TestJSONFormatter_Format_NilResult(t *testing.T) {
	formatter := &JSONFormatter{}

	// The JSONFormatter panics on nil input, which is expected behavior
	// since the function expects a valid FunnelResult pointer
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Format(nil) should panic")
		}
	}()

	formatter.Format(nil)
}

func TestFormatter_Interface(t *testing.T) {
	// Test that formatters implement the Formatter interface
	formatters := []Formatter{
		&TextFormatter{},
		&JSONFormatter{},
	}

	for i, formatter := range formatters {
		t.Run(reflect.TypeOf(formatter).String(), func(t *testing.T) {
			// Check that formatter implements Formatter interface
			var _ Formatter = formatter

			// Test that Format method exists and doesn't panic on basic calls
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Formatter[%d] Format() should not panic", i)
				}
			}()

			// Create a minimal result for testing
			result := &analyzer.FunnelResult{
				FunnelName:          "Test",
				TotalEventsAnalyzed: 1,
				FunnelCompleted:     false,
				Steps:               []analyzer.StepResult{},
				DropOffs:            []analyzer.DropOff{},
			}

			_, err := formatter.Format(result)
			if err != nil {
				t.Errorf("Formatter[%d] Format() unexpected error: %v", i, err)
			}
		})
	}
}

func TestTextFormatter_Format_SpecialCharacters(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "Test & Special \"Characters\"",
		TotalEventsAnalyzed: 5,
		FunnelCompleted:     true,
		Steps: []analyzer.StepResult{
			{Name: "Step with <brackets>", EventCount: 5, Percentage: 100.0},
		},
		DropOffs: []analyzer.DropOff{},
	}

	output, err := formatter.Format(result)
	if err != nil {
		t.Errorf("Format() unexpected error: %v", err)
		return
	}

	if !strings.Contains(output, "Test & Special \"Characters\"") {
		t.Errorf("Format() should handle special characters in funnel name")
	}
	if !strings.Contains(output, "Step with <brackets>") {
		t.Errorf("Format() should handle special characters in step names")
	}
}

func TestJSONFormatter_Format_SpecialCharacters(t *testing.T) {
	formatter := &JSONFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "Test & Special \"Characters\"",
		TotalEventsAnalyzed: 5,
		FunnelCompleted:     true,
		Steps: []analyzer.StepResult{
			{Name: "Step with <brackets>", EventCount: 5, Percentage: 100.0},
		},
		DropOffs: []analyzer.DropOff{},
	}

	output, err := formatter.Format(result)
	if err != nil {
		t.Errorf("Format() unexpected error: %v", err)
		return
	}

	// Verify it's valid JSON despite special characters
	var parsed analyzer.FunnelResult
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("Format() output with special characters is not valid JSON: %v", err)
		return
	}

	if parsed.FunnelName != result.FunnelName {
		t.Errorf("JSON should preserve special characters in funnel name")
	}
}

package output

import (
	"encoding/json"
	"github.com/parfenovvs/loglion/internal/analyzer"
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

func TestTextFormatter_FormatFunnel_EmptyResult(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "Test Funnel",
		TotalEventsAnalyzed: 0,
		FunnelCompleted:     false,
		Steps:               []analyzer.StepResult{},
		DropOffs:            []analyzer.DropOff{},
	}

	output, err := formatter.FormatFunnel(result)
	if err != nil {
		t.Errorf("FormatFunnel() unexpected error: %v", err)
		return
	}

	expected := "❌ No events found\n"
	if output != expected {
		t.Errorf("FormatFunnel() = %q, want %q", output, expected)
	}
}

func TestTextFormatter_FormatFunnel_CompletedFunnel(t *testing.T) {
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

	output, err := formatter.FormatFunnel(result)
	if err != nil {
		t.Errorf("FormatFunnel() unexpected error: %v", err)
		return
	}

	// Check key components of the output
	if !strings.Contains(output, "✅ Funnel Analysis Complete") {
		t.Errorf("FormatFunnel() should contain success icon and title")
	}
	if !strings.Contains(output, "Funnel: User Registration") {
		t.Errorf("FormatFunnel() should contain funnel name")
	}
	if !strings.Contains(output, "Total Events Analyzed: 100") {
		t.Errorf("FormatFunnel() should contain total events")
	}
	if !strings.Contains(output, "Funnel Completed: Yes") {
		t.Errorf("FormatFunnel() should indicate funnel completion")
	}
	if !strings.Contains(output, "1. App Launch: 100 events (100.0%)") {
		t.Errorf("FormatFunnel() should contain step breakdown")
	}
	if !strings.Contains(output, "Drop-off Analysis:") {
		t.Errorf("FormatFunnel() should contain drop-off analysis section")
	}
	if !strings.Contains(output, "App Launch → Sign Up Click: 50 events lost (50.0% drop-off)") {
		t.Errorf("FormatFunnel() should contain drop-off details")
	}
}

func TestTextFormatter_FormatFunnel_IncompleteFunnel(t *testing.T) {
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

	output, err := formatter.FormatFunnel(result)
	if err != nil {
		t.Errorf("FormatFunnel() unexpected error: %v", err)
		return
	}

	if !strings.Contains(output, "❌ Funnel Analysis Complete") {
		t.Errorf("FormatFunnel() should contain failure icon for incomplete funnel")
	}
	if !strings.Contains(output, "Funnel Completed: No") {
		t.Errorf("FormatFunnel() should indicate funnel is not completed")
	}
}

func TestTextFormatter_FormatFunnel_NoDropOffs(t *testing.T) {
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

	output, err := formatter.FormatFunnel(result)
	if err != nil {
		t.Errorf("FormatFunnel() unexpected error: %v", err)
		return
	}

	if strings.Contains(output, "Drop-off Analysis:") {
		t.Errorf("FormatFunnel() should not contain drop-off section when no drop-offs exist")
	}
}

func TestJSONFormatter_FormatFunnel_ValidResult(t *testing.T) {
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

	output, err := formatter.FormatFunnel(result)
	if err != nil {
		t.Errorf("FormatFunnel() unexpected error: %v", err)
		return
	}

	// Verify it's valid JSON
	var parsed analyzer.FunnelResult
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("FormatFunnel() output is not valid JSON: %v", err)
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

func TestJSONFormatter_FormatFunnel_EmptyResult(t *testing.T) {
	formatter := &JSONFormatter{}
	result := &analyzer.FunnelResult{
		FunnelName:          "Empty Funnel",
		TotalEventsAnalyzed: 0,
		FunnelCompleted:     false,
		Steps:               []analyzer.StepResult{},
		DropOffs:            []analyzer.DropOff{},
	}

	output, err := formatter.FormatFunnel(result)
	if err != nil {
		t.Errorf("FormatFunnel() unexpected error: %v", err)
		return
	}

	// Verify it's valid JSON
	var parsed analyzer.FunnelResult
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("FormatFunnel() output is not valid JSON: %v", err)
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

func TestJSONFormatter_FormatFunnel_NilResult(t *testing.T) {
	formatter := &JSONFormatter{}

	// The JSONFormatter panics on nil input, which is expected behavior
	// since the function expects a valid FunnelResult pointer
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Format(nil) should panic")
		}
	}()

	formatter.FormatFunnel(nil)
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

			// Test that FormatFunnel method exists and doesn't panic on basic calls
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Formatter[%d] FormatFunnel() should not panic", i)
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

			_, err := formatter.FormatFunnel(result)
			if err != nil {
				t.Errorf("Formatter[%d] FormatFunnel() unexpected error: %v", i, err)
			}
		})
	}
}

func TestTextFormatter_FormatFunnel_SpecialCharacters(t *testing.T) {
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

	output, err := formatter.FormatFunnel(result)
	if err != nil {
		t.Errorf("FormatFunnel() unexpected error: %v", err)
		return
	}

	if !strings.Contains(output, "Test & Special \"Characters\"") {
		t.Errorf("FormatFunnel() should handle special characters in funnel name")
	}
	if !strings.Contains(output, "Step with <brackets>") {
		t.Errorf("FormatFunnel() should handle special characters in step names")
	}
}

func TestJSONFormatter_FormatFunnel_SpecialCharacters(t *testing.T) {
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

	output, err := formatter.FormatFunnel(result)
	if err != nil {
		t.Errorf("FormatFunnel() unexpected error: %v", err)
		return
	}

	// Verify it's valid JSON despite special characters
	var parsed analyzer.FunnelResult
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("FormatFunnel() output with special characters is not valid JSON: %v", err)
		return
	}

	if parsed.FunnelName != result.FunnelName {
		t.Errorf("JSON should preserve special characters in funnel name")
	}
}

// Count formatter tests

func TestTextFormatter_FormatCount_EmptyResult(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.CountResult{
		TotalEventsAnalyzed: 0,
		PatternCounts:       []analyzer.PatternCount{},
	}

	output, err := formatter.FormatCount(result)
	if err != nil {
		t.Errorf("FormatCount() unexpected error: %v", err)
		return
	}

	expected := "❌ No events found\n"
	if output != expected {
		t.Errorf("FormatCount() = %q, want %q", output, expected)
	}
}

func TestTextFormatter_FormatCount_ValidResult(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.CountResult{
		TotalEventsAnalyzed: 100,
		PatternCounts: []analyzer.PatternCount{
			{Pattern: "login", Count: 25},
			{Pattern: "logout", Count: 20},
			{Pattern: "error", Count: 5},
		},
	}

	output, err := formatter.FormatCount(result)
	if err != nil {
		t.Errorf("FormatCount() unexpected error: %v", err)
		return
	}

	// Check key components of the output
	if !strings.Contains(output, "📊 Event Count Analysis Complete") {
		t.Errorf("FormatCount() should contain analysis complete header")
	}
	if !strings.Contains(output, "Total Events Analyzed: 100") {
		t.Errorf("FormatCount() should contain total events")
	}
	if !strings.Contains(output, "Pattern Counts:") {
		t.Errorf("FormatCount() should contain pattern counts section")
	}
	if !strings.Contains(output, "1. login: 25 matches (25.0%)") {
		t.Errorf("FormatCount() should contain login count with percentage")
	}
	if !strings.Contains(output, "2. logout: 20 matches (20.0%)") {
		t.Errorf("FormatCount() should contain logout count with percentage")
	}
	if !strings.Contains(output, "3. error: 5 matches (5.0%)") {
		t.Errorf("FormatCount() should contain error count with percentage")
	}
	if !strings.Contains(output, "Total Matches: 50") {
		t.Errorf("FormatCount() should contain total matches summary")
	}
}

func TestTextFormatter_FormatCount_ZeroCounts(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.CountResult{
		TotalEventsAnalyzed: 50,
		PatternCounts: []analyzer.PatternCount{
			{Pattern: "login", Count: 0},
			{Pattern: "signup", Count: 0},
		},
	}

	output, err := formatter.FormatCount(result)
	if err != nil {
		t.Errorf("FormatCount() unexpected error: %v", err)
		return
	}

	if !strings.Contains(output, "1. login: 0 matches (0.0%)") {
		t.Errorf("FormatCount() should handle zero counts correctly")
	}
	if !strings.Contains(output, "2. signup: 0 matches (0.0%)") {
		t.Errorf("FormatCount() should handle zero counts correctly")
	}
	if !strings.Contains(output, "Total Matches: 0") {
		t.Errorf("FormatCount() should show zero total matches")
	}
}

func TestTextFormatter_FormatCount_SinglePattern(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.CountResult{
		TotalEventsAnalyzed: 10,
		PatternCounts: []analyzer.PatternCount{
			{Pattern: "error", Count: 3},
		},
	}

	output, err := formatter.FormatCount(result)
	if err != nil {
		t.Errorf("FormatCount() unexpected error: %v", err)
		return
	}

	if !strings.Contains(output, "1. error: 3 matches (30.0%)") {
		t.Errorf("FormatCount() should correctly calculate percentage for single pattern")
	}
	if !strings.Contains(output, "Total Matches: 3") {
		t.Errorf("FormatCount() should show correct total matches for single pattern")
	}
}

func TestJSONFormatter_FormatCount_ValidResult(t *testing.T) {
	formatter := &JSONFormatter{}
	result := &analyzer.CountResult{
		TotalEventsAnalyzed: 100,
		PatternCounts: []analyzer.PatternCount{
			{Pattern: "login", Count: 25},
			{Pattern: "logout", Count: 20},
		},
	}

	output, err := formatter.FormatCount(result)
	if err != nil {
		t.Errorf("FormatCount() unexpected error: %v", err)
		return
	}

	// Verify it's valid JSON
	var parsed analyzer.CountResult
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("FormatCount() output is not valid JSON: %v", err)
		return
	}

	// Verify content matches original
	if parsed.TotalEventsAnalyzed != result.TotalEventsAnalyzed {
		t.Errorf("JSON TotalEventsAnalyzed = %v, want %v", parsed.TotalEventsAnalyzed, result.TotalEventsAnalyzed)
	}
	if len(parsed.PatternCounts) != len(result.PatternCounts) {
		t.Errorf("JSON PatternCounts length = %v, want %v", len(parsed.PatternCounts), len(result.PatternCounts))
	}

	// Check individual pattern counts
	for i, patternCount := range result.PatternCounts {
		if i >= len(parsed.PatternCounts) {
			t.Errorf("Missing pattern count in JSON output at index %d", i)
			continue
		}
		if parsed.PatternCounts[i].Pattern != patternCount.Pattern {
			t.Errorf("JSON PatternCounts[%d].Pattern = %v, want %v", i, parsed.PatternCounts[i].Pattern, patternCount.Pattern)
		}
		if parsed.PatternCounts[i].Count != patternCount.Count {
			t.Errorf("JSON PatternCounts[%d].Count = %v, want %v", i, parsed.PatternCounts[i].Count, patternCount.Count)
		}
	}
}

func TestJSONFormatter_FormatCount_EmptyResult(t *testing.T) {
	formatter := &JSONFormatter{}
	result := &analyzer.CountResult{
		TotalEventsAnalyzed: 0,
		PatternCounts:       []analyzer.PatternCount{},
	}

	output, err := formatter.FormatCount(result)
	if err != nil {
		t.Errorf("FormatCount() unexpected error: %v", err)
		return
	}

	// Verify it's valid JSON
	var parsed analyzer.CountResult
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("FormatCount() output is not valid JSON: %v", err)
		return
	}

	if parsed.TotalEventsAnalyzed != 0 {
		t.Errorf("JSON TotalEventsAnalyzed = %v, want 0", parsed.TotalEventsAnalyzed)
	}
	if len(parsed.PatternCounts) != 0 {
		t.Errorf("JSON PatternCounts should be empty array")
	}
}

func TestJSONFormatter_FormatCount_NilResult(t *testing.T) {
	formatter := &JSONFormatter{}

	// The JSONFormatter panics on nil input, which is expected behavior
	// since the function expects a valid CountResult pointer
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("FormatCount(nil) should panic")
		}
	}()

	formatter.FormatCount(nil)
}

func TestTextFormatter_FormatCount_SpecialCharacters(t *testing.T) {
	formatter := &TextFormatter{}
	result := &analyzer.CountResult{
		TotalEventsAnalyzed: 5,
		PatternCounts: []analyzer.PatternCount{
			{Pattern: "pattern with \"quotes\"", Count: 2},
			{Pattern: "pattern_with_<brackets>", Count: 3},
		},
	}

	output, err := formatter.FormatCount(result)
	if err != nil {
		t.Errorf("FormatCount() unexpected error: %v", err)
		return
	}

	if !strings.Contains(output, "pattern with \"quotes\": 2 matches") {
		t.Errorf("FormatCount() should handle special characters in pattern names")
	}
	if !strings.Contains(output, "pattern_with_<brackets>: 3 matches") {
		t.Errorf("FormatCount() should handle special characters in pattern names")
	}
}

func TestJSONFormatter_FormatCount_SpecialCharacters(t *testing.T) {
	formatter := &JSONFormatter{}
	result := &analyzer.CountResult{
		TotalEventsAnalyzed: 5,
		PatternCounts: []analyzer.PatternCount{
			{Pattern: "pattern with \"quotes\"", Count: 2},
			{Pattern: "pattern_with_<brackets>", Count: 3},
		},
	}

	output, err := formatter.FormatCount(result)
	if err != nil {
		t.Errorf("FormatCount() unexpected error: %v", err)
		return
	}

	// Verify it's valid JSON despite special characters
	var parsed analyzer.CountResult
	err = json.Unmarshal([]byte(output), &parsed)
	if err != nil {
		t.Errorf("FormatCount() output with special characters is not valid JSON: %v", err)
		return
	}

	if parsed.PatternCounts[0].Pattern != result.PatternCounts[0].Pattern {
		t.Errorf("JSON should preserve special characters in pattern names")
	}
	if parsed.PatternCounts[1].Pattern != result.PatternCounts[1].Pattern {
		t.Errorf("JSON should preserve special characters in pattern names")
	}
}

func TestFormatter_Interface_FormatCount(t *testing.T) {
	// Test that formatters implement the Formatter interface with FormatCount method
	formatters := []Formatter{
		&TextFormatter{},
		&JSONFormatter{},
	}

	for i, formatter := range formatters {
		t.Run(reflect.TypeOf(formatter).String(), func(t *testing.T) {
			// Check that formatter implements Formatter interface
			var _ Formatter = formatter

			// Test that FormatCount method exists and doesn't panic on basic calls
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Formatter[%d] FormatCount() should not panic", i)
				}
			}()

			// Create a minimal result for testing
			result := &analyzer.CountResult{
				TotalEventsAnalyzed: 1,
				PatternCounts: []analyzer.PatternCount{
					{Pattern: "test", Count: 1},
				},
			}

			_, err := formatter.FormatCount(result)
			if err != nil {
				t.Errorf("Formatter[%d] FormatCount() unexpected error: %v", i, err)
			}
		})
	}
}

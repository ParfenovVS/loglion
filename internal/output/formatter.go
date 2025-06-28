package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"loglion/internal/analyzer"
)

type OutputFormat string

const (
	TextFormat OutputFormat = "text"
	JSONFormat OutputFormat = "json"
)

type Formatter interface {
	Format(result *analyzer.FunnelResult) (string, error)
}

func NewFormatter(format OutputFormat) Formatter {
	switch format {
	case JSONFormat:
		return &JSONFormatter{}
	default:
		return &TextFormatter{}
	}
}

type TextFormatter struct{}

func (f *TextFormatter) Format(result *analyzer.FunnelResult) (string, error) {
	var output strings.Builder
	
	if result.TotalSessions == 0 {
		output.WriteString("❌ No sessions found\n")
		return output.String(), nil
	}
	
	output.WriteString("✅ Funnel Analysis Complete\n\n")
	output.WriteString(fmt.Sprintf("Funnel: %s\n", result.FunnelName))
	output.WriteString(fmt.Sprintf("Total Sessions: %d\n", result.TotalSessions))
	output.WriteString(fmt.Sprintf("Completed Funnels: %d (%.1f%%)\n\n", 
		result.CompletedFunnels, result.CompletionRate*100))
	
	output.WriteString("Step Breakdown:\n")
	for i, step := range result.Steps {
		output.WriteString(fmt.Sprintf("%d. %s: %d/%d (%.1f%%)\n", 
			i+1, step.Name, step.Completed, result.TotalSessions, step.CompletionRate*100))
	}
	
	if len(result.Steps) > 1 {
		output.WriteString("\nDrop-off Analysis:\n")
		for i := 0; i < len(result.Steps)-1; i++ {
			current := result.Steps[i]
			next := result.Steps[i+1]
			dropOff := current.Completed - next.Completed
			if dropOff > 0 {
				output.WriteString(fmt.Sprintf("- %s → %s: %d session(s) lost\n", 
					current.Name, next.Name, dropOff))
			}
		}
	}
	
	return output.String(), nil
}

type JSONFormatter struct{}

func (f *JSONFormatter) Format(result *analyzer.FunnelResult) (string, error) {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return string(jsonData), nil
}
package output

import (
	"encoding/json"
	"fmt"
	"loglion/internal/analyzer"
	"strings"
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

	if result.TotalEventsAnalyzed == 0 {
		output.WriteString("❌ No events found\n")
		return output.String(), nil
	}

	// Choose status icon
	statusIcon := "✅"
	if !result.FunnelCompleted {
		statusIcon = "❌"
	}

	output.WriteString(fmt.Sprintf("%s Funnel Analysis Complete\n\n", statusIcon))
	output.WriteString(fmt.Sprintf("Funnel: %s\n", result.FunnelName))
	output.WriteString(fmt.Sprintf("Total Events Analyzed: %d\n", result.TotalEventsAnalyzed))

	if result.FunnelCompleted {
		output.WriteString("Funnel Completed: Yes\n\n")
	} else {
		output.WriteString("Funnel Completed: No\n\n")
	}

	output.WriteString("Step Breakdown:\n")
	for i, step := range result.Steps {
		output.WriteString(fmt.Sprintf("%d. %s: %d events (%.1f%%)\n",
			i+1, step.Name, step.EventCount, step.Percentage))
	}

	if len(result.DropOffs) > 0 {
		output.WriteString("\nDrop-off Analysis:\n")
		for _, dropOff := range result.DropOffs {
			output.WriteString(fmt.Sprintf("- %s → %s: %d events lost (%.1f%% drop-off)\n",
				dropOff.From, dropOff.To, dropOff.EventsLost, dropOff.DropOffRate))
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

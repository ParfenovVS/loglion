package output

import (
	"encoding/json"
	"fmt"
	"loglion/internal/analyzer"
	"strings"

	"github.com/sirupsen/logrus"
)

type OutputFormat string

const (
	TextFormat OutputFormat = "text"
	JSONFormat OutputFormat = "json"
)

type Formatter interface {
	FormatFunnel(result *analyzer.FunnelResult) (string, error)
	FormatCount(result *analyzer.CountResult) (string, error)
}

func NewFormatter(format OutputFormat) Formatter {
	logrus.WithField("format", format).Debug("Creating new output formatter")

	switch format {
	case JSONFormat:
		logrus.Debug("Using JSON formatter")
		return &JSONFormatter{}
	default:
		logrus.Debug("Using text formatter (default)")
		return &TextFormatter{}
	}
}

type TextFormatter struct{}

func (f *TextFormatter) FormatFunnel(result *analyzer.FunnelResult) (string, error) {
	logrus.WithFields(logrus.Fields{
		"funnel_name":      result.FunnelName,
		"total_events":     result.TotalEventsAnalyzed,
		"funnel_completed": result.FunnelCompleted,
		"steps_count":      len(result.Steps),
		"dropoffs_count":   len(result.DropOffs),
	}).Debug("Formatting funnel result as text")

	var output strings.Builder

	if result.TotalEventsAnalyzed == 0 {
		logrus.Debug("No events found, generating empty result message")
		output.WriteString("❌ No events found\n")
		return output.String(), nil
	}

	// Choose status icon
	statusIcon := "✅"
	if !result.FunnelCompleted {
		statusIcon = "❌"
	}
	logrus.WithField("status_icon", statusIcon).Debug("Selected status icon")

	output.WriteString(fmt.Sprintf("%s Funnel Analysis Complete\n\n", statusIcon))
	output.WriteString(fmt.Sprintf("Funnel: %s\n", result.FunnelName))
	output.WriteString(fmt.Sprintf("Total Events Analyzed: %d\n", result.TotalEventsAnalyzed))

	if result.FunnelCompleted {
		output.WriteString("Funnel Completed: Yes\n\n")
	} else {
		output.WriteString("Funnel Completed: No\n\n")
	}

	logrus.Debug("Formatting step breakdown section")
	output.WriteString("Step Breakdown:\n")
	for i, step := range result.Steps {
		logrus.WithFields(logrus.Fields{
			"step_index":  i + 1,
			"step_name":   step.Name,
			"event_count": step.EventCount,
			"percentage":  step.Percentage,
		}).Debug("Formatting step result")

		output.WriteString(fmt.Sprintf("%d. %s: %d events (%.1f%%)\n",
			i+1, step.Name, step.EventCount, step.Percentage))
	}

	if len(result.DropOffs) > 0 {
		logrus.Debug("Formatting drop-off analysis section")
		output.WriteString("\nDrop-off Analysis:\n")
		for _, dropOff := range result.DropOffs {
			logrus.WithFields(logrus.Fields{
				"from_step":     dropOff.From,
				"to_step":       dropOff.To,
				"events_lost":   dropOff.EventsLost,
				"drop_off_rate": dropOff.DropOffRate,
			}).Debug("Formatting drop-off result")

			output.WriteString(fmt.Sprintf("- %s → %s: %d events lost (%.1f%% drop-off)\n",
				dropOff.From, dropOff.To, dropOff.EventsLost, dropOff.DropOffRate))
		}
	}

	resultStr := output.String()
	logrus.WithField("output_length", len(resultStr)).Debug("Text formatting completed")
	return resultStr, nil
}

func (f *TextFormatter) FormatCount(result *analyzer.CountResult) (string, error) {
	logrus.WithFields(logrus.Fields{
		"total_events":    result.TotalEventsAnalyzed,
		"patterns_count":  len(result.PatternCounts),
	}).Debug("Formatting count result as text")

	var output strings.Builder

	if result.TotalEventsAnalyzed == 0 {
		logrus.Debug("No events found, generating empty result message")
		output.WriteString("❌ No events found\n")
		return output.String(), nil
	}

	output.WriteString("📊 Event Count Analysis Complete\n\n")
	output.WriteString(fmt.Sprintf("Total Events Analyzed: %d\n\n", result.TotalEventsAnalyzed))

	if len(result.PatternCounts) > 0 {
		logrus.Debug("Formatting pattern counts section")
		output.WriteString("Pattern Counts:\n")
		totalMatches := 0
		for i, patternCount := range result.PatternCounts {
			logrus.WithFields(logrus.Fields{
				"pattern_index": i + 1,
				"pattern_name":  patternCount.Pattern,
				"count":         patternCount.Count,
			}).Debug("Formatting pattern count result")

			percentage := 0.0
			if result.TotalEventsAnalyzed > 0 {
				percentage = float64(patternCount.Count) / float64(result.TotalEventsAnalyzed) * 100.0
			}

			output.WriteString(fmt.Sprintf("%d. %s: %d matches (%.1f%%)\n",
				i+1, patternCount.Pattern, patternCount.Count, percentage))
			totalMatches += patternCount.Count
		}

		output.WriteString(fmt.Sprintf("\nTotal Matches: %d\n", totalMatches))
	}

	resultStr := output.String()
	logrus.WithField("output_length", len(resultStr)).Debug("Text count formatting completed")
	return resultStr, nil
}

type JSONFormatter struct{}

func (f *JSONFormatter) FormatFunnel(result *analyzer.FunnelResult) (string, error) {
	logrus.WithFields(logrus.Fields{
		"funnel_name":      result.FunnelName,
		"total_events":     result.TotalEventsAnalyzed,
		"funnel_completed": result.FunnelCompleted,
		"steps_count":      len(result.Steps),
		"dropoffs_count":   len(result.DropOffs),
	}).Debug("Formatting funnel result as JSON")

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal funnel result to JSON")
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	logrus.WithField("json_length", len(jsonData)).Debug("JSON formatting completed")
	return string(jsonData), nil
}

func (f *JSONFormatter) FormatCount(result *analyzer.CountResult) (string, error) {
	logrus.WithFields(logrus.Fields{
		"total_events":   result.TotalEventsAnalyzed,
		"patterns_count": len(result.PatternCounts),
	}).Debug("Formatting count result as JSON")

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal count result to JSON")
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	logrus.WithField("json_length", len(jsonData)).Debug("JSON count formatting completed")
	return string(jsonData), nil
}

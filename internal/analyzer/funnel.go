package analyzer

import (
	"loglion/internal/config"
	"loglion/internal/parser"
	"regexp"

	"github.com/sirupsen/logrus"
)

type FunnelAnalyzer struct {
	config *config.Config
}

type FunnelResult struct {
	FunnelName          string       `json:"funnel_name"`
	TotalEventsAnalyzed int          `json:"total_events_analyzed"`
	FunnelCompleted     bool         `json:"funnel_completed"`
	Steps               []StepResult `json:"steps"`
	DropOffs            []DropOff    `json:"drop_offs"`
}

type StepResult struct {
	Name       string  `json:"name"`
	EventCount int     `json:"event_count"`
	Percentage float64 `json:"percentage"`
}

type DropOff struct {
	From        string  `json:"from"`
	To          string  `json:"to"`
	EventsLost  int     `json:"events_lost"`
	DropOffRate float64 `json:"drop_off_rate"`
}

func NewFunnelAnalyzer(cfg *config.Config) *FunnelAnalyzer {
	logrus.WithFields(logrus.Fields{
		"funnel_name": cfg.Funnel.Name,
		"step_count":  len(cfg.Funnel.Steps),
	}).Debug("Creating new funnel analyzer")

	return &FunnelAnalyzer{
		config: cfg,
	}
}

func (fa *FunnelAnalyzer) AnalyzeFunnel(entries []*parser.LogEntry, max int) *FunnelResult {
	logrus.WithFields(logrus.Fields{
		"funnel_name": fa.config.Funnel.Name,
		"entry_count": len(entries),
		"max":         max,
	}).Info("Starting funnel analysis")

	if len(entries) == 0 {
		logrus.Warn("No log entries provided for analysis")
		return &FunnelResult{
			FunnelName:          fa.config.Funnel.Name,
			TotalEventsAnalyzed: 0,
			FunnelCompleted:     false,
			Steps:               []StepResult{},
			DropOffs:            []DropOff{},
		}
	}

	stepResults := make([]StepResult, len(fa.config.Funnel.Steps))
	stepCounts := make([]int, len(fa.config.Funnel.Steps))

	// Initialize step results
	for i, step := range fa.config.Funnel.Steps {
		stepResults[i] = StepResult{
			Name:       step.Name,
			EventCount: 0,
			Percentage: 0.0,
		}
		logrus.WithFields(logrus.Fields{
			"step_index": i + 1,
			"step_name":  step.Name,
			"pattern":    step.EventPattern,
		}).Debug("Initialized funnel step")
	}

	var matchedEvents int
	var currentStep int
	var conversionsFound int

	if max == 0 {
		// Mode 1: Track sequential funnel progression through the entire log
		logrus.Debug("Mode 1: Tracking sequential funnel progression")
		currentStep = 0

		for entryIndex, entry := range entries {
			// Check if current entry matches the expected next step
			if currentStep < len(fa.config.Funnel.Steps) {
				step := fa.config.Funnel.Steps[currentStep]
				if fa.eventMatchesStep(entry, step) {
					stepCounts[currentStep]++
					matchedEvents++
					currentStep++

					logrus.WithFields(logrus.Fields{
						"entry_index": entryIndex + 1,
						"step_index":  currentStep,
						"step_name":   step.Name,
						"timestamp":   entry.Timestamp,
						"message":     entry.Message,
					}).Debug("Event matched funnel step")

					// Check if funnel was completed
					if currentStep >= len(fa.config.Funnel.Steps) {
						conversionsFound++
						logrus.WithField("conversions_total", conversionsFound).Debug("Funnel completed")
						// Reset to look for additional complete funnels
						currentStep = 0
					}
				}
			}
		}
	} else {
		// Mode 2: Track complete funnel conversions, stop after 'max' conversions
		logrus.WithField("target_conversions", max).Debug("Mode 2: Tracking complete funnel conversions")
		conversionsFound = 0
		currentStep = 0

		for entryIndex, entry := range entries {
			if conversionsFound >= max {
				logrus.WithField("conversions_found", conversionsFound).Debug("Target conversions reached, stopping analysis")
				break
			}

			if currentStep >= len(fa.config.Funnel.Steps) {
				logrus.Debug("Funnel completed, resetting for next conversion")
				conversionsFound++
				currentStep = 0 // Reset for next conversion
				if conversionsFound >= max {
					break
				}
			}

			step := fa.config.Funnel.Steps[currentStep]
			if fa.eventMatchesStep(entry, step) {
				stepCounts[currentStep]++
				matchedEvents++
				logrus.WithFields(logrus.Fields{
					"entry_index":        entryIndex + 1,
					"step_index":         currentStep + 1,
					"step_name":          step.Name,
					"timestamp":          entry.Timestamp,
					"message":            entry.Message,
					"conversions_so_far": conversionsFound,
				}).Debug("Event matched funnel step")
				currentStep++
			}
		}

		// Check if funnel was completed at the end
		if currentStep >= len(fa.config.Funnel.Steps) {
			logrus.Debug("Funnel completed at end of log")
			conversionsFound++
		}
	}

	logrus.WithFields(logrus.Fields{
		"total_entries":   len(entries),
		"matched_events":  matchedEvents,
		"completed_steps": currentStep,
		"total_steps":     len(fa.config.Funnel.Steps),
		"mode":            map[bool]string{true: "count_all", false: "track_conversions"}[max == 0],
	}).Info("Funnel analysis completed")

	// Calculate percentages based on first step
	logrus.Debug("Calculating conversion percentages")
	var baseCount int
	if len(stepCounts) > 0 && stepCounts[0] > 0 {
		baseCount = stepCounts[0]
	}

	for i, count := range stepCounts {
		stepResults[i].EventCount = count
		if baseCount > 0 {
			stepResults[i].Percentage = float64(count) / float64(baseCount) * 100.0
		}
		logrus.WithFields(logrus.Fields{
			"step_name":   stepResults[i].Name,
			"event_count": count,
			"percentage":  stepResults[i].Percentage,
		}).Debug("Step conversion calculated")
	}

	// Calculate drop-offs
	logrus.Debug("Calculating drop-off rates")
	dropOffs := []DropOff{}
	for i := 0; i < len(stepCounts)-1; i++ {
		if stepCounts[i] > 0 {
			lost := stepCounts[i] - stepCounts[i+1]
			dropOffRate := float64(lost) / float64(stepCounts[i]) * 100.0

			dropOff := DropOff{
				From:        fa.config.Funnel.Steps[i].Name,
				To:          fa.config.Funnel.Steps[i+1].Name,
				EventsLost:  lost,
				DropOffRate: dropOffRate,
			}

			dropOffs = append(dropOffs, dropOff)

			logrus.WithFields(logrus.Fields{
				"from_step":     dropOff.From,
				"to_step":       dropOff.To,
				"events_lost":   lost,
				"drop_off_rate": dropOffRate,
			}).Debug("Drop-off calculated")
		}
	}

	// Determine if funnel was completed
	var funnelCompleted bool
	if max == 0 {
		// In Mode 1, check if we found any complete conversions
		funnelCompleted = conversionsFound > 0
	} else {
		// In Mode 2, check if we found any complete conversions
		funnelCompleted = conversionsFound > 0
	}
	logrus.WithField("funnel_completed", funnelCompleted).Debug("Funnel completion status determined")

	result := &FunnelResult{
		FunnelName:          fa.config.Funnel.Name,
		TotalEventsAnalyzed: len(entries),
		FunnelCompleted:     funnelCompleted,
		Steps:               stepResults,
		DropOffs:            dropOffs,
	}

	logrus.WithFields(logrus.Fields{
		"funnel_name":      result.FunnelName,
		"total_events":     result.TotalEventsAnalyzed,
		"funnel_completed": result.FunnelCompleted,
		"steps_analyzed":   len(result.Steps),
		"drop_offs_found":  len(result.DropOffs),
	}).Info("Funnel analysis completed")

	return result
}

func (fa *FunnelAnalyzer) eventMatchesStep(entry *parser.LogEntry, step config.Step) bool {
	logrus.WithFields(logrus.Fields{
		"step_name":      step.Name,
		"step_pattern":   step.EventPattern,
		"entry_message":  entry.Message,
		"has_event_data": entry.EventData != nil,
	}).Debug("Checking if event matches step")

	// Compile regex pattern
	eventRegex, err := regexp.Compile(step.EventPattern)
	if err != nil {
		logrus.WithError(err).WithField("step_pattern", step.EventPattern).Error("Failed to compile step regex pattern")
		return false
	}

	// If we have structured event data, match against the "event" field
	if entry.EventData != nil {
		if eventValue, exists := entry.EventData["event"]; exists {
			if eventStr, ok := eventValue.(string); ok {
				logrus.WithFields(logrus.Fields{
					"event_str": eventStr,
					"pattern":   step.EventPattern,
				}).Debug("Matching against structured event field")

				if !eventRegex.MatchString(eventStr) {
					logrus.Debug("Event string does not match pattern")
					return false
				}
			} else {
				logrus.Debug("Event field is not a string, failing match")
				return false
			}
		} else {
			// Fall back to matching the raw message if no "event" field
			logrus.Debug("No 'event' field found, falling back to raw message matching")
			if !eventRegex.MatchString(entry.Message) {
				logrus.Debug("Raw message does not match pattern")
				return false
			}
		}
	} else {
		// No structured data, match against raw message
		logrus.Debug("No structured data, matching against raw message")
		if !eventRegex.MatchString(entry.Message) {
			logrus.Debug("Raw message does not match pattern")
			return false
		}
		hasRequiredProps := len(step.RequiredProperties) == 0
		logrus.WithField("has_required_props", hasRequiredProps).Debug("No structured data available for property checking")
		return hasRequiredProps
	}

	// Check required properties
	logrus.WithField("required_props_count", len(step.RequiredProperties)).Debug("Checking required properties")
	return fa.checkRequiredProperties(entry.EventData, step.RequiredProperties)
}

func (fa *FunnelAnalyzer) checkRequiredProperties(eventData map[string]interface{}, requiredProps map[string]string) bool {
	logrus.WithField("properties_to_check", len(requiredProps)).Debug("Starting required properties validation")

	for key, pattern := range requiredProps {
		logrus.WithFields(logrus.Fields{
			"property_key": key,
			"pattern":      pattern,
		}).Debug("Checking required property")

		value, exists := eventData[key]
		if !exists {
			logrus.WithField("property_key", key).Debug("Required property not found in event data")
			return false
		}

		valueStr, ok := value.(string)
		if !ok {
			logrus.WithFields(logrus.Fields{
				"property_key": key,
				"value_type":   typeof(value),
			}).Debug("Property value is not a string")
			return false
		}

		matched, err := regexp.MatchString(pattern, valueStr)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"property_key": key,
				"pattern":      pattern,
			}).Error("Failed to compile property pattern regex")
			return false
		}

		if !matched {
			logrus.WithFields(logrus.Fields{
				"property_key":   key,
				"property_value": valueStr,
				"pattern":        pattern,
			}).Debug("Property value does not match required pattern")
			return false
		}

		logrus.WithFields(logrus.Fields{
			"property_key":   key,
			"property_value": valueStr,
		}).Debug("Property validation passed")
	}

	logrus.Debug("All required properties validated successfully")
	return true
}

// Helper function to get type name for logging
func typeof(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case int:
		return "int"
	case float64:
		return "float64"
	case bool:
		return "bool"
	case map[string]interface{}:
		return "map"
	case []interface{}:
		return "array"
	default:
		return "unknown"
	}
}

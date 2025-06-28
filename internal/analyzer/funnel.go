package analyzer

import (
	"loglion/internal/config"
	"loglion/internal/parser"
	"regexp"
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
	return &FunnelAnalyzer{
		config: cfg,
	}
}

func (fa *FunnelAnalyzer) AnalyzeFunnel(entries []*parser.LogEntry) *FunnelResult {
	if len(entries) == 0 {
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
	}

	// Track funnel progression chronologically
	currentStep := 0
	for _, entry := range entries {
		if currentStep >= len(fa.config.Funnel.Steps) {
			break // Funnel completed
		}

		step := fa.config.Funnel.Steps[currentStep]
		if fa.eventMatchesStep(entry, step) {
			stepCounts[currentStep]++
			currentStep++
		}
	}

	// Calculate percentages based on first step
	var baseCount int
	if len(stepCounts) > 0 && stepCounts[0] > 0 {
		baseCount = stepCounts[0]
	}

	for i, count := range stepCounts {
		stepResults[i].EventCount = count
		if baseCount > 0 {
			stepResults[i].Percentage = float64(count) / float64(baseCount) * 100.0
		}
	}

	// Calculate drop-offs
	dropOffs := []DropOff{}
	for i := 0; i < len(stepCounts)-1; i++ {
		if stepCounts[i] > 0 {
			lost := stepCounts[i] - stepCounts[i+1]
			dropOffRate := float64(lost) / float64(stepCounts[i]) * 100.0

			dropOffs = append(dropOffs, DropOff{
				From:        fa.config.Funnel.Steps[i].Name,
				To:          fa.config.Funnel.Steps[i+1].Name,
				EventsLost:  lost,
				DropOffRate: dropOffRate,
			})
		}
	}

	// Determine if funnel was completed
	funnelCompleted := currentStep >= len(fa.config.Funnel.Steps)

	return &FunnelResult{
		FunnelName:          fa.config.Funnel.Name,
		TotalEventsAnalyzed: len(entries),
		FunnelCompleted:     funnelCompleted,
		Steps:               stepResults,
		DropOffs:            dropOffs,
	}
}

func (fa *FunnelAnalyzer) eventMatchesStep(entry *parser.LogEntry, step config.Step) bool {
	// Compile regex pattern
	eventRegex, err := regexp.Compile(step.EventPattern)
	if err != nil {
		return false
	}

	// If we have structured event data, match against the "event" field
	if entry.EventData != nil {
		if eventValue, exists := entry.EventData["event"]; exists {
			if eventStr, ok := eventValue.(string); ok {
				if !eventRegex.MatchString(eventStr) {
					return false
				}
			} else {
				return false
			}
		} else {
			// Fall back to matching the raw message if no "event" field
			if !eventRegex.MatchString(entry.Message) {
				return false
			}
		}
	} else {
		// No structured data, match against raw message
		if !eventRegex.MatchString(entry.Message) {
			return false
		}
		return len(step.RequiredProperties) == 0
	}

	// Check required properties
	return fa.checkRequiredProperties(entry.EventData, step.RequiredProperties)
}

func (fa *FunnelAnalyzer) checkRequiredProperties(eventData map[string]interface{}, requiredProps map[string]string) bool {
	for key, pattern := range requiredProps {
		value, exists := eventData[key]
		if !exists {
			return false
		}

		valueStr, ok := value.(string)
		if !ok {
			return false
		}

		matched, err := regexp.MatchString(pattern, valueStr)
		if err != nil || !matched {
			return false
		}
	}

	return true
}

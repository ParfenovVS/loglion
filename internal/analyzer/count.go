package analyzer

import (
	"loglion/internal/parser"
	"regexp"

	"github.com/sirupsen/logrus"
)

type CountAnalyzer struct {
	patterns []EventPattern
}

type EventPattern struct {
	Name    string
	Pattern string
	Regex   *regexp.Regexp
}

type CountResult struct {
	TotalEventsAnalyzed int           `json:"total_events_analyzed"`
	PatternCounts       []PatternCount `json:"pattern_counts"`
}

type PatternCount struct {
	Pattern string `json:"pattern"`
	Count   int    `json:"count"`
}

func NewCountAnalyzer(eventPatterns []string) (*CountAnalyzer, error) {
	logrus.WithField("pattern_count", len(eventPatterns)).Debug("Creating new count analyzer")

	patterns := make([]EventPattern, len(eventPatterns))
	for i, patternStr := range eventPatterns {
		regex, err := regexp.Compile(patternStr)
		if err != nil {
			logrus.WithError(err).WithField("pattern", patternStr).Error("Failed to compile event pattern regex")
			return nil, err
		}

		patterns[i] = EventPattern{
			Name:    patternStr,
			Pattern: patternStr,
			Regex:   regex,
		}

		logrus.WithFields(logrus.Fields{
			"pattern_index": i + 1,
			"pattern":       patternStr,
		}).Debug("Compiled event pattern")
	}

	return &CountAnalyzer{
		patterns: patterns,
	}, nil
}

func (ca *CountAnalyzer) AnalyzeCount(entries []*parser.LogEntry) *CountResult {
	logrus.WithFields(logrus.Fields{
		"entry_count":   len(entries),
		"pattern_count": len(ca.patterns),
	}).Info("Starting count analysis")

	if len(entries) == 0 {
		logrus.Warn("No log entries provided for analysis")
		return &CountResult{
			TotalEventsAnalyzed: 0,
			PatternCounts:       []PatternCount{},
		}
	}

	patternCounts := make([]PatternCount, len(ca.patterns))
	counts := make([]int, len(ca.patterns))

	// Initialize pattern counts
	for i, pattern := range ca.patterns {
		patternCounts[i] = PatternCount{
			Pattern: pattern.Name,
			Count:   0,
		}
		logrus.WithFields(logrus.Fields{
			"pattern_index": i + 1,
			"pattern_name":  pattern.Name,
		}).Debug("Initialized pattern count")
	}

	// Count matches for each entry
	for entryIndex, entry := range entries {
		for patternIndex, pattern := range ca.patterns {
			if ca.eventMatchesPattern(entry, pattern) {
				counts[patternIndex]++
				logrus.WithFields(logrus.Fields{
					"entry_index":   entryIndex + 1,
					"pattern_index": patternIndex + 1,
					"pattern_name":  pattern.Name,
					"timestamp":     entry.Timestamp,
					"message":       entry.Message,
				}).Debug("Event matched pattern")
			}
		}
	}

	// Update pattern counts with final results
	for i, count := range counts {
		patternCounts[i].Count = count
		logrus.WithFields(logrus.Fields{
			"pattern_name": patternCounts[i].Pattern,
			"count":        count,
		}).Debug("Pattern count finalized")
	}

	logrus.WithFields(logrus.Fields{
		"total_entries":     len(entries),
		"patterns_checked":  len(ca.patterns),
	}).Info("Count analysis completed")

	result := &CountResult{
		TotalEventsAnalyzed: len(entries),
		PatternCounts:       patternCounts,
	}

	return result
}

func (ca *CountAnalyzer) eventMatchesPattern(entry *parser.LogEntry, pattern EventPattern) bool {
	logrus.WithFields(logrus.Fields{
		"pattern_name":   pattern.Name,
		"entry_message":  entry.Message,
		"has_event_data": entry.EventData != nil,
	}).Debug("Checking if event matches pattern")

	// If we have structured event data, match against the "event" field
	if entry.EventData != nil {
		if eventValue, exists := entry.EventData["event"]; exists {
			if eventStr, ok := eventValue.(string); ok {
				logrus.WithFields(logrus.Fields{
					"event_str": eventStr,
					"pattern":   pattern.Pattern,
				}).Debug("Matching against structured event field")

				matched := pattern.Regex.MatchString(eventStr)
				logrus.WithField("matched", matched).Debug("Structured event match result")
				return matched
			} else {
				logrus.Debug("Event field is not a string, failing match")
				return false
			}
		} else {
			// Fall back to matching the raw message if no "event" field
			logrus.Debug("No 'event' field found, falling back to raw message matching")
			matched := pattern.Regex.MatchString(entry.Message)
			logrus.WithField("matched", matched).Debug("Raw message match result")
			return matched
		}
	} else {
		// No structured data, match against raw message
		logrus.Debug("No structured data, matching against raw message")
		matched := pattern.Regex.MatchString(entry.Message)
		logrus.WithField("matched", matched).Debug("Raw message match result")
		return matched
	}
}
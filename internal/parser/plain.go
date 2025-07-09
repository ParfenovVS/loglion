package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type PlainParser struct {
	timestampFormat string
	eventRegex      *regexp.Regexp
	jsonExtraction  bool
	logLineRegex    *regexp.Regexp
}

func NewPlainParser() *PlainParser {
	return NewPlainParserWithConfig("", `^(.*)$`, false, `^(.*)$`)
}

func NewPlainParserWithConfig(timestampFormat, eventRegexPattern string, jsonExtraction bool, logLineRegexPattern string) *PlainParser {
	logrus.WithFields(logrus.Fields{
		"timestamp_format":       timestampFormat,
		"event_regex_pattern":    eventRegexPattern,
		"json_extraction":        jsonExtraction,
		"log_line_regex_pattern": logLineRegexPattern,
	}).Debug("Creating new Plain parser")

	// Default regex patterns if empty
	if eventRegexPattern == "" {
		eventRegexPattern = `^(.*)$`
		logrus.Debug("Using default event regex pattern")
	}

	if logLineRegexPattern == "" {
		logLineRegexPattern = `^(.*)$`
		logrus.Debug("Using default log line regex pattern")
	}

	// Compile event regex
	logrus.WithField("pattern", eventRegexPattern).Debug("Compiling event regex")
	eventRegex := regexp.MustCompile(eventRegexPattern)

	// Compile log line regex
	logrus.WithField("pattern", logLineRegexPattern).Debug("Compiling log line regex")
	logLineRegex := regexp.MustCompile(logLineRegexPattern)

	parser := &PlainParser{
		timestampFormat: timestampFormat,
		eventRegex:      eventRegex,
		jsonExtraction:  jsonExtraction,
		logLineRegex:    logLineRegex,
	}

	logrus.Debug("Plain parser created successfully")
	return parser
}

func (p *PlainParser) Parse(logLine string) (*LogEntry, error) {
	logrus.WithField("log_line", logLine).Debug("Parsing Plain log line")

	// Check for empty lines
	trimmedLine := strings.TrimSpace(logLine)
	if trimmedLine == "" {
		logrus.WithField("log_line", logLine).Debug("Empty log line")
		return nil, fmt.Errorf("empty log line")
	}

	// Use regex to parse the log line
	matches := p.logLineRegex.FindStringSubmatch(trimmedLine)
	if len(matches) == 0 {
		logrus.WithField("log_line", logLine).Debug("Log line does not match expected format")
		return nil, fmt.Errorf("invalid log line format: %s", logLine)
	}

	// Initialize entry with defaults
	entry := &LogEntry{
		Timestamp: time.Time{}, // Zero time if no timestamp
		Level:     "",
		Tag:       "",
		PID:       0,
		TID:       0,
		Message:   "",
	}

	// Extract fields based on available regex groups
	// Groups are in order: timestamp, pid, tid, level, tag, message
	if len(matches) > 1 && matches[1] != "" && p.timestampFormat != "" {
		// Try to parse timestamp if format is provided
		if timestamp, err := time.Parse(p.timestampFormat, matches[1]); err == nil {
			entry.Timestamp = timestamp
			logrus.WithField("timestamp", timestamp).Debug("Parsed timestamp")
		} else {
			logrus.WithError(err).WithField("timestamp_str", matches[1]).Debug("Failed to parse timestamp, using as message")
			entry.Message = matches[1]
		}
	}

	if len(matches) > 2 && matches[2] != "" {
		// Try to parse PID
		if pid, err := strconv.Atoi(matches[2]); err == nil {
			entry.PID = pid
		} else {
			logrus.WithError(err).WithField("pid_str", matches[2]).Debug("Failed to parse PID")
		}
	}

	if len(matches) > 3 && matches[3] != "" {
		// Try to parse TID
		if tid, err := strconv.Atoi(matches[3]); err == nil {
			entry.TID = tid
		} else {
			logrus.WithError(err).WithField("tid_str", matches[3]).Debug("Failed to parse TID")
		}
	}

	if len(matches) > 4 && matches[4] != "" {
		entry.Level = matches[4]
	}

	if len(matches) > 5 && matches[5] != "" {
		entry.Tag = matches[5]
	}

	if len(matches) > 6 && matches[6] != "" {
		entry.Message = matches[6]
	} else if len(matches) > 1 && entry.Message == "" {
		// If no specific message group, use the last captured group
		entry.Message = matches[len(matches)-1]
	}

	logrus.WithFields(logrus.Fields{
		"timestamp": entry.Timestamp,
		"pid":       entry.PID,
		"tid":       entry.TID,
		"level":     entry.Level,
		"tag":       entry.Tag,
		"message":   entry.Message,
	}).Debug("Extracted log line components")

	// Try to extract JSON data if enabled
	if p.jsonExtraction {
		logrus.Debug("Attempting to extract event data from log entry")
		p.extractEventData(entry, logLine)
	}

	logrus.WithFields(logrus.Fields{
		"timestamp": entry.Timestamp,
		"tag":       entry.Tag,
		"has_event": entry.EventData != nil,
	}).Debug("Log entry parsed successfully")

	return entry, nil
}

// extractEventData attempts to extract JSON event data from the log entry
func (p *PlainParser) extractEventData(entry *LogEntry, logLine string) {
	// First try the event regex pattern
	if p.eventRegex != nil {
		logrus.Debug("Trying to extract event data using regex pattern")
		matches := p.eventRegex.FindStringSubmatch(logLine)
		if len(matches) > 1 {
			jsonStr := strings.TrimSpace(matches[1])
			logrus.WithField("json_candidate", jsonStr).Debug("Found regex match, attempting JSON parse")
			if p.tryParseJSON(entry, jsonStr) {
				logrus.Debug("Successfully extracted event data using regex pattern")
				return
			}
		}
	}

	// Fallback: try to parse the message directly as JSON
	logrus.Debug("Fallback: trying to parse message directly as JSON")
	if p.tryParseJSON(entry, entry.Message) {
		logrus.Debug("Successfully extracted event data from message")
	} else {
		logrus.Debug("No valid JSON event data found")
	}
}

// tryParseJSON attempts to parse a string as JSON and populate EventData
func (p *PlainParser) tryParseJSON(entry *LogEntry, jsonStr string) bool {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &eventData); err == nil {
		entry.EventData = eventData
		logrus.WithField("event_keys", getMapKeysPlain(eventData)).Debug("JSON parsed successfully")
		return true
	}
	logrus.WithField("json_str", jsonStr).Debug("Failed to parse as JSON")
	return false
}

// Helper function to get map keys for logging
func getMapKeysPlain(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (p *PlainParser) ParseFile(filepath string) ([]*LogEntry, error) {
	logrus.WithField("filepath", filepath).Info("Starting to parse log file")

	file, err := os.Open(filepath)
	if err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to open log file")
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	return p.ParseReader(file)
}

func (p *PlainParser) ParseReader(reader io.Reader) ([]*LogEntry, error) {
	var entries []*LogEntry
	scanner := bufio.NewScanner(reader)
	lineCount := 0
	parsedCount := 0
	skippedCount := 0

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		entry, err := p.Parse(line)
		if err != nil {
			skippedCount++
			logrus.WithError(err).WithFields(logrus.Fields{
				"line_number": lineCount,
				"line":        line,
			}).Debug("Failed to parse log line, skipping")
			continue
		}

		entries = append(entries, entry)
		parsedCount++
	}

	if err := scanner.Err(); err != nil {
		logrus.WithError(err).Error("Error reading from reader")
		return nil, fmt.Errorf("error reading from reader: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"total_lines":    lineCount,
		"parsed_entries": parsedCount,
		"skipped_lines":  skippedCount,
	}).Info("Log parsing from reader completed")

	return entries, nil
}

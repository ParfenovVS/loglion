package parser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type AndroidParser struct {
	timestampFormat string
	eventRegex      *regexp.Regexp
	jsonExtraction  bool
	logLineRegex    *regexp.Regexp
}

func NewAndroidParser() *AndroidParser {
	return NewAndroidParserWithConfig("01-02 15:04:05.000", `.*Analytics.*: (.*)`, true)
}

func NewAndroidParserWithConfig(timestampFormat, eventRegexPattern string, jsonExtraction bool) *AndroidParser {
	logrus.WithFields(logrus.Fields{
		"timestamp_format":    timestampFormat,
		"event_regex_pattern": eventRegexPattern,
		"json_extraction":     jsonExtraction,
	}).Debug("Creating new Android parser")

	// Default regex if empty
	if eventRegexPattern == "" {
		eventRegexPattern = `.*Analytics.*: (.*)`
		logrus.Debug("Using default event regex pattern")
	}

	// Default timestamp format if empty
	if timestampFormat == "" {
		timestampFormat = "01-02 15:04:05.000"
		logrus.Debug("Using default timestamp format")
	}

	// Compile event regex
	logrus.WithField("pattern", eventRegexPattern).Debug("Compiling event regex")
	eventRegex := regexp.MustCompile(eventRegexPattern)

	// Regex to parse the full logcat line format
	logLineRegex := regexp.MustCompile(`^(\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3})\s+(\d+)\s+(\d+)\s+([VDIWEFS])\s+([^:]+):\s*(.*)$`)

	parser := &AndroidParser{
		timestampFormat: timestampFormat,
		eventRegex:      eventRegex,
		jsonExtraction:  jsonExtraction,
		logLineRegex:    logLineRegex,
	}

	logrus.Debug("Android parser created successfully")
	return parser
}

func (p *AndroidParser) Parse(logLine string) (*LogEntry, error) {
	logrus.WithField("log_line", logLine).Debug("Parsing Android log line")

	// Use regex to parse the logcat line
	matches := p.logLineRegex.FindStringSubmatch(strings.TrimSpace(logLine))
	if len(matches) != 7 {
		logrus.WithField("log_line", logLine).Debug("Log line does not match expected format")
		return nil, fmt.Errorf("invalid log line format: %s", logLine)
	}

	// Extract components from regex groups
	timestampStr := matches[1]
	pidStr := matches[2]
	tidStr := matches[3]
	level := matches[4]
	tag := matches[5]
	message := matches[6]

	logrus.WithFields(logrus.Fields{
		"timestamp": timestampStr,
		"pid":       pidStr,
		"tid":       tidStr,
		"level":     level,
		"tag":       tag,
		"message":   message,
	}).Debug("Extracted log line components")

	// Parse timestamp
	timestamp, err := time.Parse(p.timestampFormat, timestampStr)
	if err != nil {
		logrus.WithError(err).WithField("timestamp_str", timestampStr).Error("Failed to parse timestamp")
		return nil, fmt.Errorf("failed to parse timestamp '%s': %w", timestampStr, err)
	}

	// Parse PID and TID
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		logrus.WithError(err).WithField("pid_str", pidStr).Error("Failed to parse PID")
		return nil, fmt.Errorf("failed to parse PID '%s': %w", pidStr, err)
	}

	tid, err := strconv.Atoi(tidStr)
	if err != nil {
		logrus.WithError(err).WithField("tid_str", tidStr).Error("Failed to parse TID")
		return nil, fmt.Errorf("failed to parse TID '%s': %w", tidStr, err)
	}

	entry := &LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Tag:       tag,
		PID:       pid,
		TID:       tid,
		Message:   message,
	}

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
func (p *AndroidParser) extractEventData(entry *LogEntry, logLine string) {
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
func (p *AndroidParser) tryParseJSON(entry *LogEntry, jsonStr string) bool {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &eventData); err == nil {
		entry.EventData = eventData
		logrus.WithField("event_keys", getMapKeys(eventData)).Debug("JSON parsed successfully")
		return true
	}
	logrus.WithField("json_str", jsonStr).Debug("Failed to parse as JSON")
	return false
}

// Helper function to get map keys for logging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (p *AndroidParser) ParseFile(filepath string) ([]*LogEntry, error) {
	logrus.WithField("filepath", filepath).Info("Starting to parse log file")

	file, err := os.Open(filepath)
	if err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to open log file")
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var entries []*LogEntry
	scanner := bufio.NewScanner(file)
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
		logrus.WithError(err).WithField("filepath", filepath).Error("Error reading log file")
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"filepath":       filepath,
		"total_lines":    lineCount,
		"parsed_entries": parsedCount,
		"skipped_lines":  skippedCount,
	}).Info("Log file parsing completed")

	return entries, nil
}

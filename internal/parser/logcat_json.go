package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
)

type LogcatJSONParser struct {
	timestampFormat string
	eventRegex      *regexp.Regexp
	jsonExtraction  bool
}

// LogcatFile represents the structure of a .logcat file
type LogcatFile struct {
	Metadata struct {
		Device                map[string]interface{} `json:"device"`
		Filter                string                 `json:"filter"`
		ProjectApplicationIds []string               `json:"projectApplicationIds"`
	} `json:"metadata"`
	LogcatMessages []LogcatMessage `json:"logcatMessages"`
}

// LogcatMessage represents a single log entry in the .logcat file
type LogcatMessage struct {
	Header struct {
		LogLevel      string `json:"logLevel"`
		PID           int    `json:"pid"`
		TID           int    `json:"tid"`
		ApplicationId string `json:"applicationId"`
		ProcessName   string `json:"processName"`
		Tag           string `json:"tag"`
		Timestamp     struct {
			Seconds int64 `json:"seconds"`
			Nanos   int64 `json:"nanos"`
		} `json:"timestamp"`
	} `json:"header"`
	Message string `json:"message"`
}

func NewLogcatJSONParser() *LogcatJSONParser {
	return NewLogcatJSONParserWithConfig("", `.*Analytics: (.*)`, true)
}

func NewLogcatJSONParserWithConfig(timestampFormat, eventRegexPattern string, jsonExtraction bool) *LogcatJSONParser {
	logrus.WithFields(logrus.Fields{
		"timestamp_format":    timestampFormat,
		"event_regex_pattern": eventRegexPattern,
		"json_extraction":     jsonExtraction,
	}).Debug("Creating new LogcatJSON parser")

	// Default regex if empty
	if eventRegexPattern == "" {
		eventRegexPattern = `.*Analytics: (.*)`
		logrus.Debug("Using default event regex pattern")
	}

	// Compile event regex
	logrus.WithField("pattern", eventRegexPattern).Debug("Compiling event regex")
	eventRegex := regexp.MustCompile(eventRegexPattern)

	parser := &LogcatJSONParser{
		timestampFormat: timestampFormat, // Not used for .logcat files (uses epoch seconds + nanos)
		eventRegex:      eventRegex,
		jsonExtraction:  jsonExtraction,
	}

	logrus.Debug("LogcatJSON parser created successfully")
	return parser
}

func (p *LogcatJSONParser) Parse(logLine string) (*LogEntry, error) {
	return nil, fmt.Errorf("LogcatJSON parser does not support line-by-line parsing. Use ParseFile() instead")
}

func (p *LogcatJSONParser) ParseFile(filepath string) ([]*LogEntry, error) {
	logrus.WithField("filepath", filepath).Info("Starting to parse .logcat JSON file")

	// Read the entire file
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to read .logcat file")
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON structure
	var logcatFile LogcatFile
	if err := json.Unmarshal(fileData, &logcatFile); err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to parse .logcat JSON")
		return nil, fmt.Errorf("failed to parse .logcat JSON: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"total_messages": len(logcatFile.LogcatMessages),
		"device":         logcatFile.Metadata.Device,
		"filter":         logcatFile.Metadata.Filter,
	}).Info("Parsed .logcat file metadata")

	var entries []*LogEntry
	parsedCount := 0
	skippedCount := 0

	// Convert each LogcatMessage to LogEntry
	for i, logMsg := range logcatFile.LogcatMessages {
		entry, err := p.convertLogcatMessage(logMsg)
		if err != nil {
			skippedCount++
			logrus.WithError(err).WithField("message_index", i).Debug("Failed to convert logcat message, skipping")
			continue
		}

		entries = append(entries, entry)
		parsedCount++
	}

	logrus.WithFields(logrus.Fields{
		"filepath":         filepath,
		"total_messages":   len(logcatFile.LogcatMessages),
		"parsed_entries":   parsedCount,
		"skipped_messages": skippedCount,
	}).Info(".logcat JSON file parsing completed")

	return entries, nil
}

// convertLogcatMessage converts a LogcatMessage to a LogEntry
func (p *LogcatJSONParser) convertLogcatMessage(logMsg LogcatMessage) (*LogEntry, error) {
	// Convert timestamp from seconds + nanos to time.Time
	timestamp := time.Unix(logMsg.Header.Timestamp.Seconds, logMsg.Header.Timestamp.Nanos)

	entry := &LogEntry{
		Timestamp: timestamp,
		Level:     logMsg.Header.LogLevel,
		Tag:       logMsg.Header.Tag,
		PID:       logMsg.Header.PID,
		TID:       logMsg.Header.TID,
		Message:   logMsg.Message,
	}

	// Extract event data if enabled
	if p.jsonExtraction {
		logrus.Debug("Attempting to extract event data from .logcat message")
		p.extractEventData(entry)
	}

	return entry, nil
}

// extractEventData attempts to extract analytics event data from the log entry message
func (p *LogcatJSONParser) extractEventData(entry *LogEntry) {
	// First try the event regex pattern
	if p.eventRegex != nil {
		logrus.Debug("Trying to extract event data using regex pattern")
		matches := p.eventRegex.FindStringSubmatch(entry.Message)
		if len(matches) > 1 {
			jsonStr := matches[1]
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
func (p *LogcatJSONParser) tryParseJSON(entry *LogEntry, jsonStr string) bool {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &eventData); err == nil {
		entry.EventData = eventData
		logrus.WithField("event_keys", getMapKeysJSON(eventData)).Debug("JSON parsed successfully")
		return true
	}
	logrus.WithField("json_str", jsonStr).Debug("Failed to parse as JSON")
	return false
}

// Helper function to get map keys for logging
func getMapKeysJSON(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

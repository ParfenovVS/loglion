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
	// Default regex if empty
	if eventRegexPattern == "" {
		eventRegexPattern = `.*Analytics.*: (.*)`
	}

	// Default timestamp format if empty
	if timestampFormat == "" {
		timestampFormat = "01-02 15:04:05.000"
	}

	// Compile event regex
	eventRegex := regexp.MustCompile(eventRegexPattern)

	// Regex to parse the full logcat line format
	logLineRegex := regexp.MustCompile(`^(\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3})\s+(\d+)\s+(\d+)\s+([VDIWEFS])\s+([^:]+):\s*(.*)$`)

	return &AndroidParser{
		timestampFormat: timestampFormat,
		eventRegex:      eventRegex,
		jsonExtraction:  jsonExtraction,
		logLineRegex:    logLineRegex,
	}
}

func (p *AndroidParser) Parse(logLine string) (*LogEntry, error) {
	// Android logcat format: MM-dd HH:mm:ss.SSS  PID  TID LEVEL TAG: MESSAGE
	// Example: 01-15 10:30:15.123  1234  5678 I Analytics: {"event": "page_view"}

	// Use regex to parse the logcat line
	matches := p.logLineRegex.FindStringSubmatch(strings.TrimSpace(logLine))
	if len(matches) != 7 {
		return nil, fmt.Errorf("invalid log line format: %s", logLine)
	}

	// Extract components from regex groups
	timestampStr := matches[1]
	pidStr := matches[2]
	tidStr := matches[3]
	level := matches[4]
	tag := matches[5]
	message := matches[6]

	// Parse timestamp
	timestamp, err := time.Parse(p.timestampFormat, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp '%s': %w", timestampStr, err)
	}

	// Parse PID and TID
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PID '%s': %w", pidStr, err)
	}

	tid, err := strconv.Atoi(tidStr)
	if err != nil {
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
		p.extractEventData(entry, logLine)
	}

	return entry, nil
}

// extractEventData attempts to extract JSON event data from the log entry
func (p *AndroidParser) extractEventData(entry *LogEntry, logLine string) {
	// First try the event regex pattern
	if p.eventRegex != nil {
		matches := p.eventRegex.FindStringSubmatch(logLine)
		if len(matches) > 1 {
			jsonStr := strings.TrimSpace(matches[1])
			if p.tryParseJSON(entry, jsonStr) {
				return
			}
		}
	}

	// Fallback: try to parse the message directly as JSON
	p.tryParseJSON(entry, entry.Message)
}

// tryParseJSON attempts to parse a string as JSON and populate EventData
func (p *AndroidParser) tryParseJSON(entry *LogEntry, jsonStr string) bool {
	var eventData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &eventData); err == nil {
		entry.EventData = eventData
		return true
	}
	return false
}

func (p *AndroidParser) ParseFile(filepath string) ([]*LogEntry, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var entries []*LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		entry, err := p.Parse(line)
		if err != nil {
			// Log parsing error but continue with other lines
			continue
		}

		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return entries, nil
}

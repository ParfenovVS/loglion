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
}

func NewAndroidParser() *AndroidParser {
	// Default Android logcat format: MM-dd HH:mm:ss.SSS
	timestampFormat := "01-02 15:04:05.000"
	
	// Default regex to extract analytics events
	eventRegex := regexp.MustCompile(`.*Analytics.*: (.*)`)
	
	return &AndroidParser{
		timestampFormat: timestampFormat,
		eventRegex:      eventRegex,
		jsonExtraction:  true,
	}
}

func (p *AndroidParser) Parse(logLine string) (*LogEntry, error) {
	// Android logcat format: MM-dd HH:mm:ss.SSS  PID  TID LEVEL TAG: MESSAGE
	// Example: 01-15 10:30:15.123  1234  5678 I Analytics: {"event": "page_view"}
	
	parts := strings.Fields(logLine)
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid log line format")
	}
	
	// Parse timestamp (first two parts)
	timestampStr := parts[0] + " " + parts[1]
	timestamp, err := time.Parse(p.timestampFormat, timestampStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}
	
	// Parse PID and TID
	pid, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse PID: %w", err)
	}
	
	tid, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse TID: %w", err)
	}
	
	// Parse level and tag
	level := parts[4]
	tag := strings.TrimSuffix(parts[5], ":")
	
	// Extract message (everything after tag)
	messageStart := strings.Index(logLine, tag+":") + len(tag) + 1
	message := strings.TrimSpace(logLine[messageStart:])
	
	entry := &LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Tag:       tag,
		PID:       pid,
		TID:       tid,
		Message:   message,
	}
	
	// Try to extract JSON data if enabled
	if p.jsonExtraction && p.eventRegex != nil {
		matches := p.eventRegex.FindStringSubmatch(logLine)
		if len(matches) > 1 {
			var eventData map[string]interface{}
			if err := json.Unmarshal([]byte(matches[1]), &eventData); err == nil {
				entry.EventData = eventData
			}
		}
	}
	
	return entry, nil
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
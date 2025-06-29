package parser

import (
	"time"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Tag       string
	PID       int
	TID       int
	Message   string
	EventData map[string]interface{}
}

type Parser interface {
	Parse(logLine string) (*LogEntry, error)
	ParseFile(filepath string) ([]*LogEntry, error)
}

func NewParser() Parser {
	return NewPlainParser()
}

func NewParserWithConfig(timestampFormat, eventRegex string, jsonExtraction bool, logLineRegex string) Parser {
	return NewPlainParserWithConfig(timestampFormat, eventRegex, jsonExtraction, logLineRegex)
}

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

type LogFormat string

const (
	AndroidFormat LogFormat = "android"
)

func NewParser(format LogFormat) Parser {
	switch format {
	case AndroidFormat:
		return NewAndroidParser()
	default:
		return NewAndroidParser() // Default to Android
	}
}

func NewParserWithConfig(format LogFormat, timestampFormat, eventRegex string, jsonExtraction bool) Parser {
	switch format {
	case AndroidFormat:
		return NewAndroidParserWithConfig(timestampFormat, eventRegex, jsonExtraction)
	default:
		return NewAndroidParserWithConfig(timestampFormat, eventRegex, jsonExtraction) // Default to Android
	}
}

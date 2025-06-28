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
	PlainFormat      LogFormat = "plain"
	LogcatJSONFormat LogFormat = "logcat-json"
)

func NewParser(format LogFormat) Parser {
	switch format {
	case PlainFormat:
		return NewPlainParser()
	case LogcatJSONFormat:
		return NewLogcatJSONParser()
	default:
		return NewPlainParser() // Default to plain
	}
}

func NewParserWithConfig(format LogFormat, timestampFormat, eventRegex string, jsonExtraction bool, logLineRegex string) Parser {
	switch format {
	case PlainFormat:
		return NewPlainParserWithConfig(timestampFormat, eventRegex, jsonExtraction, logLineRegex)
	case LogcatJSONFormat:
		return NewLogcatJSONParserWithConfig(timestampFormat, eventRegex, jsonExtraction)
	default:
		return NewPlainParserWithConfig(timestampFormat, eventRegex, jsonExtraction, logLineRegex) // Default to plain
	}
}

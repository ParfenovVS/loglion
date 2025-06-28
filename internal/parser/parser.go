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
	LogcatPlainFormat LogFormat = "logcat-plain"
	LogcatJSONFormat  LogFormat = "logcat-json"
)

func NewParser(format LogFormat) Parser {
	switch format {
	case LogcatPlainFormat:
		return NewLogcatPlainParser()
	case LogcatJSONFormat:
		return NewLogcatJSONParser()
	default:
		return NewLogcatPlainParser() // Default to plain logcat
	}
}

func NewParserWithConfig(format LogFormat, timestampFormat, eventRegex string, jsonExtraction bool) Parser {
	switch format {
	case LogcatPlainFormat:
		return NewLogcatPlainParserWithConfig(timestampFormat, eventRegex, jsonExtraction)
	case LogcatJSONFormat:
		return NewLogcatJSONParserWithConfig(timestampFormat, eventRegex, jsonExtraction)
	default:
		return NewLogcatPlainParserWithConfig(timestampFormat, eventRegex, jsonExtraction) // Default to plain logcat
	}
}

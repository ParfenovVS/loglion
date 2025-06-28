package config

import (
	"fmt"
	"os"
	"regexp"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version   string    `yaml:"version"`
	Format    string    `yaml:"format"`
	Funnel    Funnel    `yaml:"funnel"`
	LogParser LogParser `yaml:"log_parser,omitempty"`
}

type Funnel struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name               string            `yaml:"name"`
	EventPattern       string            `yaml:"event_pattern"`
	RequiredProperties map[string]string `yaml:"required_properties,omitempty"`
}

type LogParser struct {
	TimestampFormat string `yaml:"timestamp_format"`
	EventRegex      string `yaml:"event_regex"`
	JSONExtraction  bool   `yaml:"json_extraction"`
	LogLineRegex    string `yaml:"log_line_regex"`
}

func LoadConfig(filepath string) (*Config, error) {
	logrus.WithField("filepath", filepath).Debug("Starting config load")

	if filepath == "" {
		logrus.Error("Config file path is empty")
		return nil, fmt.Errorf("config file path is required")
	}

	logrus.WithField("filepath", filepath).Debug("Reading config file")
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.WithField("filepath", filepath).Error("Config file not found")
			return nil, fmt.Errorf("config file not found: %s", filepath)
		}
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to read config file")
		return nil, fmt.Errorf("failed to read config file '%s': %w", filepath, err)
	}

	if len(data) == 0 {
		logrus.WithField("filepath", filepath).Error("Config file is empty")
		return nil, fmt.Errorf("config file is empty: %s", filepath)
	}

	logrus.WithFields(logrus.Fields{
		"filepath": filepath,
		"size":     len(data),
	}).Debug("Config file read successfully, parsing YAML")

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to parse YAML config")
		return nil, fmt.Errorf("failed to parse YAML config file '%s': %w", filepath, err)
	}

	logrus.WithFields(logrus.Fields{
		"version": config.Version,
		"format":  config.Format,
		"funnel":  config.Funnel.Name,
	}).Debug("Config parsed successfully, starting validation")

	if err := config.Validate(); err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Config validation failed")
		return nil, fmt.Errorf("config validation failed for '%s': %w", filepath, err)
	}

	logrus.WithField("filepath", filepath).Info("Config loaded and validated successfully")
	return &config, nil
}

func (c *Config) Validate() error {
	logrus.Debug("Starting config validation")

	if c.Version == "" {
		logrus.Error("Config validation failed: version is required")
		return fmt.Errorf("version is required")
	}
	logrus.WithField("version", c.Version).Debug("Version validation passed")

	if c.Format == "" {
		c.Format = "plain" // Default to plain format
		logrus.Debug("Format not specified, defaulting to 'plain'")
	}

	// Backward compatibility: map 'android' and 'logcat-plain' to 'plain'
	if c.Format == "android" || c.Format == "logcat-plain" {
		c.Format = "plain"
		logrus.Debug("Format mapped to 'plain' for backward compatibility")
	}

	if c.Format != "plain" && c.Format != "logcat-json" {
		logrus.WithField("format", c.Format).Error("Unsupported format")
		return fmt.Errorf("unsupported format: %s (supported formats: 'plain', 'logcat-json')", c.Format)
	}
	logrus.WithField("format", c.Format).Debug("Format validation passed")

	logrus.Debug("Validating funnel configuration")
	if err := c.validateFunnel(); err != nil {
		logrus.WithError(err).Error("Funnel validation failed")
		return fmt.Errorf("funnel validation failed: %w", err)
	}

	logrus.Debug("Validating log parser configuration")
	if err := c.validateLogParser(); err != nil {
		logrus.WithError(err).Error("Log parser validation failed")
		return fmt.Errorf("log_parser validation failed: %w", err)
	}

	logrus.Debug("Config validation completed successfully")
	return nil
}

func (c *Config) validateFunnel() error {
	if c.Funnel.Name == "" {
		logrus.Error("Funnel name is required")
		return fmt.Errorf("name is required")
	}
	logrus.WithField("funnel_name", c.Funnel.Name).Debug("Funnel name validation passed")

	if len(c.Funnel.Steps) == 0 {
		logrus.Error("Funnel must have at least one step")
		return fmt.Errorf("must have at least one step")
	}

	if len(c.Funnel.Steps) > 100 {
		logrus.WithField("step_count", len(c.Funnel.Steps)).Error("Too many funnel steps")
		return fmt.Errorf("too many steps (maximum 100)")
	}

	logrus.WithField("step_count", len(c.Funnel.Steps)).Debug("Funnel step count validation passed")

	stepNames := make(map[string]bool)
	for i, step := range c.Funnel.Steps {
		logrus.WithFields(logrus.Fields{
			"step_index": i + 1,
			"step_name":  step.Name,
		}).Debug("Validating funnel step")

		if err := c.validateStep(i, step, stepNames); err != nil {
			return err
		}
	}

	logrus.WithField("funnel_name", c.Funnel.Name).Debug("Funnel validation completed successfully")
	return nil
}

func (c *Config) validateStep(index int, step Step, stepNames map[string]bool) error {
	if step.Name == "" {
		return fmt.Errorf("step %d: name is required", index+1)
	}

	if stepNames[step.Name] {
		return fmt.Errorf("step %d: duplicate step name '%s'", index+1, step.Name)
	}
	stepNames[step.Name] = true

	if step.EventPattern == "" {
		return fmt.Errorf("step %d (%s): event_pattern is required", index+1, step.Name)
	}

	if _, err := regexp.Compile(step.EventPattern); err != nil {
		return fmt.Errorf("step %d (%s): invalid event_pattern regex: %w", index+1, step.Name, err)
	}

	for propName, propPattern := range step.RequiredProperties {
		if propName == "" {
			return fmt.Errorf("step %d (%s): property name cannot be empty", index+1, step.Name)
		}
		if propPattern == "" {
			return fmt.Errorf("step %d (%s): property pattern for '%s' cannot be empty", index+1, step.Name, propName)
		}
		if _, err := regexp.Compile(propPattern); err != nil {
			return fmt.Errorf("step %d (%s): invalid regex pattern for property '%s': %w", index+1, step.Name, propName, err)
		}
	}

	return nil
}

func (c *Config) validateLogParser() error {
	// Set defaults based on format
	if c.Format == "plain" {
		if c.LogParser.TimestampFormat == "" {
			c.LogParser.TimestampFormat = "" // No default timestamp for plain format
			logrus.Debug("Timestamp format not specified for plain format, leaving empty")
		}

		if c.LogParser.EventRegex == "" {
			c.LogParser.EventRegex = "^(.*)$" // Default: entire line as event
			logrus.Debug("Event regex not specified, using default for plain format")
		}

		if c.LogParser.LogLineRegex == "" {
			c.LogParser.LogLineRegex = "^(.*)$" // Default: entire line
			logrus.Debug("Log line regex not specified, using default for plain format")
		}
	} else {
		// For logcat-json, use legacy defaults
		if c.LogParser.TimestampFormat == "" {
			c.LogParser.TimestampFormat = "01-02 15:04:05.000"
			logrus.Debug("Timestamp format not specified, using default")
		}

		if c.LogParser.EventRegex == "" {
			c.LogParser.EventRegex = ".*Analytics: (.*)"
			logrus.Debug("Event regex not specified, using default")
		}
	}

	logrus.WithField("timestamp_format", c.LogParser.TimestampFormat).Debug("Timestamp format validation passed")

	logrus.WithField("event_regex", c.LogParser.EventRegex).Debug("Validating event regex pattern")
	if _, err := regexp.Compile(c.LogParser.EventRegex); err != nil {
		logrus.WithError(err).WithField("event_regex", c.LogParser.EventRegex).Error("Invalid event regex pattern")
		return fmt.Errorf("invalid event_regex: %w", err)
	}

	if c.LogParser.LogLineRegex != "" {
		logrus.WithField("log_line_regex", c.LogParser.LogLineRegex).Debug("Validating log line regex pattern")
		if _, err := regexp.Compile(c.LogParser.LogLineRegex); err != nil {
			logrus.WithError(err).WithField("log_line_regex", c.LogParser.LogLineRegex).Error("Invalid log line regex pattern")
			return fmt.Errorf("invalid log_line_regex: %w", err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"timestamp_format": c.LogParser.TimestampFormat,
		"event_regex":      c.LogParser.EventRegex,
		"log_line_regex":   c.LogParser.LogLineRegex,
		"json_extraction":  c.LogParser.JSONExtraction,
	}).Debug("Log parser validation completed successfully")

	return nil
}

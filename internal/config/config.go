package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

type ParserConfig struct {
	TimestampFormat string `yaml:"timestamp_format"`
	EventRegex      string `yaml:"event_regex"`
	JSONExtraction  bool   `yaml:"json_extraction"`
	LogLineRegex    string `yaml:"log_line_regex"`
}

type FunnelConfig struct {
	Name  string `yaml:"name"`
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name               string            `yaml:"name"`
	EventPattern       string            `yaml:"event_pattern"`
	RequiredProperties map[string]string `yaml:"required_properties,omitempty"`
}

func LoadParserConfig(filepath string) (*ParserConfig, error) {
	logrus.WithField("filepath", filepath).Debug("Starting parser config load")

	if filepath == "" {
		logrus.Error("Parser config file path is empty")
		return nil, fmt.Errorf("parser config file path is required")
	}

	logrus.WithField("filepath", filepath).Debug("Reading parser config file")
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.WithField("filepath", filepath).Error("Parser config file not found")
			return nil, fmt.Errorf("parser config file not found: %s", filepath)
		}
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to read parser config file")
		return nil, fmt.Errorf("failed to read parser config file '%s': %w", filepath, err)
	}

	if len(data) == 0 {
		logrus.WithField("filepath", filepath).Error("Parser config file is empty")
		return nil, fmt.Errorf("parser config file is empty: %s", filepath)
	}

	logrus.WithFields(logrus.Fields{
		"filepath": filepath,
		"size":     len(data),
	}).Debug("Parser config file read successfully, parsing YAML")

	var config ParserConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to parse YAML parser config")
		return nil, fmt.Errorf("failed to parse YAML parser config file '%s': %w", filepath, err)
	}

	logrus.Debug("Parser config parsed successfully, starting schema validation")

	if err := validateParserSchema(data); err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Parser schema validation failed")
		return nil, fmt.Errorf("parser schema validation failed for '%s': %w", filepath, err)
	}

	logrus.Debug("Parser schema validation passed, starting struct validation")
	if err := config.Validate(); err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Parser config validation failed")
		return nil, fmt.Errorf("parser config validation failed for '%s': %w", filepath, err)
	}

	logrus.WithField("filepath", filepath).Info("Parser config loaded and validated successfully")
	return &config, nil
}

func LoadFunnelConfig(filepath string) (*FunnelConfig, error) {
	logrus.WithField("filepath", filepath).Debug("Starting funnel config load")

	if filepath == "" {
		logrus.Error("Funnel config file path is empty")
		return nil, fmt.Errorf("funnel config file path is required")
	}

	logrus.WithField("filepath", filepath).Debug("Reading funnel config file")
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.WithField("filepath", filepath).Error("Funnel config file not found")
			return nil, fmt.Errorf("funnel config file not found: %s", filepath)
		}
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to read funnel config file")
		return nil, fmt.Errorf("failed to read funnel config file '%s': %w", filepath, err)
	}

	if len(data) == 0 {
		logrus.WithField("filepath", filepath).Error("Funnel config file is empty")
		return nil, fmt.Errorf("funnel config file is empty: %s", filepath)
	}

	logrus.WithFields(logrus.Fields{
		"filepath": filepath,
		"size":     len(data),
	}).Debug("Funnel config file read successfully, parsing YAML")

	var config FunnelConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Failed to parse YAML funnel config")
		return nil, fmt.Errorf("failed to parse YAML funnel config file '%s': %w", filepath, err)
	}

	logrus.WithField("funnel", config.Name).Debug("Funnel config parsed successfully, starting schema validation")

	if err := validateFunnelSchema(data); err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Funnel schema validation failed")
		return nil, fmt.Errorf("funnel schema validation failed for '%s': %w", filepath, err)
	}

	logrus.Debug("Funnel schema validation passed, starting struct validation")
	if err := config.Validate(); err != nil {
		logrus.WithError(err).WithField("filepath", filepath).Error("Funnel config validation failed")
		return nil, fmt.Errorf("funnel config validation failed for '%s': %w", filepath, err)
	}

	logrus.WithField("filepath", filepath).Info("Funnel config loaded and validated successfully")
	return &config, nil
}

func validateParserSchema(yamlData []byte) error {
	// Get the schema file path relative to the project root
	schemaPath := "schema/parser-config.schema.json"

	// Try to find the schema file
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// If not found in current directory, try to find it relative to the config package
		wd, _ := os.Getwd()
		projectRoot := filepath.Dir(filepath.Dir(wd)) // Go up from internal/config to project root
		schemaPath = filepath.Join(projectRoot, "schema", "parser-config.schema.json")

		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			logrus.Warn("Parser schema file not found, skipping schema validation")
			return nil
		}
	}

	logrus.WithField("schema_path", schemaPath).Debug("Loading parser JSON schema")
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)

	// Convert YAML to JSON for validation
	var yamlObj interface{}
	if err := yaml.Unmarshal(yamlData, &yamlObj); err != nil {
		return fmt.Errorf("failed to parse YAML for parser schema validation: %w", err)
	}

	jsonData, err := json.Marshal(yamlObj)
	if err != nil {
		return fmt.Errorf("failed to convert YAML to JSON for parser schema validation: %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonData)

	logrus.Debug("Performing parser JSON schema validation")
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("parser schema validation error: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, fmt.Sprintf("- %s", desc))
		}
		return fmt.Errorf("parser schema validation failed:\n%s", fmt.Sprintf("%v", errors))
	}

	logrus.Debug("Parser schema validation completed successfully")
	return nil
}

func validateFunnelSchema(yamlData []byte) error {
	// Get the schema file path relative to the project root
	schemaPath := "schema/funnel-config.schema.json"

	// Try to find the schema file
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		// If not found in current directory, try to find it relative to the config package
		wd, _ := os.Getwd()
		projectRoot := filepath.Dir(filepath.Dir(wd)) // Go up from internal/config to project root
		schemaPath = filepath.Join(projectRoot, "schema", "funnel-config.schema.json")

		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			logrus.Warn("Funnel schema file not found, skipping schema validation")
			return nil
		}
	}

	logrus.WithField("schema_path", schemaPath).Debug("Loading funnel JSON schema")
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)

	// Convert YAML to JSON for validation
	var yamlObj interface{}
	if err := yaml.Unmarshal(yamlData, &yamlObj); err != nil {
		return fmt.Errorf("failed to parse YAML for funnel schema validation: %w", err)
	}

	jsonData, err := json.Marshal(yamlObj)
	if err != nil {
		return fmt.Errorf("failed to convert YAML to JSON for funnel schema validation: %w", err)
	}

	documentLoader := gojsonschema.NewBytesLoader(jsonData)

	logrus.Debug("Performing funnel JSON schema validation")
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("funnel schema validation error: %w", err)
	}

	if !result.Valid() {
		var errors []string
		for _, desc := range result.Errors() {
			errors = append(errors, fmt.Sprintf("- %s", desc))
		}
		return fmt.Errorf("funnel schema validation failed:\n%s", fmt.Sprintf("%v", errors))
	}

	logrus.Debug("Funnel schema validation completed successfully")
	return nil
}

func (c *ParserConfig) Validate() error {
	logrus.Debug("Starting parser config validation")

	// Set defaults for plain format (the only supported format)
	if c.TimestampFormat == "" {
		c.TimestampFormat = "" // No default timestamp for plain format
		logrus.Debug("Timestamp format not specified for plain format, leaving empty")
	}

	if c.EventRegex == "" {
		c.EventRegex = "^(.*)$" // Default: entire line as event
		logrus.Debug("Event regex not specified, using default for plain format")
	}

	if c.LogLineRegex == "" {
		c.LogLineRegex = "^(.*)$" // Default: entire line
		logrus.Debug("Log line regex not specified, using default for plain format")
	}

	logrus.WithField("timestamp_format", c.TimestampFormat).Debug("Timestamp format validation passed")

	logrus.WithField("event_regex", c.EventRegex).Debug("Validating event regex pattern")
	if _, err := regexp.Compile(c.EventRegex); err != nil {
		logrus.WithError(err).WithField("event_regex", c.EventRegex).Error("Invalid event regex pattern")
		return fmt.Errorf("invalid event_regex: %w", err)
	}

	if c.LogLineRegex != "" {
		logrus.WithField("log_line_regex", c.LogLineRegex).Debug("Validating log line regex pattern")
		if _, err := regexp.Compile(c.LogLineRegex); err != nil {
			logrus.WithError(err).WithField("log_line_regex", c.LogLineRegex).Error("Invalid log line regex pattern")
			return fmt.Errorf("invalid log_line_regex: %w", err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"timestamp_format": c.TimestampFormat,
		"event_regex":      c.EventRegex,
		"log_line_regex":   c.LogLineRegex,
		"json_extraction":  c.JSONExtraction,
	}).Debug("Parser config validation completed successfully")

	return nil
}

func (c *FunnelConfig) Validate() error {
	logrus.Debug("Starting funnel config validation")

	if c.Name == "" {
		logrus.Error("Funnel name is required")
		return fmt.Errorf("name is required")
	}
	logrus.WithField("funnel_name", c.Name).Debug("Funnel name validation passed")

	if len(c.Steps) == 0 {
		logrus.Error("Funnel must have at least one step")
		return fmt.Errorf("must have at least one step")
	}

	if len(c.Steps) > 100 {
		logrus.WithField("step_count", len(c.Steps)).Error("Too many funnel steps")
		return fmt.Errorf("too many steps (maximum 100)")
	}

	logrus.WithField("step_count", len(c.Steps)).Debug("Funnel step count validation passed")

	stepNames := make(map[string]bool)
	for i, step := range c.Steps {
		logrus.WithFields(logrus.Fields{
			"step_index": i + 1,
			"step_name":  step.Name,
		}).Debug("Validating funnel step")

		if err := c.validateStep(i, step, stepNames); err != nil {
			return err
		}
	}

	logrus.WithField("funnel_name", c.Name).Debug("Funnel config validation completed successfully")
	return nil
}

func (c *FunnelConfig) validateStep(index int, step Step, stepNames map[string]bool) error {
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

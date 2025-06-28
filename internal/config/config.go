package config

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string `yaml:"version"`
	Format  string `yaml:"format"`
	Funnel  Funnel `yaml:"funnel"`
	AndroidParser AndroidParser `yaml:"android_parser,omitempty"`
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

type AndroidParser struct {
	TimestampFormat string `yaml:"timestamp_format"`
	EventRegex      string `yaml:"event_regex"`
	JSONExtraction  bool   `yaml:"json_extraction"`
}

func LoadConfig(filepath string) (*Config, error) {
	if filepath == "" {
		return nil, fmt.Errorf("config file path is required")
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", filepath)
		}
		return nil, fmt.Errorf("failed to read config file '%s': %w", filepath, err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("config file is empty: %s", filepath)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config file '%s': %w", filepath, err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed for '%s': %w", filepath, err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	
	if c.Format == "" {
		c.Format = "android" // Default to android format
	}
	
	if c.Format != "android" {
		return fmt.Errorf("unsupported format: %s (only 'android' is supported)", c.Format)
	}
	
	if err := c.validateFunnel(); err != nil {
		return fmt.Errorf("funnel validation failed: %w", err)
	}
	
	if err := c.validateAndroidParser(); err != nil {
		return fmt.Errorf("android_parser validation failed: %w", err)
	}
	
	return nil
}

func (c *Config) validateFunnel() error {
	if c.Funnel.Name == "" {
		return fmt.Errorf("name is required")
	}
	
	if len(c.Funnel.Steps) == 0 {
		return fmt.Errorf("must have at least one step")
	}
	
	if len(c.Funnel.Steps) > 100 {
		return fmt.Errorf("too many steps (maximum 100)")
	}
	
	stepNames := make(map[string]bool)
	for i, step := range c.Funnel.Steps {
		if err := c.validateStep(i, step, stepNames); err != nil {
			return err
		}
	}
	
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

func (c *Config) validateAndroidParser() error {
	if c.AndroidParser.TimestampFormat == "" {
		c.AndroidParser.TimestampFormat = "01-02 15:04:05.000" // Default format
	}
	
	if c.AndroidParser.EventRegex == "" {
		c.AndroidParser.EventRegex = ".*Analytics.*: (.*)" // Default regex
	}
	
	if _, err := regexp.Compile(c.AndroidParser.EventRegex); err != nil {
		return fmt.Errorf("invalid event_regex: %w", err)
	}
	
	return nil
}
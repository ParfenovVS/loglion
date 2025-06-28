package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string `yaml:"version"`
	Format  string `yaml:"format"`
	Funnel  Funnel `yaml:"funnel"`
	AndroidParser AndroidParser `yaml:"android_parser,omitempty"`
}

type Funnel struct {
	Name           string `yaml:"name"`
	SessionKey     string `yaml:"session_key"`
	TimeoutMinutes int    `yaml:"timeout_minutes"`
	Steps          []Step `yaml:"steps"`
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
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	
	if c.Funnel.Name == "" {
		return fmt.Errorf("funnel name is required")
	}
	
	if c.Funnel.SessionKey == "" {
		return fmt.Errorf("funnel session_key is required")
	}
	
	if len(c.Funnel.Steps) == 0 {
		return fmt.Errorf("funnel must have at least one step")
	}
	
	for i, step := range c.Funnel.Steps {
		if step.Name == "" {
			return fmt.Errorf("step %d: name is required", i+1)
		}
		if step.EventPattern == "" {
			return fmt.Errorf("step %d: event_pattern is required", i+1)
		}
	}
	
	return nil
}
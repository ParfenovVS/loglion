package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadParserConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_parser_config",
			content: `timestamp_format: "01-02 15:04:05.000"
event_regex: "^(.*)$"
json_extraction: false
log_line_regex: "^(.*)$"`,
			expectError: false,
		},
		{
			name: "minimal_parser_config",
			content: `event_regex: "test.*"
json_extraction: true`,
			expectError: false,
		},
		{
			name: "invalid_event_regex",
			content: `event_regex: "[invalid"
json_extraction: false`,
			expectError: true,
			errorMsg:    "invalid event_regex",
		},
		{
			name: "invalid_log_line_regex",
			content: `event_regex: "valid"
log_line_regex: "[invalid"`,
			expectError: true,
			errorMsg:    "invalid log_line_regex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "parser.yaml")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test LoadParserConfig
			config, err := LoadParserConfig(tmpFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got none", tt.errorMsg)
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if config == nil {
					t.Error("Expected config to be non-nil")
				}
			}
		})
	}
}

func TestLoadFunnelConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_funnel_config",
			content: `name: "Test Funnel"
steps:
  - name: "Step1"
    event_pattern: "analytics.*test"
    required_properties:
      page: "/test"`,
			expectError: false,
		},
		{
			name: "minimal_funnel_config",
			content: `name: "Simple Test"
steps:
  - name: "Step1"
    event_pattern: "test"`,
			expectError: false,
		},
		{
			name: "invalid_regex",
			content: `name: "Test"
steps:
  - name: "Step1"
    event_pattern: "[invalid"`,
			expectError: true,
			errorMsg:    "invalid event_pattern regex",
		},
		{
			name: "empty_funnel_name",
			content: `steps:
  - name: "Step1"
    event_pattern: "test"`,
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "no_steps",
			content: `name: "Test"
steps: []`,
			expectError: true,
			errorMsg:    "Array must have at least 1 items",
		},
		{
			name: "duplicate_step_names",
			content: `name: "Test"
steps:
  - name: "Step1"
    event_pattern: "test1"
  - name: "Step1"
    event_pattern: "test2"`,
			expectError: true,
			errorMsg:    "duplicate step name",
		},
		{
			name: "invalid_property_regex",
			content: `name: "Test"
steps:
  - name: "Step1"
    event_pattern: "test"
    required_properties:
      prop: "[invalid"`,
			expectError: true,
			errorMsg:    "invalid regex pattern for property",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "funnel.yaml")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test LoadFunnelConfig
			config, err := LoadFunnelConfig(tmpFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got none", tt.errorMsg)
				} else if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if config == nil {
					t.Error("Expected config to be non-nil")
				}
			}
		})
	}
}

func TestParserConfigFileErrors(t *testing.T) {
	t.Run("empty_filepath", func(t *testing.T) {
		_, err := LoadParserConfig("")
		if err == nil {
			t.Error("Expected error for empty filepath")
		}
		if !containsString(err.Error(), "parser config file path is required") {
			t.Errorf("Expected error about required path, got: %v", err)
		}
	})

	t.Run("nonexistent_file", func(t *testing.T) {
		_, err := LoadParserConfig("/nonexistent/file.yaml")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
		if !containsString(err.Error(), "parser config file not found") {
			t.Errorf("Expected error about file not found, got: %v", err)
		}
	})

	t.Run("empty_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "empty.yaml")

		err := os.WriteFile(tmpFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		_, err = LoadParserConfig(tmpFile)
		if err == nil {
			t.Error("Expected error for empty file")
		}
		if !containsString(err.Error(), "parser config file is empty") {
			t.Errorf("Expected error about empty file, got: %v", err)
		}
	})

	t.Run("invalid_yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "invalid.yaml")

		err := os.WriteFile(tmpFile, []byte("invalid: yaml: content: ["), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid YAML file: %v", err)
		}

		_, err = LoadParserConfig(tmpFile)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
		if !containsString(err.Error(), "failed to parse YAML") {
			t.Errorf("Expected YAML parse error, got: %v", err)
		}
	})
}

func TestFunnelConfigFileErrors(t *testing.T) {
	t.Run("empty_filepath", func(t *testing.T) {
		_, err := LoadFunnelConfig("")
		if err == nil {
			t.Error("Expected error for empty filepath")
		}
		if !containsString(err.Error(), "funnel config file path is required") {
			t.Errorf("Expected error about required path, got: %v", err)
		}
	})

	t.Run("nonexistent_file", func(t *testing.T) {
		_, err := LoadFunnelConfig("/nonexistent/file.yaml")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
		if !containsString(err.Error(), "funnel config file not found") {
			t.Errorf("Expected error about file not found, got: %v", err)
		}
	})

	t.Run("empty_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "empty.yaml")

		err := os.WriteFile(tmpFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		_, err = LoadFunnelConfig(tmpFile)
		if err == nil {
			t.Error("Expected error for empty file")
		}
		if !containsString(err.Error(), "funnel config file is empty") {
			t.Errorf("Expected error about empty file, got: %v", err)
		}
	})

	t.Run("invalid_yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "invalid.yaml")

		err := os.WriteFile(tmpFile, []byte("invalid: yaml: content: ["), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid YAML file: %v", err)
		}

		_, err = LoadFunnelConfig(tmpFile)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
		if !containsString(err.Error(), "failed to parse YAML") {
			t.Errorf("Expected YAML parse error, got: %v", err)
		}
	})
}

func TestParserConfigValidateDefaults(t *testing.T) {
	config := &ParserConfig{}

	err := config.Validate()
	if err != nil {
		t.Errorf("Expected no error with valid config, got: %v", err)
	}

	// Check defaults were applied
	if config.TimestampFormat != "" {
		t.Errorf("Expected default timestamp format to be empty, got: %s", config.TimestampFormat)
	}
	if config.EventRegex != "^(.*)$" {
		t.Errorf("Expected default event regex, got: %s", config.EventRegex)
	}
	if config.LogLineRegex != "^(.*)$" {
		t.Errorf("Expected default log line regex, got: %s", config.LogLineRegex)
	}
}

func TestFunnelConfigValidateStepLimits(t *testing.T) {
	config := &FunnelConfig{
		Name:  "Test",
		Steps: make([]Step, 101), // Too many steps
	}

	// Fill with valid steps
	for i := 0; i < 101; i++ {
		config.Steps[i] = Step{
			Name:         "Step" + string(rune(i+65)), // Use letters A, B, C, etc.
			EventPattern: "test",
		}
	}

	err := config.Validate()
	if err == nil {
		t.Error("Expected error for too many steps")
	}
	if !containsString(err.Error(), "too many steps") {
		t.Errorf("Expected error about too many steps, got: %v", err)
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(substr) > 0 && len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
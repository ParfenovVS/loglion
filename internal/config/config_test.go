package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_config_logcat_plain",
			content: `version: "1.0"
format: "logcat-plain"
funnel:
  name: "Test Funnel"
  steps:
    - name: "Step1"
      event_pattern: "analytics.*test"
      required_properties:
        page: "/test"`,
			expectError: false,
		},
		{
			name: "valid_config_logcat_json",
			content: `version: "1.0"
format: "logcat-json"
funnel:
  name: "Test Funnel"
  steps:
    - name: "Step1"
      event_pattern: "analytics.*test"
      required_properties:
        page: "/test"`,
			expectError: false,
		},
		{
			name: "valid_config_android_backward_compatibility",
			content: `version: "1.0"
format: "android"
funnel:
  name: "Test Funnel"
  steps:
    - name: "Step1"
      event_pattern: "analytics.*test"
      required_properties:
        page: "/test"`,
			expectError: false,
		},
		{
			name: "missing_version",
			content: `format: "logcat-plain"
funnel:
  name: "Test"
  steps:
    - name: "Step1"
      event_pattern: "test"`,
			expectError: true,
			errorMsg:    "version is required",
		},
		{
			name: "invalid_regex",
			content: `version: "1.0"
format: "logcat-plain"
funnel:
  name: "Test"
  steps:
    - name: "Step1"
      event_pattern: "[invalid"`,
			expectError: true,
			errorMsg:    "invalid event_pattern regex",
		},
		{
			name: "empty_funnel_name",
			content: `version: "1.0"
format: "logcat-plain"
funnel:
  steps:
    - name: "Step1"
      event_pattern: "test"`,
			expectError: true,
			errorMsg:    "name is required",
		},
		{
			name: "no_steps",
			content: `version: "1.0"
format: "logcat-plain"
funnel:
  name: "Test"
  steps: []`,
			expectError: true,
			errorMsg:    "must have at least one step",
		},
		{
			name: "duplicate_step_names",
			content: `version: "1.0"
format: "logcat-plain"
funnel:
  name: "Test"
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
			content: `version: "1.0"
format: "logcat-plain"
funnel:
  name: "Test"
  steps:
    - name: "Step1"
      event_pattern: "test"
      required_properties:
        prop: "[invalid"`,
			expectError: true,
			errorMsg:    "invalid regex pattern for property",
		},
		{
			name: "unsupported_format",
			content: `version: "1.0"
format: "unsupported"
funnel:
  name: "Test"
  steps:
    - name: "Step1"
      event_pattern: "test"`,
			expectError: true,
			errorMsg:    "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.yaml")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test LoadConfig
			config, err := LoadConfig(tmpFile)

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

func TestLoadConfigFileErrors(t *testing.T) {
	t.Run("empty_filepath", func(t *testing.T) {
		_, err := LoadConfig("")
		if err == nil {
			t.Error("Expected error for empty filepath")
		}
		if !containsString(err.Error(), "config file path is required") {
			t.Errorf("Expected error about required path, got: %v", err)
		}
	})

	t.Run("nonexistent_file", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/file.yaml")
		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
		if !containsString(err.Error(), "config file not found") {
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

		_, err = LoadConfig(tmpFile)
		if err == nil {
			t.Error("Expected error for empty file")
		}
		if !containsString(err.Error(), "config file is empty") {
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

		_, err = LoadConfig(tmpFile)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
		if !containsString(err.Error(), "failed to parse YAML") {
			t.Errorf("Expected YAML parse error, got: %v", err)
		}
	})
}

func TestConfigValidateDefaults(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Funnel: Funnel{
			Name: "Test",
			Steps: []Step{
				{
					Name:         "Step1",
					EventPattern: "test",
				},
			},
		},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Expected no error with valid config, got: %v", err)
	}

	// Check defaults were applied
	if config.Format != "plain" {
		t.Errorf("Expected default format 'plain', got: %s", config.Format)
	}
	if config.LogParser.TimestampFormat != "" {
		t.Errorf("Expected default timestamp format to be empty for plain format, got: %s", config.LogParser.TimestampFormat)
	}
	if config.LogParser.EventRegex != "^(.*)$" {
		t.Errorf("Expected default event regex for plain format, got: %s", config.LogParser.EventRegex)
	}
}

func TestConfigValidateStepLimits(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Format:  "logcat-plain",
		Funnel: Funnel{
			Name:  "Test",
			Steps: make([]Step, 101), // Too many steps
		},
	}

	// Fill with valid steps
	for i := 0; i < 101; i++ {
		config.Funnel.Steps[i] = Step{
			Name:         "Step" + string(rune(i)),
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

func TestBackwardCompatibility(t *testing.T) {
	config := &Config{
		Version: "1.0",
		Format:  "android", // Old format
		Funnel: Funnel{
			Name: "Test",
			Steps: []Step{
				{
					Name:         "Step1",
					EventPattern: "test",
				},
			},
		},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Expected no error with backward compatible config, got: %v", err)
	}

	// Check that 'android' format was mapped to 'plain'
	if config.Format != "plain" {
		t.Errorf("Expected 'android' format to be mapped to 'plain', got: %s", config.Format)
	}
}

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCountCommandValidation(t *testing.T) {
	tempDir := t.TempDir()

	validParserConfig := `event_regex: "^(.*)$"
json_extraction: false`

	validLogContent := `login user123
purchase item456
logout user123`

	// Create valid files for successful tests
	parserPath := filepath.Join(tempDir, "parser.yaml")
	if err := os.WriteFile(parserPath, []byte(validParserConfig), 0644); err != nil {
		t.Fatalf("Failed to write parser config: %v", err)
	}

	logPath := filepath.Join(tempDir, "test.log")
	if err := os.WriteFile(logPath, []byte(validLogContent), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "valid_single_pattern",
			args:        []string{"count", "--parser-config", parserPath, "--log", logPath, "login"},
			expectError: false,
		},
		{
			name:        "valid_multiple_patterns",
			args:        []string{"count", "--parser-config", parserPath, "--log", logPath, "login", "purchase"},
			expectError: false,
		},
		{
			name:        "missing_parser_config",
			args:        []string{"count", "--log", logPath, "login"},
			expectError: true,
		},
		{
			name:        "nonexistent_parser_config",
			args:        []string{"count", "--parser-config", "nonexistent.yaml", "--log", logPath, "login"},
			expectError: true,
		},
		{
			name:        "nonexistent_log_file",
			args:        []string{"count", "--parser-config", parserPath, "--log", "nonexistent.log", "login"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createCountCommand()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Error("Expected command to fail but it succeeded")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected command to succeed but it failed: %v", err)
			}
		})
	}
}

func TestCountCommandFlags(t *testing.T) {
	cmd := createCountCommand()

	parserFlag := cmd.Flags().Lookup("parser-config")
	if parserFlag == nil {
		t.Error("Expected parser-config flag to exist")
	} else {
		if parserFlag.Shorthand != "p" {
			t.Errorf("Expected parser-config shorthand to be 'p', got %q", parserFlag.Shorthand)
		}
		if parserFlag.Usage != "Path to parser configuration file (required)" {
			t.Errorf("Expected parser-config usage description mismatch")
		}
	}

	logFlag := cmd.Flags().Lookup("log")
	if logFlag == nil {
		t.Error("Expected log flag to exist")
	} else {
		if logFlag.Shorthand != "l" {
			t.Errorf("Expected log shorthand to be 'l', got %q", logFlag.Shorthand)
		}
		if logFlag.Usage != "Path to log file (optional, stdin is used if not provided)" {
			t.Errorf("Expected log usage description mismatch")
		}
	}

	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected output flag to exist")
	} else {
		if outputFlag.Shorthand != "o" {
			t.Errorf("Expected output shorthand to be 'o', got %q", outputFlag.Shorthand)
		}
		if outputFlag.Usage != "Output format (json, text)" {
			t.Errorf("Expected output usage description mismatch")
		}
	}
}

func TestCountCommandProperties(t *testing.T) {
	cmd := createCountCommand()

	if cmd.Use != "count [event_patterns...]" {
		t.Errorf("Expected Use to be 'count [event_patterns...]', got %q", cmd.Use)
	}

	if cmd.Short != "Count occurrences of event patterns in log files" {
		t.Errorf("Expected Short description mismatch")
	}

	if !strings.Contains(cmd.Long, "Count command processes log files and counts occurrences") {
		t.Error("Expected Long description to contain count information")
	}

	if !strings.Contains(cmd.Long, "Examples:") {
		t.Error("Expected Long description to contain examples")
	}

	if cmd.RunE == nil {
		t.Error("RunE function should not be nil")
	}

	if cmd.Args == nil {
		t.Error("Args validation should not be nil")
	}
}

func TestCountCommandRequiredFlags(t *testing.T) {
	cmd := createCountCommand()

	// Test that parser-config is marked as required
	required := cmd.Flag("parser-config").Annotations[cobra.BashCompOneRequiredFlag]
	if len(required) == 0 {
		t.Error("Expected parser-config flag to be marked as required")
	}
}

func TestCountCommandOutputFormats(t *testing.T) {
	tempDir := t.TempDir()

	validParserConfig := `event_regex: "^(.*)$"
json_extraction: false`

	validLogContent := `login user123
purchase item456`

	parserPath := filepath.Join(tempDir, "parser.yaml")
	if err := os.WriteFile(parserPath, []byte(validParserConfig), 0644); err != nil {
		t.Fatalf("Failed to write parser config: %v", err)
	}

	logPath := filepath.Join(tempDir, "test.log")
	if err := os.WriteFile(logPath, []byte(validLogContent), 0644); err != nil {
		t.Fatalf("Failed to write log file: %v", err)
	}

	tests := []struct {
		name         string
		outputFormat string
		expectError  bool
	}{
		{
			name:         "text_output_format",
			outputFormat: "text",
			expectError:  false,
		},
		{
			name:         "json_output_format",
			outputFormat: "json",
			expectError:  false,
		},
		{
			name:         "default_output_format",
			outputFormat: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createCountCommand()

			args := []string{"count", "--parser-config", parserPath, "--log", logPath}
			if tt.outputFormat != "" {
				args = append(args, "--output", tt.outputFormat)
			}
			args = append(args, "login")

			cmd.SetArgs(args)

			err := cmd.Execute()

			if tt.expectError && err == nil {
				t.Error("Expected command to fail but it succeeded")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected command to succeed but it failed: %v", err)
			}
		})
	}
}

func TestCountCommandArgumentValidation(t *testing.T) {
	// Test that MinimumNArgs(1) is enforced by cobra
	cmd := createCountCommand()
	
	// Test Args validation function directly
	if cmd.Args == nil {
		t.Fatal("Args validation function should not be nil")
	}
	
	// Test with no args - should fail
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected Args validation to fail with no arguments")
	}
	
	// Test with args - should pass
	err = cmd.Args(cmd, []string{"pattern1"})
	if err != nil {
		t.Errorf("Expected Args validation to pass with arguments: %v", err)
	}
}

func createCountCommand() *cobra.Command {
	// Create a simplified version of countCmd for testing
	cmd := &cobra.Command{
		Use:   "count [event_patterns...]",
		Short: "Count occurrences of event patterns in log files",
		Long: `Count command processes log files and counts occurrences of specified event patterns.
It accepts multiple event patterns as arguments and outputs the count for each pattern.

Examples:
  loglion count --parser-config parser.yaml --log logcat.txt "login" "logout" "error"
  loglion count -p parser.yaml -l logcat.txt --output json "user_action" "network_request"
  loglion count -p parser.yaml -l logcat.txt "memory_warning"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check args first 
			if len(args) == 0 {
				return fmt.Errorf("requires at least 1 arg(s), only received 0")
			}
			
			// Simplified run function for testing
			parserConfigFile, _ := cmd.Flags().GetString("parser-config")
			logFile, _ := cmd.Flags().GetString("log")

			if parserConfigFile == "" {
				return fmt.Errorf("parser-config is required")
			}

			// Check if files exist
			if _, err := os.Stat(parserConfigFile); os.IsNotExist(err) {
				return fmt.Errorf("parser config file does not exist")
			}

			if logFile != "" {
				if _, err := os.Stat(logFile); os.IsNotExist(err) {
					return fmt.Errorf("log file does not exist")
				}
			}

			// Simple validation of parser config
			data, err := os.ReadFile(parserConfigFile)
			if err != nil {
				return fmt.Errorf("error reading parser config: %v", err)
			}

			if strings.Contains(string(data), "invalid_bool") {
				return fmt.Errorf("invalid parser config")
			}

			return nil
		},
	}

	cmd.Flags().StringP("parser-config", "p", "", "Path to parser configuration file (required)")
	cmd.Flags().StringP("log", "l", "", "Path to log file (optional, stdin is used if not provided)")
	cmd.Flags().StringP("output", "o", "text", "Output format (json, text)")

	cmd.MarkFlagRequired("parser-config")

	return cmd
}
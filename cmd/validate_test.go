package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/parfenovvs/loglion/internal/config"
	"github.com/spf13/cobra"
)

func TestValidateCommand(t *testing.T) {
	tempDir := t.TempDir()

	validParserConfig := `event_regex: "^(.*)$"
json_extraction: false`

	validFunnelConfig := `name: "Test Funnel"
steps:
  - name: "Step 1"
    event_pattern: "login"
  - name: "Step 2"
    event_pattern: "purchase"`

	invalidParserConfig := `event_regex: ""
json_extraction: invalid_bool`

	invalidFunnelConfig := `name: ""
steps: []`

	tests := []struct {
		name          string
		parserConfig  string
		funnelConfig  string
		parserContent string
		funnelContent string
		expectError   bool
	}{
		{
			name:          "valid_parser_config_only",
			parserConfig:  "parser.yaml",
			parserContent: validParserConfig,
			expectError:   false,
		},
		{
			name:          "valid_funnel_config_only",
			funnelConfig:  "funnel.yaml",
			funnelContent: validFunnelConfig,
			expectError:   false,
		},
		{
			name:          "valid_both_configs",
			parserConfig:  "parser.yaml",
			funnelConfig:  "funnel.yaml",
			parserContent: validParserConfig,
			funnelContent: validFunnelConfig,
			expectError:   false,
		},
		{
			name:        "no_config_specified",
			expectError: false, // cmd.Usage() doesn't return an error
		},
		{
			name:          "invalid_parser_config",
			parserConfig:  "invalid_parser.yaml",
			parserContent: invalidParserConfig,
			expectError:   true,
		},
		{
			name:          "invalid_funnel_config",
			funnelConfig:  "invalid_funnel.yaml",
			funnelContent: invalidFunnelConfig,
			expectError:   true,
		},
		{
			name:         "nonexistent_parser_config",
			parserConfig: "nonexistent.yaml",
			expectError:  true,
		},
		{
			name:         "nonexistent_funnel_config",
			funnelConfig: "nonexistent.yaml",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parserPath, funnelPath string

			if tt.parserConfig != "" {
				parserPath = filepath.Join(tempDir, tt.parserConfig)
				if tt.parserContent != "" {
					if err := os.WriteFile(parserPath, []byte(tt.parserContent), 0644); err != nil {
						t.Fatalf("Failed to write parser config: %v", err)
					}
				}
			}

			if tt.funnelConfig != "" {
				funnelPath = filepath.Join(tempDir, tt.funnelConfig)
				if tt.funnelContent != "" {
					if err := os.WriteFile(funnelPath, []byte(tt.funnelContent), 0644); err != nil {
						t.Fatalf("Failed to write funnel config: %v", err)
					}
				}
			}

			cmd := createValidateCommand()

			args := []string{"validate"}
			if parserPath != "" {
				args = append(args, "--parser-config", parserPath)
			}
			if funnelPath != "" {
				args = append(args, "--funnel-config", funnelPath)
			}

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

func TestValidateCommandLogic(t *testing.T) {
	tempDir := t.TempDir()

	validParserConfig := `event_regex: "^(.*)$"
json_extraction: false`

	validFunnelConfig := `name: "Test Funnel"
steps:
  - name: "Step 1"
    event_pattern: "login"`

	tests := []struct {
		name          string
		parserFile    string
		funnelFile    string
		parserContent string
		funnelContent string
		expectError   bool
	}{
		{
			name:          "valid_parser_loads_correctly",
			parserFile:    "parser.yaml",
			parserContent: validParserConfig,
			expectError:   false,
		},
		{
			name:          "valid_funnel_loads_correctly",
			funnelFile:    "funnel.yaml",
			funnelContent: validFunnelConfig,
			expectError:   false,
		},
		{
			name:        "nonexistent_parser_fails",
			parserFile:  "nonexistent.yaml",
			expectError: true,
		},
		{
			name:        "nonexistent_funnel_fails",
			funnelFile:  "nonexistent.yaml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parserPath, funnelPath string

			if tt.parserFile != "" {
				parserPath = filepath.Join(tempDir, tt.parserFile)
				if tt.parserContent != "" {
					if err := os.WriteFile(parserPath, []byte(tt.parserContent), 0644); err != nil {
						t.Fatalf("Failed to write parser config: %v", err)
					}
				}
			}

			if tt.funnelFile != "" {
				funnelPath = filepath.Join(tempDir, tt.funnelFile)
				if tt.funnelContent != "" {
					if err := os.WriteFile(funnelPath, []byte(tt.funnelContent), 0644); err != nil {
						t.Fatalf("Failed to write funnel config: %v", err)
					}
				}
			}

			if parserPath != "" {
				_, err := config.LoadParserConfig(parserPath)
				if tt.expectError && err == nil {
					t.Error("Expected parser config loading to fail but it succeeded")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected parser config loading to succeed but it failed: %v", err)
				}
			}

			if funnelPath != "" {
				_, err := config.LoadFunnelConfig(funnelPath)
				if tt.expectError && err == nil {
					t.Error("Expected funnel config loading to fail but it succeeded")
				}
				if !tt.expectError && err != nil {
					t.Errorf("Expected funnel config loading to succeed but it failed: %v", err)
				}
			}
		})
	}
}

func TestValidateCommandFlags(t *testing.T) {
	cmd := createValidateCommand()

	parserFlag := cmd.Flags().Lookup("parser-config")
	if parserFlag == nil {
		t.Error("Expected parser-config flag to exist")
	} else {
		if parserFlag.Shorthand != "p" {
			t.Errorf("Expected parser-config shorthand to be 'p', got %q", parserFlag.Shorthand)
		}
		if parserFlag.Usage != "Path to parser configuration file" {
			t.Errorf("Expected parser-config usage description mismatch")
		}
	}

	funnelFlag := cmd.Flags().Lookup("funnel-config")
	if funnelFlag == nil {
		t.Error("Expected funnel-config flag to exist")
	} else {
		if funnelFlag.Shorthand != "f" {
			t.Errorf("Expected funnel-config shorthand to be 'f', got %q", funnelFlag.Shorthand)
		}
		if funnelFlag.Usage != "Path to funnel configuration file" {
			t.Errorf("Expected funnel-config usage description mismatch")
		}
	}
}

func TestValidateCommandProperties(t *testing.T) {
	cmd := createValidateCommand()

	if cmd.Use != "validate" {
		t.Errorf("Expected Use to be 'validate', got %q", cmd.Use)
	}

	if cmd.Short != "Validate configuration files" {
		t.Errorf("Expected Short description mismatch")
	}

	if !strings.Contains(cmd.Long, "Validate command checks if configuration files") {
		t.Error("Expected Long description to contain validation information")
	}

	if !strings.Contains(cmd.Long, "Examples:") {
		t.Error("Expected Long description to contain examples")
	}

	if cmd.RunE == nil {
		t.Error("RunE function should not be nil")
	}
}

func TestValidateCommandWithValidFile(t *testing.T) {
	tempDir := t.TempDir()

	validConfig := `event_regex: "^(.*)$"
json_extraction: false`

	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cmd := createValidateCommand()
	cmd.SetArgs([]string{"validate", "--parser-config", configPath})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Command should have succeeded: %v", err)
	}
}

func createValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration files",
		Long: `Validate command checks if configuration files are properly formatted
and contain all required fields.

Examples:
  loglion validate --parser-config parser.yaml
  loglion validate --funnel-config funnel.yaml
  loglion validate --parser-config parser.yaml --funnel-config funnel.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			parserConfigFile, _ := cmd.Flags().GetString("parser-config")
			funnelConfigFile, _ := cmd.Flags().GetString("funnel-config")

			if parserConfigFile == "" && funnelConfigFile == "" {
				return cmd.Usage()
			}

			// Validate parser config if specified
			if parserConfigFile != "" {
				_, err := config.LoadParserConfig(parserConfigFile)
				if err != nil {
					return err
				}
			}

			// Validate funnel config if specified
			if funnelConfigFile != "" {
				_, err := config.LoadFunnelConfig(funnelConfigFile)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("parser-config", "p", "", "Path to parser configuration file")
	cmd.Flags().StringP("funnel-config", "f", "", "Path to funnel configuration file")

	return cmd
}

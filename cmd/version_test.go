package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name           string
		expectedOutput string
	}{
		{
			name:           "version_output",
			expectedOutput: Version,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new command instance to avoid interference
			cmd := &cobra.Command{
				Use:   "version",
				Short: "Show version information",
				Long:  `Display version for LogLion.`,
				Run: func(cmd *cobra.Command, args []string) {
					cmd.Print(Version)
				},
			}

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			// Execute the command
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			// Verify output
			output := strings.TrimSpace(buf.String())
			if output != tt.expectedOutput {
				t.Errorf("Expected output %q, got %q", tt.expectedOutput, output)
			}
		})
	}
}

func TestVersionValue(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	// Version should follow semantic versioning pattern (basic check)
	if !strings.Contains(Version, ".") {
		t.Error("Version should contain at least one dot (semantic versioning)")
	}
}

func TestVersionCommandProperties(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("Expected Use to be 'version', got %q", versionCmd.Use)
	}

	if versionCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if versionCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if versionCmd.Run == nil {
		t.Error("Run function should not be nil")
	}
}

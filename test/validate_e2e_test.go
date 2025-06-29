package test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestValidateCommandE2E(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "loglion_test", "../main.go")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Clean up binary after test
	defer func() {
		exec.Command("rm", "-f", "loglion_test").Run()
	}()

	tests := []struct {
		name           string
		args           []string
		expected       []string
		shouldFail     bool
		expectedErrMsg []string
	}{
		{
			name: "validate parser config only",
			args: []string{"validate", "--parser-config", "../examples/simple/simple-parser.yaml"},
			expected: []string{
				"Validating parser config file: ../examples/simple/simple-parser.yaml",
				"✅ Parser configuration is valid!",
				"Event Regex:",
				"JSON Extraction:",
			},
		},
		{
			name: "validate funnel config only",
			args: []string{"validate", "--funnel-config", "../examples/simple/simple-funnel.yaml"},
			expected: []string{
				"Validating funnel config file: ../examples/simple/simple-funnel.yaml",
				"✅ Funnel configuration is valid!",
				"Funnel:",
				"Steps:",
			},
		},
		{
			name: "validate both parser and funnel configs",
			args: []string{"validate", "--parser-config", "../examples/android/logcat-parser.yaml", "--funnel-config", "../examples/android/purchase-funnel.yaml"},
			expected: []string{
				"Validating parser config file: ../examples/android/logcat-parser.yaml",
				"✅ Parser configuration is valid!",
				"Validating funnel config file: ../examples/android/purchase-funnel.yaml",
				"✅ Funnel configuration is valid!",
				"Event Regex:",
				"JSON Extraction:",
				"Funnel:",
				"Steps:",
			},
		},
		{
			name:       "validate with no config files specified",
			args:       []string{"validate"},
			shouldFail: true,
			expectedErrMsg: []string{
				"Error: At least one of --parser-config or --funnel-config must be specified.",
			},
		},
		{
			name:       "validate with non-existent parser config",
			args:       []string{"validate", "--parser-config", "non-existent.yaml"},
			shouldFail: true,
			expectedErrMsg: []string{
				"❌ Parser configuration validation failed:",
				"non-existent.yaml",
			},
		},
		{
			name:       "validate with non-existent funnel config",
			args:       []string{"validate", "--funnel-config", "non-existent.yaml"},
			shouldFail: true,
			expectedErrMsg: []string{
				"❌ Funnel configuration validation failed:",
				"non-existent.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./loglion_test", tt.args...)
			cmd.Dir = "."

			output, err := cmd.CombinedOutput()
			actual := string(output)

			if tt.shouldFail {
				if err == nil {
					t.Fatalf("Expected command to fail, but it succeeded. Output:\n%s", actual)
				}
				for _, expected := range tt.expectedErrMsg {
					if !strings.Contains(actual, expected) {
						t.Errorf("Expected error output to contain %q, but it didn't. Output:\n%s", expected, actual)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("Command failed unexpectedly: %v. Output:\n%s", err, actual)
				}
				for _, expected := range tt.expected {
					if !strings.Contains(actual, expected) {
						t.Errorf("Expected output to contain %q, but it didn't. Output:\n%s", expected, actual)
					}
				}
			}
		})
	}
}

func TestValidateCommandHelp(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "loglion_test", "../main.go")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Clean up binary after test
	defer func() {
		exec.Command("rm", "-f", "loglion_test").Run()
	}()

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name: "validate help with --help flag",
			args: []string{"validate", "--help"},
			expected: []string{
				"Validate command checks if configuration files are properly formatted",
				"Usage:",
				"validate [flags]",
				"Examples:",
				"loglion validate --parser-config parser.yaml",
				"loglion validate --funnel-config funnel.yaml",
				"loglion validate --parser-config parser.yaml --funnel-config funnel.yaml",
				"Flags:",
				"-f, --funnel-config string",
				"-h, --help",
				"-p, --parser-config string",
			},
		},
		{
			name: "validate help with -h flag",
			args: []string{"validate", "-h"},
			expected: []string{
				"Validate command checks if configuration files are properly formatted",
				"Usage:",
				"validate [flags]",
				"Flags:",
				"-f, --funnel-config string",
				"-h, --help",
				"-p, --parser-config string",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./loglion_test", tt.args...)
			cmd.Dir = "."

			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			actual := string(output)
			for _, expected := range tt.expected {
				if !strings.Contains(actual, expected) {
					t.Errorf("Expected output to contain %q, but it didn't. Output:\n%s", expected, actual)
				}
			}
		})
	}
}

func TestValidateCommandWithInvalidYAML(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "loglion_test", "../main.go")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Clean up binary after test
	defer func() {
		exec.Command("rm", "-f", "loglion_test").Run()
	}()

	// Create temporary invalid YAML files for testing
	invalidParserYAML := `
event_regex: "[invalid"
json_extraction: false
`
	invalidFunnelYAML := `
name: ""
steps:
  - name: "Step 1"
    event_pattern: "login"
`

	// Write temporary files
	tmpParserFile := "test_invalid_parser.yaml"
	tmpFunnelFile := "test_invalid_funnel.yaml"

	if err := os.WriteFile(tmpParserFile, []byte(invalidParserYAML), 0644); err != nil {
		t.Fatalf("Failed to create temporary parser file: %v", err)
	}
	defer os.Remove(tmpParserFile)

	if err := os.WriteFile(tmpFunnelFile, []byte(invalidFunnelYAML), 0644); err != nil {
		t.Fatalf("Failed to create temporary funnel file: %v", err)
	}
	defer os.Remove(tmpFunnelFile)

	tests := []struct {
		name           string
		args           []string
		shouldFail     bool
		expectedErrMsg []string
	}{
		{
			name:       "validate invalid parser config",
			args:       []string{"validate", "--parser-config", tmpParserFile},
			shouldFail: true,
			expectedErrMsg: []string{
				"❌ Parser configuration validation failed:",
			},
		},
		{
			name:       "validate invalid funnel config",
			args:       []string{"validate", "--funnel-config", tmpFunnelFile},
			shouldFail: true,
			expectedErrMsg: []string{
				"❌ Funnel configuration validation failed:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./loglion_test", tt.args...)
			cmd.Dir = "."

			output, err := cmd.CombinedOutput()
			actual := string(output)

			if tt.shouldFail {
				if err == nil {
					t.Fatalf("Expected command to fail, but it succeeded. Output:\n%s", actual)
				}
				for _, expected := range tt.expectedErrMsg {
					if !strings.Contains(actual, expected) {
						t.Errorf("Expected error output to contain %q, but it didn't. Output:\n%s", expected, actual)
					}
				}
			}
		})
	}
}

func TestValidateCommandVerboseFlag(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "loglion_test", "../main.go")
	buildCmd.Dir = "."
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Clean up binary after test
	defer func() {
		exec.Command("rm", "-f", "loglion_test").Run()
	}()

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name: "validate with verbose flag",
			args: []string{"--verbose", "validate", "--parser-config", "../examples/simple/simple-parser.yaml"},
			expected: []string{
				"Validating parser config file: ../examples/simple/simple-parser.yaml",
				"✅ Parser configuration is valid!",
			},
		},
		{
			name: "validate with short verbose flag",
			args: []string{"-v", "validate", "--funnel-config", "../examples/simple/simple-funnel.yaml"},
			expected: []string{
				"Validating funnel config file: ../examples/simple/simple-funnel.yaml",
				"✅ Funnel configuration is valid!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./loglion_test", tt.args...)
			cmd.Dir = "."

			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			actual := string(output)
			for _, expected := range tt.expected {
				if !strings.Contains(actual, expected) {
					t.Errorf("Expected output to contain %q, but it didn't. Output:\n%s", expected, actual)
				}
			}
		})
	}
}

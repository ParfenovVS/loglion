package test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCountCommandE2E(t *testing.T) {
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
			name: "count basic events with simple parser",
			args: []string{"count", "--parser-config", "sample/parsers/simple.yaml", "--log", "sample/logs/simple.txt", "login", "logout"},
			expected: []string{
				"📊 Event Count Analysis Complete",
				"Pattern Counts:",
				"login:",
				"logout:",
			},
		},
		{
			name: "count with short flags",
			args: []string{"count", "-p", "sample/parsers/simple.yaml", "-l", "sample/logs/simple.txt", "action"},
			expected: []string{
				"📊 Event Count Analysis Complete",
				"Pattern Counts:",
				"action:",
			},
		},
		{
			name: "count with structured parser",
			args: []string{"count", "--parser-config", "sample/parsers/structured.yaml", "--log", "sample/logs/structured.txt", "login", "purchase"},
			expected: []string{
				"📊 Event Count Analysis Complete",
				"Pattern Counts:",
				"login:",
				"purchase:",
			},
		},
		{
			name: "count with regex patterns",
			args: []string{"count", "--parser-config", "sample/parsers/structured.yaml", "--log", "sample/logs/structured.txt", "user_\\d+", "product_\\d+"},
			expected: []string{
				"📊 Event Count Analysis Complete",
				"Pattern Counts:",
				"user_\\d+:",
				"product_\\d+:",
			},
		},
		{
			name: "count with JSON output format",
			args: []string{"count", "--parser-config", "sample/parsers/simple.yaml", "--log", "sample/logs/simple.txt", "--output", "json", "login"},
			expected: []string{
				`"login"`,
				`"count"`,
			},
		},
		{
			name: "count with short output flag",
			args: []string{"count", "-p", "sample/parsers/simple.yaml", "-l", "sample/logs/simple.txt", "-o", "json", "error"},
			expected: []string{
				`"error"`,
				`"count"`,
			},
		},
		{
			name: "count multiple patterns with structured parser",
			args: []string{"count", "--parser-config", "sample/parsers/structured.yaml", "--log", "sample/logs/structured.txt", "login", "purchase", "logout"},
			expected: []string{
				"📊 Event Count Analysis Complete",
				"Pattern Counts:",
				"login:",
				"purchase:",
                "logout:",
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

	t.Run("count with stdin", func(t *testing.T) {
		cmd := exec.Command("./loglion_test", "count", "-p", "sample/parsers/simple.yaml", "login")
		cmd.Dir = "."
		cmd.Stdin = strings.NewReader("login\nlogout\nlogin")

		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		actual := string(output)
		expected := "login: 2"
		if !strings.Contains(actual, expected) {
			t.Errorf("Expected output to contain %q, but it didn't. Output:\n%s", expected, actual)
		}
	})
}

func TestCountCommandErrorCasesE2E(t *testing.T) {
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
		shouldFail     bool
		expectedErrMsg []string
	}{
		{
			name:       "count with no event patterns",
			args:       []string{"count", "--parser-config", "sample/parsers/simple.yaml", "--log", "sample/logs/simple.txt"},
			shouldFail: true,
			expectedErrMsg: []string{
				"requires at least 1 arg(s)",
			},
		},
		{
			name:       "count with missing parser config",
			args:       []string{"count", "--log", "sample/logs/simple.txt", "login"},
			shouldFail: true,
			expectedErrMsg: []string{
				"required flag(s)",
				"parser-config",
			},
		},
		{
			name:       "count with non-existent parser config",
			args:       []string{"count", "--parser-config", "non-existent.yaml", "--log", "sample/logs/simple.txt", "login"},
			shouldFail: true,
			expectedErrMsg: []string{
				"Error loading parser config:",
				"non-existent.yaml",
			},
		},
		{
			name:       "count with non-existent log file",
			args:       []string{"count", "--parser-config", "sample/parsers/simple.yaml", "--log", "non-existent.txt", "login"},
			shouldFail: true,
			expectedErrMsg: []string{
				"Error parsing log file: open non-existent.txt: no such file or directory",
			},
		},
		{
			name:           "count with invalid output format",
			args:           []string{"count", "--parser-config", "sample/parsers/simple.yaml", "--log", "sample/logs/simple.txt", "--output", "invalid", "login"},
			shouldFail:     false, // Invalid output format defaults to text format
			expectedErrMsg: []string{},
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
				if err != nil && len(tt.expectedErrMsg) > 0 {
					t.Fatalf("Command failed unexpectedly: %v. Output:\n%s", err, actual)
				}
			}
		})
	}
}

func TestCountCommandHelpE2E(t *testing.T) {
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
			name: "count help with --help flag",
			args: []string{"count", "--help"},
			expected: []string{
				"Count command processes log files and counts occurrences of specified event patterns",
				"Usage:",
				"count [event_patterns...]",
				"Examples:",
				"loglion count --parser-config parser.yaml --log logcat.txt",
				"loglion count -p parser.yaml -l logcat.txt --output json",
				"Flags:",
				"-h, --help",
				"-l, --log string",
				"-o, --output string",
				"-p, --parser-config string",
			},
		},
		{
			name: "count help with -h flag",
			args: []string{"count", "-h"},
			expected: []string{
				"Count command processes log files and counts occurrences of specified event patterns",
				"Usage:",
				"count [event_patterns...]",
				"Flags:",
				"-h, --help",
				"-l, --log string",
				"-o, --output string",
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

func TestCountCommandWithInvalidRegexE2E(t *testing.T) {
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

	// Create temporary invalid parser config for testing
	invalidParserYAML := `
event_regex: "^(.*)$"
json_extraction: false
`
	tmpParserFile := "test_parser_count.yaml"

	if err := os.WriteFile(tmpParserFile, []byte(invalidParserYAML), 0644); err != nil {
		t.Fatalf("Failed to create temporary parser file: %v", err)
	}
	defer os.Remove(tmpParserFile)

	tests := []struct {
		name           string
		args           []string
		shouldFail     bool
		expectedErrMsg []string
	}{
		{
			name:       "count with invalid regex pattern",
			args:       []string{"count", "--parser-config", tmpParserFile, "--log", "sample/logs/simple.txt", "[invalid"},
			shouldFail: true,
			expectedErrMsg: []string{
				"Error creating count analyzer:",
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

func TestCountCommandVerboseFlagE2E(t *testing.T) {
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
			name: "count with verbose flag",
			args: []string{"--verbose", "count", "--parser-config", "sample/parsers/simple.yaml", "--log", "sample/logs/simple.txt", "login"},
			expected: []string{
				"📊 Event Count Analysis Complete",
				"Pattern Counts:",
				"login:",
			},
		},
		{
			name: "count with short verbose flag",
			args: []string{"-v", "count", "-p", "sample/parsers/simple.yaml", "-l", "sample/logs/simple.txt", "logout"},
			expected: []string{
				"📊 Event Count Analysis Complete",
				"Pattern Counts:",
				"logout:",
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

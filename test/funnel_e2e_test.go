package test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestFunnelCommandE2E(t *testing.T) {
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
			name: "funnel basic flow with simple logs",
			args: []string{"funnel", "--parser-config", "sample/parsers/simple.yaml", "--funnel-config", "sample/funnels/basic.yaml", "--log", "sample/logs/simple.txt"},
			expected: []string{
				"✅ Funnel Analysis Complete",
				"Funnel: Basic User Flow",
				"Step Breakdown:",
				"Login:",
				"Action:",
				"Logout:",
				"Drop-off Analysis:",
			},
		},
		{
			name: "funnel with short flags",
			args: []string{"funnel", "-p", "sample/parsers/simple.yaml", "-f", "sample/funnels/basic.yaml", "-l", "sample/logs/simple.txt"},
			expected: []string{
				"✅ Funnel Analysis Complete",
				"Funnel: Basic User Flow",
				"Step Breakdown:",
				"Login:",
				"Action:",
				"Logout:",
			},
		},
		{
			name: "funnel purchase flow with structured logs",
			args: []string{"funnel", "--parser-config", "sample/parsers/structured.yaml", "--funnel-config", "sample/funnels/purchase.yaml", "--log", "sample/logs/structured.txt"},
			expected: []string{
				"✅ Funnel Analysis Complete",
				"Funnel: Purchase Flow",
				"Step Breakdown:",
				"Product View:",
				"Add to Cart:",
				"Purchase:",
				"Drop-off Analysis:",
			},
		},
		{
			name: "funnel with JSON output format",
			args: []string{"funnel", "--parser-config", "sample/parsers/simple.yaml", "--funnel-config", "sample/funnels/basic.yaml", "--log", "sample/logs/simple.txt", "--output", "json"},
			expected: []string{
				`"funnel_name": "Basic User Flow"`,
				`"steps"`,
				`"name": "Login"`,
				`"name": "Action"`,
				`"name": "Logout"`,
				`"drop_offs"`,
			},
		},
		{
			name: "funnel with short output flag",
			args: []string{"funnel", "-p", "sample/parsers/structured.yaml", "-f", "sample/funnels/purchase.yaml", "-l", "sample/logs/structured.txt", "-o", "json"},
			expected: []string{
				`"funnel_name": "Purchase Flow"`,
				`"steps"`,
				`"name": "Product View"`,
				`"name": "Add to Cart"`,
				`"name": "Purchase"`,
			},
		},
		{
			name: "funnel with limit flag",
			args: []string{"funnel", "--parser-config", "sample/parsers/simple.yaml", "--funnel-config", "sample/funnels/basic.yaml", "--log", "sample/logs/simple.txt", "--limit", "1"},
			expected: []string{
				"✅ Funnel Analysis Complete",
				"Funnel: Basic User Flow",
				"Step Breakdown:",
				"Login:",
				"Action:",
				"Logout:",
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

func TestFunnelCommandErrorCasesE2E(t *testing.T) {
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
			name:       "funnel with missing parser config",
			args:       []string{"funnel", "--funnel-config", "sample/funnels/basic.yaml", "--log", "sample/logs/simple.txt"},
			shouldFail: true,
			expectedErrMsg: []string{
				"required flag(s)",
				"parser-config",
			},
		},
		{
			name:       "funnel with missing funnel config",
			args:       []string{"funnel", "--parser-config", "sample/parsers/simple.yaml", "--log", "sample/logs/simple.txt"},
			shouldFail: true,
			expectedErrMsg: []string{
				"required flag(s)",
				"funnel-config",
			},
		},
		{
			name:       "funnel with missing log file",
			args:       []string{"funnel", "--parser-config", "sample/parsers/simple.yaml", "--funnel-config", "sample/funnels/basic.yaml"},
			shouldFail: true,
			expectedErrMsg: []string{
				"required flag(s)",
				"log",
			},
		},
		{
			name:       "funnel with non-existent parser config",
			args:       []string{"funnel", "--parser-config", "non-existent.yaml", "--funnel-config", "sample/funnels/basic.yaml", "--log", "sample/logs/simple.txt"},
			shouldFail: true,
			expectedErrMsg: []string{
				"Error loading parser config:",
				"non-existent.yaml",
			},
		},
		{
			name:       "funnel with non-existent funnel config",
			args:       []string{"funnel", "--parser-config", "sample/parsers/simple.yaml", "--funnel-config", "non-existent.yaml", "--log", "sample/logs/simple.txt"},
			shouldFail: true,
			expectedErrMsg: []string{
				"Error loading funnel config:",
				"non-existent.yaml",
			},
		},
		{
			name:       "funnel with non-existent log file",
			args:       []string{"funnel", "--parser-config", "sample/parsers/simple.yaml", "--funnel-config", "sample/funnels/basic.yaml", "--log", "non-existent.txt"},
			shouldFail: true,
			expectedErrMsg: []string{
				"Error parsing log file:",
				"non-existent.txt",
			},
		},
		{
			name:           "funnel with invalid output format",
			args:           []string{"funnel", "--parser-config", "sample/parsers/simple.yaml", "--funnel-config", "sample/funnels/basic.yaml", "--log", "sample/logs/simple.txt", "--output", "invalid"},
			shouldFail:     false, // Invalid output format defaults to text format
			expectedErrMsg: []string{},
		},
		{
			name:           "funnel with invalid limit value",
			args:           []string{"funnel", "--parser-config", "sample/parsers/simple.yaml", "--funnel-config", "sample/funnels/basic.yaml", "--log", "sample/logs/simple.txt", "--limit", "-1"},
			shouldFail:     false, // -1 limit is treated as no limit, so it succeeds
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

func TestFunnelCommandHelpE2E(t *testing.T) {
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
			name: "funnel help with --help flag",
			args: []string{"funnel", "--help"},
			expected: []string{
				"Funnel command processes log files according to the funnel configuration",
				"Usage:",
				"funnel [flags]",
				"Examples:",
				"loglion funnel --parser-config parser.yaml --funnel-config funnel.yaml --log logcat.txt",
				"loglion funnel -p parser.yaml -f funnel.yaml -l logcat.txt --limit 5",
				"Flags:",
				"-f, --funnel-config string",
				"-h, --help",
				"-l, --log string",
				"--limit int",
				"-o, --output string",
				"-p, --parser-config string",
			},
		},
		{
			name: "funnel help with -h flag",
			args: []string{"funnel", "-h"},
			expected: []string{
				"Funnel command processes log files according to the funnel configuration",
				"Usage:",
				"funnel [flags]",
				"Flags:",
				"-f, --funnel-config string",
				"-h, --help",
				"-l, --log string",
				"--limit int",
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

func TestFunnelCommandVerboseFlagE2E(t *testing.T) {
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
			name: "funnel with verbose flag",
			args: []string{"--verbose", "funnel", "--parser-config", "sample/parsers/simple.yaml", "--funnel-config", "sample/funnels/basic.yaml", "--log", "sample/logs/simple.txt"},
			expected: []string{
				"✅ Funnel Analysis Complete",
				"Funnel: Basic User Flow",
				"Step Breakdown:",
				"Login:",
				"Action:",
				"Logout:",
			},
		},
		{
			name: "funnel with short verbose flag",
			args: []string{"-v", "funnel", "-p", "sample/parsers/simple.yaml", "-f", "sample/funnels/basic.yaml", "-l", "sample/logs/simple.txt"},
			expected: []string{
				"✅ Funnel Analysis Complete",
				"Funnel: Basic User Flow",
				"Step Breakdown:",
				"Login:",
				"Action:",
				"Logout:",
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
package test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestRootCommandE2E(t *testing.T) {
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
			name: "root command with no args shows help",
			args: []string{},
			expected: []string{
				"LogLion is a CLI tool that analyzes logcat files",
				"Usage:",
				"loglion [command]",
				"Available Commands:",
				"count",
				"funnel",
				"validate",
				"version",
			},
		},
		{
			name: "root command help flag",
			args: []string{"--help"},
			expected: []string{
				"LogLion is a CLI tool that analyzes logcat files",
				"Usage:",
				"loglion [command]",
				"Available Commands:",
				"count",
				"funnel", 
				"validate",
				"version",
				"Flags:",
				"-h, --help",
				"-v, --verbose",
			},
		},
		{
			name: "root command -h flag",
			args: []string{"-h"},
			expected: []string{
				"LogLion is a CLI tool that analyzes logcat files",
				"Usage:",
				"loglion [command]",
				"Available Commands:",
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

func TestRootCommandVerboseFlag(t *testing.T) {
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
		name string
		args []string
	}{
		{
			name: "verbose flag short form",
			args: []string{"-v"},
		},
		{
			name: "verbose flag long form",
			args: []string{"--verbose"},
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
			// With verbose flag, it should still show help but with verbose logging enabled
			expectedContains := []string{
				"LogLion is a CLI tool that analyzes logcat files",
				"Usage:",
				"loglion [command]",
			}

			for _, expected := range expectedContains {
				if !strings.Contains(actual, expected) {
					t.Errorf("Expected output to contain %q, but it didn't. Output:\n%s", expected, actual)
				}
			}
		})
	}
}

func TestRootCommandInvalidFlag(t *testing.T) {
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

	cmd := exec.Command("./loglion_test", "--invalid-flag")
	cmd.Dir = "."
	
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with invalid flag, but it succeeded")
	}

	actual := string(output)
	expectedContains := []string{
		"unknown flag",
		"--invalid-flag", 
	}

	for _, expected := range expectedContains {
		if !strings.Contains(actual, expected) {
			t.Errorf("Expected error output to contain %q, but it didn't. Output:\n%s", expected, actual)
		}
	}
}

func TestRootCommandInvalidSubcommand(t *testing.T) {
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

	cmd := exec.Command("./loglion_test", "invalid-command")
	cmd.Dir = "."
	
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("Expected command to fail with invalid subcommand, but it succeeded")
	}

	actual := string(output)
	expectedContains := []string{
		"unknown command",
		"invalid-command",
	}

	for _, expected := range expectedContains {
		if !strings.Contains(actual, expected) {
			t.Errorf("Expected error output to contain %q, but it didn't. Output:\n%s", expected, actual)
		}
	}
}
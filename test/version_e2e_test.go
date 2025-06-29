package test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestVersionCommandE2E(t *testing.T) {
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
		expected string
	}{
		{
			name:     "version command",
			args:     []string{"version"},
			expected: "0.1.4",
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

			actual := strings.TrimSpace(string(output))
			if actual != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestVersionCommandHelp(t *testing.T) {
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

	cmd := exec.Command("./loglion_test", "version", "--help")
	cmd.Dir = "."

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	actual := string(output)
	expectedContains := []string{
		"Display version for LogLion",
		"Usage:",
		"version",
	}

	for _, expected := range expectedContains {
		if !strings.Contains(actual, expected) {
			t.Errorf("Expected output to contain %q, but it didn't. Output: %s", expected, actual)
		}
	}
}

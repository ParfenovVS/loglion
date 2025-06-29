package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func TestRootCommandProperties(t *testing.T) {
	if rootCmd.Use != "loglion" {
		t.Errorf("Expected Use to be 'loglion', got %q", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if rootCmd.Long == "" {
		t.Error("Long description should not be empty")
	}

	if rootCmd.PersistentPreRun == nil {
		t.Error("PersistentPreRun function should not be nil")
	}
}

func TestRootCommandFlags(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("verbose")
	if flag == nil {
		t.Error("verbose flag should be defined")
	}

	if flag.Shorthand != "v" {
		t.Errorf("Expected verbose flag shorthand to be 'v', got %q", flag.Shorthand)
	}

	if flag.DefValue != "false" {
		t.Errorf("Expected verbose flag default to be 'false', got %q", flag.DefValue)
	}
}

func TestSetupLogging(t *testing.T) {
	tests := []struct {
		name          string
		verboseFlag   bool
		expectedLevel logrus.Level
	}{
		{
			name:          "verbose_enabled",
			verboseFlag:   true,
			expectedLevel: logrus.DebugLevel,
		},
		{
			name:          "verbose_disabled",
			verboseFlag:   false,
			expectedLevel: logrus.PanicLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			originalLevel := logrus.GetLevel()
			originalVerbose := verbose

			// Set test values
			verbose = tt.verboseFlag

			// Call setupLogging
			setupLogging()

			// Check level
			if logrus.GetLevel() != tt.expectedLevel {
				t.Errorf("Expected log level %v, got %v", tt.expectedLevel, logrus.GetLevel())
			}

			// Restore original values
			logrus.SetLevel(originalLevel)
			verbose = originalVerbose
		})
	}
}

func TestSetupLoggingFormatter(t *testing.T) {
	// Save original values
	originalVerbose := verbose
	originalFormatter := logrus.StandardLogger().Formatter

	// Test verbose mode formatter
	verbose = true
	setupLogging()

	formatter, ok := logrus.StandardLogger().Formatter.(*logrus.TextFormatter)
	if !ok {
		t.Error("Expected TextFormatter when verbose is enabled")
	} else {
		if !formatter.ForceColors {
			t.Error("Expected ForceColors to be true when verbose is enabled")
		}
		if !formatter.FullTimestamp {
			t.Error("Expected FullTimestamp to be true when verbose is enabled")
		}
	}

	// Restore original values
	verbose = originalVerbose
	logrus.StandardLogger().Formatter = originalFormatter
}

func TestExecuteSuccess(t *testing.T) {
	// Create a simple test command
	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			// Do nothing, just succeed
		},
	}

	// Replace rootCmd temporarily
	originalRootCmd := rootCmd
	rootCmd = testCmd

	// Capture any output
	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)

	// Test Execute function
	Execute()

	// Restore original rootCmd
	rootCmd = originalRootCmd
}

func TestExecuteError(t *testing.T) {
	// Create a command that returns an error
	testCmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("test error")
		},
	}

	// Replace rootCmd temporarily
	originalRootCmd := rootCmd
	rootCmd = testCmd

	// Capture output
	var buf bytes.Buffer
	testCmd.SetOut(&buf)
	testCmd.SetErr(&buf)

	// Test that Execute function would call os.Exit(1) on error
	// We can't actually test os.Exit without complex mocking
	// So we test that rootCmd.Execute() returns an error
	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error from Execute, got nil")
	}

	// Restore original values
	rootCmd = originalRootCmd
}

func TestPersistentPreRun(t *testing.T) {
	// Save original values
	originalVerbose := verbose
	originalLevel := logrus.GetLevel()

	tests := []struct {
		name     string
		verbose  bool
		expected logrus.Level
	}{
		{
			name:     "verbose_true",
			verbose:  true,
			expected: logrus.DebugLevel,
		},
		{
			name:     "verbose_false",
			verbose:  false,
			expected: logrus.PanicLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test values
			verbose = tt.verbose

			// Call PersistentPreRun
			rootCmd.PersistentPreRun(rootCmd, []string{})

			// Verify logging was set up correctly
			if logrus.GetLevel() != tt.expected {
				t.Errorf("Expected log level %v, got %v", tt.expected, logrus.GetLevel())
			}
		})
	}

	// Restore original values
	verbose = originalVerbose
	logrus.SetLevel(originalLevel)
}

func TestRootCommandHelp(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	// Execute help
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	output := buf.String()

	// Check that help contains expected content
	expectedStrings := []string{
		"loglion",
		"LogLion",
		"Usage:",
		"Available Commands:",
		"Flags:",
		"-v, --verbose",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Help output should contain %q", expected)
		}
	}
}

func TestVerboseFlagParsing(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "short_flag",
			args:     []string{"-v"},
			expected: true,
		},
		{
			name:     "long_flag",
			args:     []string{"--verbose"},
			expected: true,
		},
		{
			name:     "no_flag",
			args:     []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalVerbose := verbose

			// Reset verbose flag
			verbose = false

			// Create a test command with the same flag structure
			testCmd := &cobra.Command{
				Use: "test",
				PersistentPreRun: func(cmd *cobra.Command, args []string) {
					setupLogging()
				},
				Run: func(cmd *cobra.Command, args []string) {
					// Do nothing
				},
			}
			testCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

			testCmd.SetArgs(tt.args)
			err := testCmd.Execute()
			if err != nil {
				t.Fatalf("Command execution failed: %v", err)
			}

			if verbose != tt.expected {
				t.Errorf("Expected verbose to be %v, got %v", tt.expected, verbose)
			}

			// Restore original value
			verbose = originalVerbose
		})
	}
}

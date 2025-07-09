package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestFunnelCommandFlags(t *testing.T) {
	cmd := funnelCmd

	// Test parser-config flag
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

	// Test funnel-config flag
	funnelFlag := cmd.Flags().Lookup("funnel-config")
	if funnelFlag == nil {
		t.Error("Expected funnel-config flag to exist")
	} else {
		if funnelFlag.Shorthand != "f" {
			t.Errorf("Expected funnel-config shorthand to be 'f', got %q", funnelFlag.Shorthand)
		}
		if funnelFlag.Usage != "Path to funnel configuration file (required)" {
			t.Errorf("Expected funnel-config usage description mismatch")
		}
	}

	// Test log flag
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

	// Test output flag
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
		if outputFlag.DefValue != "text" {
			t.Errorf("Expected output default value to be 'text', got %q", outputFlag.DefValue)
		}
	}

	// Test limit flag
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("Expected limit flag to exist")
	} else {
		if limitFlag.Usage != "Maximum number of successful funnels to analyze (0 = analyze all funnels)" {
			t.Errorf("Expected limit usage description mismatch")
		}
		if limitFlag.DefValue != "0" {
			t.Errorf("Expected limit default value to be '0', got %q", limitFlag.DefValue)
		}
	}
}

func TestFunnelCommandProperties(t *testing.T) {
	cmd := funnelCmd

	if cmd.Use != "funnel" {
		t.Errorf("Expected Use to be 'funnel', got %q", cmd.Use)
	}

	if cmd.Short != "Analyze log files for funnel validation" {
		t.Errorf("Expected Short description mismatch")
	}

	if !strings.Contains(cmd.Long, "Funnel command processes log files according to the funnel configuration") {
		t.Error("Expected Long description to contain funnel processing information")
	}

	if !strings.Contains(cmd.Long, "Examples:") {
		t.Error("Expected Long description to contain examples")
	}

	if !strings.Contains(cmd.Long, "loglion funnel --parser-config parser.yaml --funnel-config funnel.yaml --log logcat.txt") {
		t.Error("Expected Long description to contain complete example")
	}

	if cmd.Run == nil {
		t.Error("Run function should not be nil")
	}
}

func TestFunnelCommandRequiredFlags(t *testing.T) {
	cmd := funnelCmd

	// Check if required flags are marked as required
	requiredFlags := []string{"parser-config", "funnel-config"}
	
	for _, flagName := range requiredFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Required flag %s not found", flagName)
			continue
		}
		
		// Check if flag is in required flags list
		requiredAnnotation := flag.Annotations[cobra.BashCompOneRequiredFlag]
		if len(requiredAnnotation) == 0 {
			t.Errorf("Flag %s should be marked as required", flagName)
		}
	}
}

func TestFunnelCommandFlagTypes(t *testing.T) {
	cmd := funnelCmd

	// Test string flags
	stringFlags := map[string]string{
		"parser-config": "",
		"funnel-config": "",
		"log":           "",
		"output":        "text",
	}

	for flagName, expectedDefault := range stringFlags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Flag %s not found", flagName)
			continue
		}
		
		if flag.Value.Type() != "string" {
			t.Errorf("Expected flag %s to be of type string, got %s", flagName, flag.Value.Type())
		}
		
		if flag.DefValue != expectedDefault {
			t.Errorf("Expected flag %s default value to be %q, got %q", flagName, expectedDefault, flag.DefValue)
		}
	}

	// Test int flag
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("Limit flag not found")
	} else {
		if limitFlag.Value.Type() != "int" {
			t.Errorf("Expected limit flag to be of type int, got %s", limitFlag.Value.Type())
		}
		if limitFlag.DefValue != "0" {
			t.Errorf("Expected limit flag default value to be '0', got %q", limitFlag.DefValue)
		}
	}
}

func TestFunnelCommandHelpText(t *testing.T) {
	cmd := funnelCmd

	// Test that help text contains key information
	helpText := cmd.Long

	expectedPhrases := []string{
		"Funnel command processes log files",
		"outputs completion rates",
		"drop-off analysis",
		"Examples:",
		"--parser-config",
		"--funnel-config",
		"--log",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(helpText, phrase) {
			t.Errorf("Expected help text to contain %q", phrase)
		}
	}
}

func TestFunnelCommandFlagShorthands(t *testing.T) {
	cmd := funnelCmd

	expectedShorthands := map[string]string{
		"parser-config": "p",
		"funnel-config": "f",
		"log":           "l",
		"output":        "o",
	}

	for flagName, expectedShorthand := range expectedShorthands {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Flag %s not found", flagName)
			continue
		}
		
		if flag.Shorthand != expectedShorthand {
			t.Errorf("Expected flag %s shorthand to be %q, got %q", flagName, expectedShorthand, flag.Shorthand)
		}
	}
}

func TestFunnelCommandStructure(t *testing.T) {
	cmd := funnelCmd

	// Test that command is properly structured
	if cmd.Use == "" {
		t.Error("Command Use should not be empty")
	}
	
	if cmd.Short == "" {
		t.Error("Command Short description should not be empty")
	}
	
	if cmd.Long == "" {
		t.Error("Command Long description should not be empty")
	}
	
	if cmd.Run == nil {
		t.Error("Command Run function should not be nil")
	}
	
	// Test that required flags are present
	flags := cmd.Flags()
	if flags == nil {
		t.Error("Command should have flags")
	}
	
	flagCount := 0
	flags.VisitAll(func(flag *pflag.Flag) {
		flagCount++
	})
	
	if flagCount < 5 {
		t.Errorf("Expected at least 5 flags, got %d", flagCount)
	}
}

func TestFunnelCommandExamples(t *testing.T) {
	cmd := funnelCmd

	// Check that examples are present and properly formatted
	if !strings.Contains(cmd.Long, "Examples:") {
		t.Error("Command should contain examples section")
	}

	// Check for specific example patterns
	expectedExamples := []string{
		"loglion funnel --parser-config parser.yaml --funnel-config funnel.yaml --log logcat.txt",
		"loglion funnel -p parser.yaml -f funnel.yaml -l logcat.txt --limit 5",
	}

	for _, example := range expectedExamples {
		if !strings.Contains(cmd.Long, example) {
			t.Errorf("Expected to find example: %s", example)
		}
	}
}
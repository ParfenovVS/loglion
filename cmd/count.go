package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/parfenovvs/loglion/internal/analyzer"
	"github.com/parfenovvs/loglion/internal/config"
	"github.com/parfenovvs/loglion/internal/output"
	"github.com/parfenovvs/loglion/internal/parser"
)

var countCmd = &cobra.Command{
	Use:   "count [event_patterns...]",
	Short: "Count occurrences of event patterns in log files",
	Long: `Count command processes log files and counts occurrences of specified event patterns.
It accepts multiple event patterns as arguments and outputs the count for each pattern.

Examples:
  loglion count --parser-config parser.yaml --log logcat.txt "login" "logout" "error"
  loglion count -p parser.yaml -l logcat.txt --output json "user_action" "network_request"
  loglion count -p parser.yaml -l logcat.txt "memory_warning"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		parserConfigFile, _ := cmd.Flags().GetString("parser-config")
		logFile, _ := cmd.Flags().GetString("log")
		outputFormat, _ := cmd.Flags().GetString("output")

		logrus.WithFields(logrus.Fields{
			"parser_config_file": parserConfigFile,
			"log_file":           logFile,
			"output_format":      outputFormat,
			"event_patterns":     args,
		}).Info("Starting count analysis")

		// Load parser configuration
		logrus.Debug("Loading parser configuration file")
		parserCfg, err := config.LoadParserConfig(parserConfigFile)
		if err != nil {
			logrus.WithError(err).WithField("parser_config_file", parserConfigFile).Error("Failed to load parser config")
			fmt.Fprintf(os.Stderr, "Error loading parser config: %v\n", err)
			os.Exit(1)
		}

		// Create parser
		logrus.Debug("Creating log parser")
		logParser := parser.NewParserWithConfig(
			parserCfg.TimestampFormat,
			parserCfg.EventRegex,
			parserCfg.JSONExtraction,
			parserCfg.LogLineRegex)

		// Create count analyzer
		logrus.Debug("Creating count analyzer")
		countAnalyzer, err := analyzer.NewCountAnalyzer(args)
		if err != nil {
			logrus.WithError(err).Error("Failed to create count analyzer")
			fmt.Fprintf(os.Stderr, "Error creating count analyzer: %v\n", err)
			os.Exit(1)
		}

		// Parse log file
		logrus.WithField("log_file", logFile).Debug("Starting log file parsing")
		entries, err := logParser.ParseFile(logFile)
		if err != nil {
			logrus.WithError(err).WithField("log_file", logFile).Error("Failed to parse log file")
			fmt.Fprintf(os.Stderr, "Error parsing log file: %v\n", err)
			os.Exit(1)
		}

		logrus.Debug("Starting count analysis")
		result := countAnalyzer.AnalyzeCount(entries)

		// Format and output results
		logrus.WithField("output_format", outputFormat).Debug("Creating output formatter")
		var formatter output.Formatter
		switch outputFormat {
		case "json":
			formatter = output.NewFormatter(output.JSONFormat)
		default:
			formatter = output.NewFormatter(output.TextFormat)
		}

		logrus.Debug("Formatting count analysis results")
		formattedOutput, err := formatter.FormatCount(result)
		if err != nil {
			logrus.WithError(err).Error("Failed to format count analysis output")
			fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
			os.Exit(1)
		}

		logrus.WithField("output_length", len(formattedOutput)).Info("Count analysis completed successfully")
		fmt.Print(formattedOutput)
	},
}

func init() {
	rootCmd.AddCommand(countCmd)

	countCmd.Flags().StringP("parser-config", "p", "", "Path to parser configuration file (required)")
	countCmd.Flags().StringP("log", "l", "", "Path to log file (required)")
	countCmd.Flags().StringP("output", "o", "text", "Output format (json, text)")

	countCmd.MarkFlagRequired("parser-config")
	countCmd.MarkFlagRequired("log")
}
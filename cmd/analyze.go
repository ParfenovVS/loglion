package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"loglion/internal/analyzer"
	"loglion/internal/config"
	"loglion/internal/output"
	"loglion/internal/parser"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze log files for funnel validation",
	Long: `Analyze command processes log files according to the funnel configuration
and outputs completion rates and drop-off analysis.

Example:
  loglion analyze --config funnel.yaml --log logcat.txt --format logcat-plain`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		logFile, _ := cmd.Flags().GetString("log")
		format, _ := cmd.Flags().GetString("format")
		outputFormat, _ := cmd.Flags().GetString("output")

		logrus.WithFields(logrus.Fields{
			"config_file":   configFile,
			"log_file":      logFile,
			"format":        format,
			"output_format": outputFormat,
		}).Info("Starting funnel analysis")

		// Load configuration
		logrus.Debug("Loading configuration file")
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			logrus.WithError(err).WithField("config_file", configFile).Error("Failed to load config")
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Validate format matches config (with backward compatibility)
		logrus.WithFields(logrus.Fields{
			"config_format": cfg.Format,
			"flag_format":   format,
		}).Debug("Validating format consistency")

		if cfg.Format != format {
			logrus.WithFields(logrus.Fields{
				"config_format": cfg.Format,
				"flag_format":   format,
			}).Error("Format mismatch between config and flag")
			fmt.Fprintf(os.Stderr, "Format mismatch: config specifies '%s', but flag specifies '%s'\n", cfg.Format, format)
			os.Exit(1)
		}

		// Create parser
		logrus.WithField("format", format).Debug("Creating log parser")
		var logParser parser.Parser
		
		
		switch format {
		case "logcat-plain":
			logParser = parser.NewParserWithConfig(
				parser.LogcatPlainFormat,
				cfg.LogParser.TimestampFormat,
				cfg.LogParser.EventRegex,
				cfg.LogParser.JSONExtraction)
		case "logcat-json":
			logParser = parser.NewParserWithConfig(
				parser.LogcatJSONFormat,
				cfg.LogParser.TimestampFormat,
				cfg.LogParser.EventRegex,
				cfg.LogParser.JSONExtraction)
		default:
			logrus.WithField("format", format).Error("Unsupported log format")
			fmt.Fprintf(os.Stderr, "Unsupported format: %s (supported: logcat-plain, logcat-json)\n", format)
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

		// Create analyzer and analyze funnel
		logrus.Debug("Creating funnel analyzer")
		funnelAnalyzer := analyzer.NewFunnelAnalyzer(cfg)

		logrus.Debug("Starting funnel analysis")
		result := funnelAnalyzer.AnalyzeFunnel(entries)

		// Format and output results
		logrus.WithField("output_format", outputFormat).Debug("Creating output formatter")
		var formatter output.Formatter
		switch outputFormat {
		case "json":
			formatter = output.NewFormatter(output.JSONFormat)
		default:
			formatter = output.NewFormatter(output.TextFormat)
		}

		logrus.Debug("Formatting analysis results")
		formattedOutput, err := formatter.Format(result)
		if err != nil {
			logrus.WithError(err).Error("Failed to format analysis output")
			fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
			os.Exit(1)
		}

		logrus.WithField("output_length", len(formattedOutput)).Info("Analysis completed successfully")
		fmt.Print(formattedOutput)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringP("config", "c", "", "Path to funnel configuration file (required)")
	analyzeCmd.Flags().StringP("log", "l", "", "Path to log file (required)")
	analyzeCmd.Flags().StringP("format", "f", "logcat-plain", "Log format (logcat-plain, logcat-json)")
	analyzeCmd.Flags().StringP("output", "o", "text", "Output format (json, text)")

	analyzeCmd.MarkFlagRequired("config")
	analyzeCmd.MarkFlagRequired("log")
}

package cmd

import (
	"fmt"
	"os"

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
  loglion analyze --config funnel.yaml --log logcat.txt --format android`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		logFile, _ := cmd.Flags().GetString("log")
		format, _ := cmd.Flags().GetString("format")
		outputFormat, _ := cmd.Flags().GetString("output")

		// Load configuration
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Validate format matches config
		if cfg.Format != format {
			fmt.Fprintf(os.Stderr, "Format mismatch: config specifies '%s', but flag specifies '%s'\n", cfg.Format, format)
			os.Exit(1)
		}

		// Create parser
		var logParser parser.Parser
		switch format {
		case "android":
			logParser = parser.NewAndroidParserWithConfig(
				cfg.AndroidParser.TimestampFormat,
				cfg.AndroidParser.EventRegex,
				cfg.AndroidParser.JSONExtraction)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported format: %s\n", format)
			os.Exit(1)
		}

		// Parse log file
		entries, err := logParser.ParseFile(logFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing log file: %v\n", err)
			os.Exit(1)
		}

		// Create analyzer and analyze funnel
		funnelAnalyzer := analyzer.NewFunnelAnalyzer(cfg)
		result := funnelAnalyzer.AnalyzeFunnel(entries)

		// Format and output results
		var formatter output.Formatter
		switch outputFormat {
		case "json":
			formatter = output.NewFormatter(output.JSONFormat)
		default:
			formatter = output.NewFormatter(output.TextFormat)
		}

		formattedOutput, err := formatter.Format(result)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(formattedOutput)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringP("config", "c", "", "Path to funnel configuration file (required)")
	analyzeCmd.Flags().StringP("log", "l", "", "Path to log file (required)")
	analyzeCmd.Flags().StringP("format", "f", "android", "Log format preset")
	analyzeCmd.Flags().StringP("output", "o", "text", "Output format (json, text)")
	analyzeCmd.Flags().IntP("timeout", "t", 30, "Session timeout in minutes")

	analyzeCmd.MarkFlagRequired("config")
	analyzeCmd.MarkFlagRequired("log")
}

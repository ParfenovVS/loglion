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

var funnelCmd = &cobra.Command{
	Use:   "funnel",
	Short: "Analyze log files for funnel validation",
	Long: `Funnel command processes log files according to the funnel configuration
and outputs completion rates and drop-off analysis.

Examples:
  loglion funnel --parser-config parser.yaml --funnel-config funnel.yaml --log logcat.txt
  loglion funnel -p parser.yaml -f funnel.yaml -l logcat.txt --max 5`,
	Run: func(cmd *cobra.Command, args []string) {
		parserConfigFile, _ := cmd.Flags().GetString("parser-config")
		funnelConfigFile, _ := cmd.Flags().GetString("funnel-config")
		logFile, _ := cmd.Flags().GetString("log")
		outputFormat, _ := cmd.Flags().GetString("output")
		max, _ := cmd.Flags().GetInt("max")

		logrus.WithFields(logrus.Fields{
			"parser_config_file": parserConfigFile,
			"funnel_config_file": funnelConfigFile,
			"log_file":           logFile,
			"output_format":      outputFormat,
			"max":                max,
		}).Info("Starting funnel analysis")

		// Load parser configuration
		logrus.Debug("Loading parser configuration file")
		parserCfg, err := config.LoadParserConfig(parserConfigFile)
		if err != nil {
			logrus.WithError(err).WithField("parser_config_file", parserConfigFile).Error("Failed to load parser config")
			fmt.Fprintf(os.Stderr, "Error loading parser config: %v\n", err)
			os.Exit(1)
		}

		// Load funnel configuration
		logrus.Debug("Loading funnel configuration file")
		funnelCfg, err := config.LoadFunnelConfig(funnelConfigFile)
		if err != nil {
			logrus.WithError(err).WithField("funnel_config_file", funnelConfigFile).Error("Failed to load funnel config")
			fmt.Fprintf(os.Stderr, "Error loading funnel config: %v\n", err)
			os.Exit(1)
		}

		// Create parser
		logrus.Debug("Creating log parser")
		logParser := parser.NewParserWithConfig(
			parserCfg.TimestampFormat,
			parserCfg.EventRegex,
			parserCfg.JSONExtraction,
			parserCfg.LogLineRegex)

		// Create analyzer
		logrus.Debug("Creating funnel analyzer")
		funnelAnalyzer := analyzer.NewFunnelAnalyzer(funnelCfg)

		// Parse log file
		logrus.WithField("log_file", logFile).Debug("Starting log file parsing")
		entries, err := logParser.ParseFile(logFile)
		if err != nil {
			logrus.WithError(err).WithField("log_file", logFile).Error("Failed to parse log file")
			fmt.Fprintf(os.Stderr, "Error parsing log file: %v\n", err)
			os.Exit(1)
		}

		logrus.Debug("Starting funnel analysis")
		result := funnelAnalyzer.AnalyzeFunnel(entries, max)

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
	rootCmd.AddCommand(funnelCmd)

	funnelCmd.Flags().StringP("parser-config", "p", "", "Path to parser configuration file (required)")
	funnelCmd.Flags().StringP("funnel-config", "f", "", "Path to funnel configuration file (required)")
	funnelCmd.Flags().StringP("log", "l", "", "Path to log file (required)")
	funnelCmd.Flags().StringP("output", "o", "text", "Output format (json, text)")
	funnelCmd.Flags().IntP("max", "m", 0, "Maximum number of successful funnels to analyze (0 = analyze all funnels)")

	funnelCmd.MarkFlagRequired("parser-config")
	funnelCmd.MarkFlagRequired("funnel-config")
	funnelCmd.MarkFlagRequired("log")
}

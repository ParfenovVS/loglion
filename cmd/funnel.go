package cmd

import (
	"fmt"
	"os"

	"github.com/parfenovvs/loglion/internal/analyzer"
	"github.com/parfenovvs/loglion/internal/config"
	"github.com/parfenovvs/loglion/internal/output"
	"github.com/parfenovvs/loglion/internal/parser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var funnelCmd = &cobra.Command{
	Use:   "funnel",
	Short: "Analyze log files for funnel validation",
	Long: `Funnel command processes log files according to the funnel configuration
and outputs completion rates and drop-off analysis.

Examples:
  loglion funnel --parser-config parser.yaml --funnel-config funnel.yaml --log logcat.txt
  loglion funnel -p parser.yaml -f funnel.yaml -l logcat.txt --limit 5`,
	Run: func(cmd *cobra.Command, args []string) {
		parserConfigFile, _ := cmd.Flags().GetString("parser-config")
		funnelConfigFile, _ := cmd.Flags().GetString("funnel-config")
		logFile, _ := cmd.Flags().GetString("log")
		outputFormat, _ := cmd.Flags().GetString("output")
		limit, _ := cmd.Flags().GetInt("limit")

		logrus.WithFields(logrus.Fields{
			"parser_config_file": parserConfigFile,
			"funnel_config_file": funnelConfigFile,
			"log_file":           logFile,
			"output_format":      outputFormat,
			"limit":              limit,
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
		var entries []*parser.LogEntry
		var parseErr error

		if logFile != "" {
			_, err := os.Stat(logFile)
			if os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error parsing log file: open %s: no such file or directory\n", logFile)
				os.Exit(1)
			}
			entries, parseErr = logParser.ParseFile(logFile)
		} else {
			entries, parseErr = logParser.ParseReader(os.Stdin)
		}

		if parseErr != nil {
			logrus.WithError(parseErr).WithField("log_file", logFile).Error("Failed to parse log file")
			fmt.Fprintf(os.Stderr, "Error parsing log file: %v\n", parseErr)
			os.Exit(1)
		}

		logrus.Debug("Starting funnel analysis")
		result := funnelAnalyzer.AnalyzeFunnel(entries, limit)

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
		formattedOutput, err := formatter.FormatFunnel(result)
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
	funnelCmd.Flags().StringP("log", "l", "", "Path to log file (optional, stdin is used if not provided)")
	funnelCmd.Flags().StringP("output", "o", "text", "Output format (json, text)")
	funnelCmd.Flags().Int("limit", 0, "Maximum number of successful funnels to analyze (0 = analyze all funnels)")

	funnelCmd.MarkFlagRequired("parser-config")
	funnelCmd.MarkFlagRequired("funnel-config")
}

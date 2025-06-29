package cmd

import (
	"fmt"
	"os"

	"github.com/parfenovvs/loglion/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration files",
	Long: `Validate command checks if configuration files are properly formatted
and contain all required fields.

Examples:
  loglion validate --parser-config parser.yaml
  loglion validate --funnel-config funnel.yaml
  loglion validate --parser-config parser.yaml --funnel-config funnel.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		parserConfigFile, _ := cmd.Flags().GetString("parser-config")
		funnelConfigFile, _ := cmd.Flags().GetString("funnel-config")

		if parserConfigFile == "" && funnelConfigFile == "" {
			fmt.Fprintf(os.Stderr, "Error: At least one of --parser-config or --funnel-config must be specified.\n")
			os.Exit(1)
		}

		logrus.Info("Starting configuration validation")

		// Validate parser config if specified
		if parserConfigFile != "" {
			fmt.Printf("Validating parser config file: %s\n", parserConfigFile)
			logrus.Debug("Attempting to load and validate parser configuration")
			parserCfg, err := config.LoadParserConfig(parserConfigFile)
			if err != nil {
				logrus.WithError(err).WithField("parser_config_file", parserConfigFile).Error("Parser configuration validation failed")
				fmt.Fprintf(os.Stderr, "❌ Parser configuration validation failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Parser configuration is valid!\n")
			fmt.Printf("Event Regex: %s\n", parserCfg.EventRegex)
			fmt.Printf("JSON Extraction: %t\n", parserCfg.JSONExtraction)
		}

		// Validate funnel config if specified
		if funnelConfigFile != "" {
			fmt.Printf("Validating funnel config file: %s\n", funnelConfigFile)
			logrus.Debug("Attempting to load and validate funnel configuration")
			funnelCfg, err := config.LoadFunnelConfig(funnelConfigFile)
			if err != nil {
				logrus.WithError(err).WithField("funnel_config_file", funnelConfigFile).Error("Funnel configuration validation failed")
				fmt.Fprintf(os.Stderr, "❌ Funnel configuration validation failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Funnel configuration is valid!\n")
			fmt.Printf("Funnel: %s\n", funnelCfg.Name)
			fmt.Printf("Steps: %d\n", len(funnelCfg.Steps))
		}

		logrus.Info("Configuration validation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringP("parser-config", "p", "", "Path to parser configuration file")
	validateCmd.Flags().StringP("funnel-config", "f", "", "Path to funnel configuration file")
}

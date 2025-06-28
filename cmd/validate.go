package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"loglion/internal/config"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate funnel configuration file",
	Long: `Validate command checks if the funnel configuration file is properly formatted
and contains all required fields.

Example:
  loglion validate --config funnel.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")

		logrus.WithField("config_file", configFile).Info("Starting configuration validation")

		fmt.Printf("Validating config file: %s\n", configFile)

		// Load and validate configuration
		logrus.Debug("Attempting to load and validate configuration")
		cfg, err := config.LoadConfig(configFile)
		if err != nil {
			logrus.WithError(err).WithField("config_file", configFile).Error("Configuration validation failed")
			fmt.Fprintf(os.Stderr, "❌ Configuration validation failed: %v\n", err)
			os.Exit(1)
		}

		logrus.WithFields(logrus.Fields{
			"funnel_name": cfg.Funnel.Name,
			"step_count":  len(cfg.Funnel.Steps),
			"format":      cfg.Format,
		}).Info("Configuration validation completed successfully")

		fmt.Printf("✅ Configuration is valid!\n")
		fmt.Printf("Funnel: %s\n", cfg.Funnel.Name)
		fmt.Printf("Format: %s\n", cfg.Format)
		fmt.Printf("Steps: %d\n", len(cfg.Funnel.Steps))
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringP("config", "c", "", "Path to funnel configuration file (required)")
	validateCmd.MarkFlagRequired("config")
}

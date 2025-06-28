package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "loglion",
	Short: "LogLion - Analytics event funnel validator for log files",
	Long: `LogLion is a CLI tool that analyzes logcat files to validate 
analytics event funnels for automated testing.

It helps you track user conversion funnels by parsing log files
and checking if users complete expected sequences of analytics events.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
}

func setupLogging() {
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		})
	} else {
		logrus.SetLevel(logrus.PanicLevel)
	}
}

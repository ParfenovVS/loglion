package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "loglion",
	Short: "LogLion - Analytics event funnel validator for Android logs",
	Long: `LogLion is a CLI tool that analyzes ADB logcat logs to validate 
analytics event funnels for automated testing.

It helps you track user conversion funnels by parsing Android log files
and checking if users complete expected sequences of analytics events.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

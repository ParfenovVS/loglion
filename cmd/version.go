package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Version = "0.1.0-dev"
var BuildDate = "unknown"
var GitCommit = "unknown"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version, build date, and git commit information for LogLion.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.WithFields(logrus.Fields{
			"version":    Version,
			"build_date": BuildDate,
			"git_commit": GitCommit,
		}).Debug("Displaying version information")

		fmt.Printf("LogLion version %s\n", Version)
		fmt.Printf("Build date: %s\n", BuildDate)
		fmt.Printf("Git commit: %s\n", GitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

package cmd

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Version = "0.1.3"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version for LogLion.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.WithFields(logrus.Fields{
			"version":    Version,
		}).Debug("Displaying version information")

		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

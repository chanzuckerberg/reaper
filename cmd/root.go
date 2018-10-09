package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debug bool
	quiet bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "do not output to console; use return code to determine success/failure")
}

var rootCmd = &cobra.Command{
	Use:   "reaper",
	Short: "",
}

// Execute executes the command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

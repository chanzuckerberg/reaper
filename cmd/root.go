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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		logLevel := log.InfoLevel
		if debug { // debug overrides quiet
			logLevel = log.DebugLevel
		} else if quiet {
			logLevel = log.FatalLevel
		}
		log.SetLevel(logLevel)
		log.Infof("log level set to %s", log.GetLevel())
		return nil
	},
}

// Execute executes the command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

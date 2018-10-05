package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
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

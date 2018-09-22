package cmd

import (
	"fmt"
	"os"

	"github.com/chanzuckerberg/reaper/notifier"
	"github.com/chanzuckerberg/reaper/pkg/runner"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	flagConfig = "config"
	modeConfig = "mode"
)

func init() {
	runCmd.Flags().StringP(flagConfig, "c", "config.yml", "Use this to override the reaper config file.")
	runCmd.Flags().StringP(modeConfig, "m", "dry", "Run mode, must be one of [dry, interactive].")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run reaper",
	Long:  "Will run reaper and execute any policies defined in the config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return Run(cmd, args)
	},
}

// Run runs the policies and potentially takes action on them.
func Run(cmd *cobra.Command, args []string) error {

	// TODO maybe turn this to an enum with https://github.com/alvaroloes/enumer
	mode, err := cmd.Flags().GetString(modeConfig)
	if err != nil {
		return errors.Wrap(err, "Could not parse mode flag.")
	}
	if !(mode == "dry" || mode == "interactive") {
		return errors.Wrap(err, "mode must be one of [dry, interactive].")
	}

	conf, err := getConfig(cmd)
	if err != nil {
		return errors.Wrap(err, "could not read config")
	}

	notifier := notifier.New(os.Getenv("SLACK_TOKEN"))

	runner := runner.New(conf)
	violations, err := runner.Run()

	log.Info("VIOLATIONS")
	for _, v := range violations {
		fmt.Printf("resource %s is in violation of policy %s\n", v.Subject.GetID(), v.Policy.Name)
		err = notifier.Send(v)
		if err != nil {
			return err
		}
	}
	return nil
}

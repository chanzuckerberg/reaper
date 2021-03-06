package cmd

import (
	"fmt"
	"os"

	"github.com/chanzuckerberg/reaper/pkg/notifier"
	"github.com/chanzuckerberg/reaper/pkg/runner"
	"github.com/chanzuckerberg/reaper/pkg/ui"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	addCommonFlags(runCmd)
	runCmd.Flags().StringP(modeFlag, "m", "dry", "Run mode, must be one of [dry, interactive, non-interactive].")
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
	mode, err := cmd.Flags().GetString(modeFlag)
	if err != nil {
		return errors.Wrap(err, "Could not parse mode flag.")
	}
	if !(mode == "dry" || mode == "interactive" || mode == "non-interactive") {
		return errors.Wrap(err, "mode must be one of [dry, interactive, non-interactive].")
	}

	only, err := cmd.Flags().GetStringArray(onlyFlag)

	if err != nil {
		return errors.Wrap(err, "error when parsing `only` flag")
	}

	conf, err := getConfig(cmd)
	if err != nil {
		return errors.Wrap(err, "could not read config")
	}

	if conf == nil {
		return errors.New("nil config")
	}

	valid := validateConfigVersion(conf.Version, validConfigVersions)

	if !valid {
		return errors.Errorf("invalid config version: %d. Valid options are %v", conf.Version, validConfigVersions)
	}

	iMap, err := conf.GetIdentityMap()
	if err != nil {
		return err
	}

	ui := ui.NewInteractive()
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return errors.New("please supply a SLACK_TOKEN environment variable")
	}
	notifier := notifier.New(slackToken, ui, iMap)

	runner := runner.New(conf)
	violations, err := runner.Run(only)
	if err != nil {
		return err
	}

	log.Info("VIOLATIONS")
	for _, v := range violations {
		fmt.Printf("resource %s is in violation of policy %s\n", v.Subject.GetID(), v.Policy.Name)
		if mode == "dry" {
			continue
		}
		err = notifier.Send(v, mode == "non-interactive")
		// TODO report this to sentry
		log.Error(err)
	}
	return nil
}

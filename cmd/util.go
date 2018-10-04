package cmd

import (
	"github.com/chanzuckerberg/reaper/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	configFlag = "config"
	modeFlag   = "mode"
	onlyFlag   = "only"
)

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(configFlag, "c", "config.yml", "Use this to override the reaper config file.")
	cmd.Flags().StringArrayP(onlyFlag, "o", []string{}, "Run only listed policies.")

}

func getConfig(cmd *cobra.Command) (*config.Config, error) {
	configFile, err := cmd.Flags().GetString(configFlag)
	if err != nil {
		return nil, errors.Wrapf(err, "Missing required argument %s", configFlag)
	}
	return config.FromFile(configFile)
}

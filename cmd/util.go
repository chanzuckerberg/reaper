package cmd

import (
	"github.com/chanzuckerberg/reaper/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	configFlag = "config"
	modeFlag   = "mode"
	onlyFlag   = "only"
)

var validConfigVersions = []int{1}

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(configFlag, "c", "config.yml", "Use this to override the reaper config file.")
	cmd.Flags().StringArrayP(onlyFlag, "o", []string{}, "Run only listed policies.")

}

func getConfig(cmd *cobra.Command) (*config.Config, error) {
	configFile, err := cmd.Flags().GetString(configFlag)
	if err != nil {
		return nil, errors.Wrapf(err, "Missing required argument %s", configFlag)
	}
	fs := afero.NewOsFs()
	return config.FromFile(fs, configFile)
}

func validateConfigVersion(version int, validVersions []int) bool {
	for _, v := range validVersions {
		if v == version {
			return true
		}
	}
	return false
}

package cmd

import (
	"github.com/chanzuckerberg/aws-tidy/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func getConfig(cmd *cobra.Command) (*config.Config, error) {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, errors.Wrapf(err, "Missing required argument %s", flagConfig)
	}
	return config.FromFile(configFile)
}

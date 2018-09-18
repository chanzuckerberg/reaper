package cmd

import (
	"github.com/apex/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	flagConfig = "config"
	modeConfig = "mode"
)

func init() {
	runCmd.Flags().StringP(flagConfig, "c", "config.yml", "Use this to override the aws-tidy config file.")
	runCmd.Flags().StringP(modeConfig, "m", "dry", "Run mode, must be one of [dry, interactive].")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run aws-tidy",
	Long:  "Will run aws-tidy and execute any policies defined in the config",
	RunE: func(cmd *cobra.Command, args []string) error {
		return Run(cmd, args)
	},
}

// Run runs aws tidy
// TODO: mv this so this can be imported as a library as well
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
	policies, err := conf.GetPolicies()
	if err != nil {
		return err
	}

	awsClient, err := awsClient(conf.AWSRegions)
	if err != nil {
		return err
	}

	for _, p := range policies {
		log.Infof("Executing policy: \n%s \n=================", p.String())
		if p.MatchResource(map[string]string{"name": "s3"}) {
			err := awsClient.S3.Eval(&p, mode)
			if err != nil {
				return err
			}
		}

		// if p.MatchResource(map[string]string{"name": "ec2_instance"}) {
		// 	err := awsClient.EC2Instance.Walk(&p, mode)
		// 	if err != nil {
		// 		return err
		// 	}
		// }

		// if p.MatchResource(map[string]string{"name": "ec2_ebs_vol"}) {
		// 	err := awsClient.EC2EBSVol.Walk(&p, mode)
		// 	if err != nil {
		// 		return err
		// 	}
		// }

	}
	return nil
}

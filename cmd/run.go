package cmd

import (
	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws/session"
	cziAws "github.com/chanzuckerberg/aws-tidy/pkg/aws"
	"github.com/chanzuckerberg/aws-tidy/pkg/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const (
	flagConfig = "config"
)

func init() {
	runCmd.Flags().StringP(flagConfig, "c", "config.yml", "Use this to override the aws-tidy config file.")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run aws-tidy",
	Long:  "Will run aws-tidy and execute any policies defined in the config",
	Run: func(cmd *cobra.Command, args []string) {

		err := Run(cmd, args)
		if err != nil {
			panic(err)
		}
	},
}

// Run runs aws tidy
// TODO: mv this so this can be imported as a library as well
func Run(cmd *cobra.Command, args []string) error {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return errors.Wrapf(err, "Missing required argument %s", flagConfig)
	}
	conf, err := config.FromFile(configFile)
	if err != nil {
		return err
	}
	policies, err := conf.GetPolicies()
	if err != nil {
		return err
	}

	s := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	awsClient, err := cziAws.NewClient(s, conf.AWSRegions)
	if err != nil {
		return err
	}

	for _, p := range policies {
		log.Infof("Executing polify: \n%s \n=================", p.String())
		if p.MatchResource(map[string]string{"name": "s3"}) {
			err := awsClient.S3.Walk(&p)
			if err != nil {
				return err
			}
		}
		if p.MatchResource(map[string]string{"name": "ec2_instance"}) {
			err := awsClient.EC2Instance.Walk(&p)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

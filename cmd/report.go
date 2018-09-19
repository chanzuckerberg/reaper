package cmd

import (
	"os"

	"github.com/chanzuckerberg/aws-tidy/pkg/runner"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	reportCmd.Flags().StringP(flagConfig, "c", "config.yml", "Use this to override the aws-tidy config file.")
	rootCmd.AddCommand(reportCmd)
}

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		conf, err := getConfig(cmd)
		if err != nil {
			return errors.Wrap(err, "could not read config")
		}

		runner := runner.New(conf)
		violations, err := runner.Run()

		if err != nil {
			return err
		}

		log.Info("VIOLATIONS")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Entity", "Policy"})

		for _, v := range violations {
			// fmt.Printf("%#v\n", v)
			// fmt.Printf("%#v\n", v.Subject)
			table.Append([]string{v.Subject.GetID(), v.Policy.Name})
		}
		table.Render()
		return nil
	},
}

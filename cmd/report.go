package cmd

import (
	"os"
	"strconv"

	"github.com/chanzuckerberg/reaper/pkg/runner"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	reportCmd.Flags().StringP(flagConfig, "c", "config.yml", "Use this to override the reaper config file.")
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
		table.SetHeader([]string{"Entity", "Policy", "owner", "Account ID", "Account Name"})

		for _, v := range violations {
			table.Append([]string{v.Subject.GetID(), v.Policy.Name, v.Subject.GetOwner(), strconv.FormatInt(v.AccountID, 10), v.AccountName})
		}
		table.Render()
		return nil
	},
}

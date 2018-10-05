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
	addCommonFlags(reportCmd)
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

		only, err := cmd.Flags().GetStringArray(onlyFlag)
		if err != nil {
			return errors.Wrap(err, "error reading the `only` flag")
		}

		runner := runner.New(conf)
		violations, err := runner.Run(only)

		if err != nil {
			return err
		}

		log.Info("VIOLATIONS")
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Entity", "Policy", "owner", "Account ID", "Account Name", "Region"})

		for _, v := range violations {
			table.Append([]string{v.Subject.GetID(), v.Policy.Name, v.Subject.GetOwner(), strconv.FormatInt(v.AccountID, 10), v.AccountName, v.Subject.GetRegion()})
		}
		table.Render()
		return nil
	},
}

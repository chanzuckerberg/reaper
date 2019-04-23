package cmd

import (
	"fmt"

	"github.com/chanzuckerberg/go-misc/ver"
	"github.com/spf13/cobra"
)

var (
	Version = "undefined"
	GitSha  = "undefined"
	Release = "false"
	Dirty   = "true"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of reaper",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, e := ver.VersionString(Version, GitSha, Release, Dirty)
		if e != nil {
			return e
		}
		fmt.Println(v)
		return nil
	},
}

package cmd

import (
	"fmt"
	"os"

	"github.com/mgjules/minion/build"
	"github.com/urfave/cli/v2"
)

func init() {
	cli.VersionPrinter = func(c *cli.Context) {
		info, err := build.New()
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}

		fmt.Fprintf(
			c.App.Writer,
			"Revision: %v\nGo Version: %v\nLast Commit: %v\nDirty Build: %v\n",
			info.Revision, info.GoVersion, info.LastCommit, info.DirtyBuild,
		)
	}
}

var version = &cli.Command{
	Name:    "version",
	Aliases: []string{"v"},
	Usage:   "Shows the version",
	Action: func(c *cli.Context) error {
		cli.VersionPrinter(c)

		return nil
	},
}

package main

import (
	"fmt"
	"os"

	"github.com/mgjules/minion/cmd"
	"github.com/urfave/cli/v2"
)

// @title        Minion
// @version      v1.0.0
// @description  A little minion that can be replicated to create more minions.

// @contact.name   Michaël Giovanni Jules
// @contact.url    https://mgjules.dev
// @contact.email  julesmichaelgiovanni@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	app := cli.NewApp()
	app.Name = "Minion"
	app.Description = "A little minion that can be replicated to create more minions."
	app.Authors = []*cli.Author{
		{
			Name:  "Michaël Giovanni Jules",
			Email: "julesmichaelgiovanni@gmail.com",
		},
	}
	app.Copyright = "(c) 2022 Michaël Giovanni Jules"
	app.Commands = cmd.Commands

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("failed to execute cmd: %v\n", err)
		os.Exit(1)
	}
}

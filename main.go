package main

import (
	"fmt"
	"os"

	"github.com/mgjules/harvit/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "Harvit"
	app.Usage = "Harvest It!"
	app.Description = "Harvit harvests data from different sources (e.g websites, APIs) and transforms it."
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

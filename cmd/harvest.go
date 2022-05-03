package cmd

import (
	"fmt"

	"github.com/mgjules/harvit/harvit"
	"github.com/mgjules/harvit/logger"
	"github.com/mgjules/harvit/plan"
	"github.com/urfave/cli/v2"
)

var harvest = &cli.Command{
	Name:      "harvest",
	Usage:     "Let's harvest some data!",
	UsageText: "harvit harvest [command options] plan",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "prod",
			Value:   false,
			Usage:   "whether running in PROD or DEBUG mode",
			EnvVars: []string{"HARVIT_PROD"},
		},
	},
	Action: func(c *cli.Context) error {
		prod := c.Bool("prod")

		if _, err := logger.New(prod); err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
		}

		planFile := c.Args().Get(0)
		if planFile == "" {
			planFile = "plan.yml"
		}

		plan, err := plan.Load(planFile)
		if err != nil {
			return fmt.Errorf("failed to load plan: %w", err)
		}

		harvested, err := harvit.Harvest(plan)
		if err != nil {
			return fmt.Errorf("failed to harvest data: %w", err)
		}

		logger.Log.Debugw("harvesting done", "harvested", harvested)

		transformed, err := harvit.Transform(c.Context, plan, harvested)
		if err != nil {
			return fmt.Errorf("failed to transform data: %w", err)
		}

		logger.Log.Debugw("transformation done", "transformed", transformed)

		return nil
	},
}

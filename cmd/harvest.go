package cmd

import (
	"fmt"

	"github.com/mgjules/harvit/harvit"
	"github.com/mgjules/harvit/logger"
	"github.com/mgjules/harvit/plan"
	"github.com/urfave/cli/v2"
)

var harvest = &cli.Command{
	Name:  "harvest",
	Usage: "Let's harvest some data!",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "prod",
			Value:   false,
			Usage:   "whether running in PROD or DEBUG mode",
			EnvVars: []string{"HARVIT_PROD"},
		},
		&cli.StringFlag{
			Name:    "plan-file",
			Value:   "plan.yml",
			Usage:   "filepath of the plan file",
			EnvVars: []string{"HARVIT_PLAN_FILE"},
		},
	},
	Action: func(c *cli.Context) error {
		prod := c.Bool("prod")
		planFile := c.String("plan-file")

		if _, err := logger.New(prod); err != nil {
			return fmt.Errorf("failed to create logger: %w", err)
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

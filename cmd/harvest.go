package cmd

import (
	"fmt"

	"github.com/mgjules/harvit/conformer"
	"github.com/mgjules/harvit/harvester"
	"github.com/mgjules/harvit/json"
	"github.com/mgjules/harvit/logger"
	"github.com/mgjules/harvit/plan"
	"github.com/mgjules/harvit/transformer"
	"github.com/urfave/cli/v2"
)

var harvest = &cli.Command{
	Name:      "harvest",
	Usage:     "Let's harvest some data!",
	UsageText: "harvit harvest [command options] plan",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "debug",
			Value:   false,
			Usage:   "whether running in PROD or DEBUG mode",
			EnvVars: []string{"HARVIT_DEBUG"},
		},
	},
	Action: func(c *cli.Context) error {
		debug := c.Bool("debug")

		if _, err := logger.New(debug); err != nil {
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

		logger.Log.Debugw("loaded plan", "plan", plan)

		h, err := harvester.New(plan.Type)
		if err != nil {
			return fmt.Errorf("failed to create harvester: %w", err)
		}

		harvested, err := h.Harvest(c.Context, plan)
		if err != nil {
			return fmt.Errorf("failed to harvest data: %w", err)
		}

		logger.Log.Debugw("harvesting done", "harvested", harvested)

		conformed, err := conformer.Conform(c.Context, plan.Fields, harvested)
		if err != nil {
			return fmt.Errorf("failed to conform data: %w", err)
		}

		logger.Log.Debugw("conforming done", "conformed", conformed)

		var transformed any = conformed
		if plan.Transformer != "" {
			transformed, err = transformer.Transform(c.Context, plan.Transformer, plan.Fields, conformed)
			if err != nil {
				return fmt.Errorf("failed to transform data: %w", err)
			}

			logger.Log.Debugw("transformation done", "transformed", transformed)
		}

		marshaled, err := json.Marshal(transformed)
		if err != nil {
			return fmt.Errorf("failed to marshal transformed data: %w", err)
		}

		fmt.Println(string(marshaled))

		return nil
	},
}

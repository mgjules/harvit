package harvester

import (
	"context"

	"github.com/mgjules/harvit/plan"
)

// Harvester types.
const (
	TypeWebsite = "website"
)

// Harvester harvests data using a given plan.
type Harvester interface {
	Harvest(context.Context, *plan.Plan) (map[string]any, error)
}

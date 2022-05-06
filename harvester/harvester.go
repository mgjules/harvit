package harvester

import (
	"context"
	"fmt"

	"github.com/mgjules/harvit/plan"
)

// Harvester types.
const (
	TypeWebsite = "website"
)

// New returns a new Harvester.
func New(typ string) (Harvester, error) {
	switch typ {
	case TypeWebsite:
		return &Website{}, nil
	default:
		return nil, fmt.Errorf("unknown harvester type: %s", typ)
	}
}

// Harvester harvests data using a given plan.
type Harvester interface {
	Harvest(context.Context, *plan.Plan) (map[string]any, error)
}

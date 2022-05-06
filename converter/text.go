package converter

import (
	"context"

	"github.com/mgjules/harvit/plan"
)

// Text is a converter that converts a string to a string.
type Text struct{}

// Convert converts a string to a string.
func (Text) Convert(_ context.Context, s string, _ *plan.Field) any {
	return s
}

package converter

import (
	"context"
	"strconv"

	"github.com/mgjules/harvit/plan"
)

// Decimal is a converter that converts a string to a decimal.
type Decimal struct{}

// Convert converts a string to a decimal.
func (Decimal) Convert(_ context.Context, s string, _ *plan.Field) any {
	sanitized, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		sanitized = 0.0
	}

	return sanitized
}

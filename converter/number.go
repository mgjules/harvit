package converter

import (
	"context"
	"strconv"

	"github.com/mgjules/harvit/plan"
)

const (
	base    = 10
	bitSize = 64
)

// Number is a converter that converts a string to a number.
type Number struct{}

// Convert converts a string to a number.
func (Number) Convert(_ context.Context, s string, _ *plan.Field) any {
	sanitized, err := strconv.ParseInt(s, base, bitSize)
	if err != nil {
		sanitized = 0
	}

	return sanitized
}

package converter

import (
	"context"

	"github.com/golang-module/carbon/v2"
	"github.com/mgjules/harvit/plan"
)

// DateTime is a converter that converts a string to a date.
type DateTime struct{}

// Convert converts a string to a date.
func (DateTime) Convert(_ context.Context, s string, field *plan.Field) any {
	var parsed carbon.Carbon
	if field.Format == "" {
		parsed = carbon.Parse(s)
	} else {
		parsed = carbon.ParseByFormat(s, field.Format)
	}

	if field.Timezone != "" {
		parsed = parsed.SetTimezone(field.Timezone)
	}

	return parsed.ToIso8601String()
}

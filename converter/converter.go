package converter

import (
	"context"
	"fmt"

	"github.com/mgjules/harvit/plan"
)

// Field types.
const (
	TypeRaw      = "raw"
	TypeText     = "text"
	TypeNumber   = "number"
	TypeDecimal  = "decimal"
	TypeDateTime = "datetime"
)

// New returns a new Converter.
func New(typ string) (Converter, error) {
	switch typ {
	case TypeRaw, TypeText:
		return &Text{}, nil
	case TypeNumber:
		return &Number{}, nil
	case TypeDecimal:
		return &Decimal{}, nil
	case TypeDateTime:
		return &DateTime{}, nil
	default:
		return nil, fmt.Errorf("unknown converter type: %s", typ)
	}
}

// Converter converts a string to another format using plan.Field.
type Converter interface {
	Convert(context.Context, string, *plan.Field) any
}

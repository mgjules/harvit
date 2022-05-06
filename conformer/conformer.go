package conformer

import (
	"context"
	"regexp"
	"strings"

	"github.com/go-playground/mold/v4/modifiers"
	"github.com/iancoleman/strcase"
	"github.com/mgjules/harvit/converter"
	"github.com/mgjules/harvit/logger"
	"github.com/mgjules/harvit/plan"
	"github.com/samber/lo"
)

// Conform conforms any harvested data to a set of rules.
func Conform(ctx context.Context, fields []plan.Field, data map[string]any) (any, error) {
	conformed := make(map[string]any)

	for name, raw := range data {
		field, found := lo.Find(fields, func(f plan.Field) bool {
			return f.Name == name
		})

		if !found {
			continue
		}

		name = strcase.ToSnake(name)

		switch r := raw.(type) {
		case string:
			conformed[name] = conformField(ctx, &field, r)
		case []string:
			conformed[name] = make([]any, 0)
			for i := range r {
				conformed[name] = append( //nolint:forcetypeassert
					conformed[name].([]any),
					conformField(ctx, &field, r[i]),
				)
			}
		}
	}

	return conformed, nil
}

func conformField(ctx context.Context, field *plan.Field, val string) any {
	var err error

	if field.Regex != "" {
		var re *regexp.Regexp
		re, err = regexp.Compile(field.Regex)
		if err != nil {
			logger.Log.Warnw("failed to compile regex", "name", field.Name, "regex", field.Regex, "error", err)

			return val
		}

		matches := re.FindStringSubmatch(val)

		logger.Log.Debugw(
			"regex matches", "name", field.Name, "val", val, "regex", field.Regex, "matches", matches,
		)

		val = matches[1]
	}

	conform := modifiers.New()

	tags := []string{"trim"}
	switch field.Type {
	case converter.TypeNumber, converter.TypeDecimal:
		tags = append(tags, "strip_alpha", "strip_alpha_unicode", "strip_punctuation")
	}

	tags = lo.Uniq(tags)

	if err = conform.Field(ctx, &val, strings.Join(tags, ",")); err != nil {
		logger.Log.ErrorwContext(ctx, "failed to conform field", "error", err, "val", val, "tags", tags)

		return val
	}

	c, err := converter.New(field.Type)
	if err != nil {
		logger.Log.ErrorwContext(ctx, "failed to create converter", "error", err, "field", field)

		return val
	}

	return c.Convert(ctx, val, field)
}

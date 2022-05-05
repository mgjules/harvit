package harvit

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/mold/v4/modifiers"
	"github.com/golang-module/carbon/v2"
	"github.com/iancoleman/strcase"
	"github.com/mgjules/harvit/logger"
	"github.com/mgjules/harvit/plan"
	"github.com/samber/lo"
)

const (
	base    = 10
	bitSize = 64
)

// Transform transforms any harvested data.
func Transform(ctx context.Context, fields []plan.Field, data map[string]any) (any, error) {
	sanitizeds := make(map[string]any)

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
			sanitizeds[name] = Sanitize(ctx, &field, r)
		case []string:
			sanitizeds[name] = make([]any, 0)
			for i := range r {
				sanitizeds[name] = append( //nolint:forcetypeassert
					sanitizeds[name].([]any),
					Sanitize(ctx, &field, r[i]),
				)
			}
		}
	}

	logger.Log.Debugw("sanitized data", "sanitized", sanitizeds)

	return sanitizeds, nil
}

// Sanitize sanitizes a value according to a field.
func Sanitize(ctx context.Context, field *plan.Field, val string) any {
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
	case plan.FieldTypeNumber, plan.FieldTypeDecimal:
		tags = append(tags, "strip_alpha", "strip_alpha_unicode", "strip_punctuation")
	}

	tags = lo.Uniq(tags)

	if err = conform.Field(ctx, &val, strings.Join(tags, ",")); err != nil {
		logger.Log.ErrorwContext(ctx, "failed to conform field", "error", err, "val", val, "tags", tags)

		return val
	}

	var sanitized any
	switch field.Type {
	case plan.FieldTypeText:
		sanitized = val
	case plan.FieldTypeNumber:
		sanitized, err = strconv.ParseInt(val, base, bitSize)
		if err != nil {
			logger.Log.WarnwContext(ctx, "failed to parse number", "error", err, "val", val)
			sanitized = 0
		}
	case plan.FieldTypeDecimal:
		sanitized, err = strconv.ParseFloat(val, bitSize)
		if err != nil {
			logger.Log.WarnwContext(ctx, "failed to parse decimal", "error", err, "val", val)
			sanitized = 0.0
		}
	case plan.FieldTypeDateTime:
		var parsed carbon.Carbon
		if field.Format == "" {
			parsed = carbon.Parse(val)
		} else {
			parsed = carbon.ParseByFormat(val, field.Format)
		}

		if field.Timezone != "" {
			parsed = parsed.SetTimezone(field.Timezone)
		}

		sanitized = parsed.ToIso8601String()
	}

	return sanitized
}

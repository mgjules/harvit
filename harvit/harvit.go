package harvit

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	dynamicstruct "github.com/Ompluscator/dynamic-struct"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/go-playground/mold/v4/modifiers"
	"github.com/golang-module/carbon"
	"github.com/iancoleman/strcase"
	"github.com/mgjules/harvit/json"
	"github.com/mgjules/harvit/logger"
	"github.com/mgjules/harvit/plan"
	"github.com/samber/lo"
)

const (
	base    = 10
	bitSize = 64
)

// Harvest extracts data from a source using a plan.
func Harvest(ctx context.Context, p *plan.Plan) (map[string]any, error) {
	if _, err := url.Parse(p.Source); err != nil {
		return nil, fmt.Errorf("failed to parse source URL: %w", err)
	}

	// create context
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	harvested := make(map[string]any)

	actions := []chromedp.Action{
		chromedp.Navigate(p.Source),
	}

	for i := range p.Fields {
		field := p.Fields[i]

		actions = append(
			actions,
			chromedp.QueryAfter(field.Selector,
				func(ctx context.Context, eci runtime.ExecutionContextID, nodes ...*cdp.Node) error {
					logger.Log.Debugw("querying", "name", field.Name, "selector", field.Selector, "nodes", nodes)

					if len(nodes) > 1 {
						harvested[field.Name] = make([]string, 0)
						for i := range nodes {
							if nodes[i].ChildNodeCount == 0 || nodes[i].Children[0].NodeType != cdp.NodeTypeText {
								continue
							}

							harvested[field.Name] = append( //nolint:forcetypeassert
								harvested[field.Name].([]string),
								nodes[i].Children[0].NodeValue,
							)
						}
					} else if len(nodes) == 1 &&
						nodes[0].ChildNodeCount > 0 ||
						nodes[0].Children[0].NodeType != cdp.NodeTypeText {
						harvested[field.Name] = nodes[0].Children[0].NodeValue
					}

					return nil
				},
			),
		)
	}

	if err := chromedp.Run(ctx, actions...); err != nil {
		return nil, fmt.Errorf("failed to navigate to source: %w", err)
	}

	return harvested, nil
}

// Transform transforms any harvested data.
func Transform(ctx context.Context, fields []plan.Field, data map[string]any) (any, error) {
	var err error

	sanitizeds := make(map[string]any)

	for name, raw := range data {
		field, found := lo.Find(fields, func(d plan.Field) bool {
			return d.Name == name
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

	d, err := json.Marshal(sanitizeds)
	if err != nil {
		return nil, fmt.Errorf("failed to encode sanitized data from map: %w", err)
	}

	builder, err := dynamicStructBuilder(fields)
	if err != nil {
		return nil, fmt.Errorf("failed to build dynamic struct: %w", err)
	}

	transformed := builder.Build().New()

	if err := json.Unmarshal(d, &transformed); err != nil {
		return nil, fmt.Errorf("failed to decode sanitized data into dynamic struct: %w", err)
	}

	return transformed, nil
}

// Sanitize sanitizes a value according to a datum.
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
	case plan.FieldTypeText, plan.FieldTypeTextList:
		sanitized = val
	case plan.FieldTypeNumber, plan.FieldTypeNumberList:
		sanitized, err = strconv.ParseInt(val, base, bitSize)
		if err != nil {
			logger.Log.WarnwContext(ctx, "failed to parse number", "error", err, "val", val)
			sanitized = 0
		}
	case plan.FieldTypeDecimal, plan.FieldTypeDecimalList:
		sanitized, err = strconv.ParseFloat(val, bitSize)
		if err != nil {
			logger.Log.WarnwContext(ctx, "failed to parse decimal", "error", err, "val", val)
			sanitized = 0.0
		}
	case plan.FieldTypeDateTime, plan.FieldTypeDateTimeList:
		if field.Format == "" {
			sanitized = carbon.Parse(val).ToIso8601String()
		} else {
			sanitized = carbon.ParseByFormat(val, field.Format).ToIso8601String()
		}
	}

	return sanitized
}

func dynamicStructBuilder(fields []plan.Field) (dynamicstruct.Builder, error) {
	builder := dynamicstruct.NewStruct()

	for i := range fields {
		field := fields[i]

		var (
			typ  interface{}
			tags = map[string][]string{
				"json": {strcase.ToSnake(field.Name)},
			}
		)
		switch field.Type {
		case plan.FieldTypeText:
			typ = ""
		case plan.FieldTypeNumber:
			typ = 0
		case plan.FieldTypeDecimal:
			typ = 0.0
		case plan.FieldTypeDateTime:
			typ = time.Time{}
		case plan.FieldTypeTextList:
			typ = []string{}
		case plan.FieldTypeNumberList:
			typ = []int{}
		case plan.FieldTypeDecimalList:
			typ = []float64{}
		case plan.FieldTypeDateTimeList:
			typ = []time.Time{}
		}

		name := strcase.ToCamel(field.Name)

		var tt []string
		for k, v := range tags {
			tt = append(tt, fmt.Sprintf(`%s:"%s"`, k, strings.Join(v, ",")))
		}

		logger.Log.Debugw("adding field", "name", name, "type", typ, "tags", strings.Join(tt, " "))

		builder.AddField(name, typ, strings.Join(tt, " "))
	}

	return builder, nil
}

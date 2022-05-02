package harvit

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	dynamicstruct "github.com/Ompluscator/dynamic-struct"
	"github.com/go-playground/mold/v4/modifiers"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
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
func Harvest(p *plan.Plan) (map[string]string, error) {
	c := colly.NewCollector(
		colly.AllowedDomains(p.Domain),
	)

	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	c.OnRequest(func(r *colly.Request) {
		logger.Log.Debugw("visiting...", "url", r.URL.String())
	})

	harvested := make(map[string]string)

	for i := range p.Data {
		d := p.Data[i]
		c.OnHTML(d.Selector, func(e *colly.HTMLElement) {
			switch d.Type {
			case plan.DatumTypeText, plan.DatumTypeNumber, plan.DatumTypeDecimal:
				harvested[d.Name] = e.Text
			}
		})
	}

	url := p.Scheme + "://" + p.Domain + p.Path
	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("failed to visit site: %w", err)
	}

	return harvested, nil
}

// Transform transforms any harvested data.
func Transform(ctx context.Context, p *plan.Plan, data map[string]string) (any, error) {
	var err error

	conform := modifiers.New()

	sanitized := make(map[string]any)

	for name, raw := range data {
		d, found := lo.Find(p.Data, func(d plan.Datum) bool {
			return d.Name == name
		})

		if !found {
			continue
		}

		tags := []string{"trim"}
		switch d.Type {
		case plan.DatumTypeNumber, plan.DatumTypeDecimal:
			tags = append(tags, "strip_alpha", "strip_alpha_unicode", "strip_punctuation")
		}

		tags = lo.Uniq(tags)

		r := raw
		if err = conform.Field(ctx, &r, strings.Join(tags, ",")); err != nil {
			logger.Log.ErrorwContext(ctx, "failed to conform field", "error", err, "raw", r, "tags", tags)

			continue
		}

		name = strcase.ToSnake(name)

		var parsed any
		switch d.Type {
		case plan.DatumTypeText:
			parsed = raw
		case plan.DatumTypeNumber:
			parsed, err = strconv.ParseInt(raw, base, bitSize)
			if err != nil {
				logger.Log.WarnwContext(ctx, "failed to parse number", "error", err, "raw", raw)
				parsed = 0
			}
		case plan.DatumTypeDecimal:
			parsed, err = strconv.ParseFloat(raw, bitSize)
			if err != nil {
				logger.Log.WarnwContext(ctx, "failed to parse decimal", "error", err, "darawta", raw)
				parsed = 0.0
			}
		}

		sanitized[name] = parsed
	}

	logger.Log.Debugw("sanitized data", "sanitized", sanitized)

	d, err := json.Marshal(sanitized)
	if err != nil {
		return nil, fmt.Errorf("failed to encode sanitized data from map: %w", err)
	}

	builder, err := dynamicStructBuilder(p)
	if err != nil {
		return nil, fmt.Errorf("failed to build dynamic struct: %w", err)
	}

	transformed := builder.Build().New()

	if err := json.Unmarshal(d, &transformed); err != nil {
		return nil, fmt.Errorf("failed to decode sanitized data into dynamic struct: %w", err)
	}

	return transformed, nil
}

func dynamicStructBuilder(p *plan.Plan) (dynamicstruct.Builder, error) {
	builder := dynamicstruct.NewStruct()

	for i := range p.Data {
		d := p.Data[i]

		var (
			typ  interface{}
			tags = map[string][]string{
				"json": {strcase.ToSnake(d.Name)},
			}
		)
		switch d.Type {
		case plan.DatumTypeText:
			typ = ""
		case plan.DatumTypeNumber:
			typ = 0
		case plan.DatumTypeDecimal:
			typ = 0.0
		}

		name := strcase.ToCamel(d.Name)

		var tt []string
		for k, v := range tags {
			tt = append(tt, fmt.Sprintf(`%s:"%s"`, k, strings.Join(v, ",")))
		}

		logger.Log.Debugw("adding field", "name", name, "type", typ, "tags", strings.Join(tt, " "))

		builder.AddField(name, typ, strings.Join(tt, " "))
	}

	return builder, nil
}

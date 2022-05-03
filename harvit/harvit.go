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
	"github.com/go-playground/mold/v4/modifiers"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
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
func Harvest(p *plan.Plan) (map[string]string, error) {
	parsed, err := url.Parse(p.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URL: %w", err)
	}

	c := colly.NewCollector(
		colly.AllowedDomains(parsed.Host),
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
			case plan.DatumTypeText, plan.DatumTypeNumber, plan.DatumTypeDecimal, plan.DatumTypeDateTime:
				text := e.Text
				if d.Regex != "" {
					re, err := regexp.Compile(d.Regex)
					if err != nil {
						logger.Log.Warnw("failed to compile regex", "name", d.Name, "regex", d.Regex, "error", err)

						return
					}

					matches := re.FindStringSubmatch(text)

					logger.Log.Debugw(
						"regex matches", "name", d.Name, "text", text, "regex", d.Regex, "matches", matches,
					)

					text = matches[1]
				}

				harvested[d.Name] = text
			}
		})
	}

	if err := c.Visit(p.Source); err != nil {
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

		var parsed any
		switch d.Type {
		case plan.DatumTypeText:
			parsed = raw
		case plan.DatumTypeNumber:
			parsed, err = strconv.ParseInt(r, base, bitSize)
			if err != nil {
				logger.Log.WarnwContext(ctx, "failed to parse number", "error", err, "raw", r)
				parsed = 0
			}
		case plan.DatumTypeDecimal:
			parsed, err = strconv.ParseFloat(r, bitSize)
			if err != nil {
				logger.Log.WarnwContext(ctx, "failed to parse decimal", "error", err, "raw", r)
				parsed = 0.0
			}
		case plan.DatumTypeDateTime:
			if d.Format == "" {
				parsed = carbon.Parse(r).ToIso8601String()
			} else {
				parsed = carbon.ParseByFormat(r, d.Format).ToIso8601String()
			}
		}

		name = strcase.ToSnake(name)

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
		case plan.DatumTypeDateTime:
			typ = time.Time{}
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

package harvit

import (
	"context"
	"fmt"
	"net"
	"net/http"
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
func Harvest(ctx context.Context, p *plan.Plan) (map[string]any, error) {
	parsed, err := url.Parse(p.Source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source URL: %w", err)
	}

	c := colly.NewCollector(
		colly.AllowedDomains(parsed.Host),
	)

	c.WithTransport(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	c.OnRequest(func(r *colly.Request) {
		logger.Log.Debugw("visiting...", "url", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		logger.Log.ErrorwContext(ctx, "failed to visit", "url", r.Request.URL.String(), "err", err)
	})

	harvested := make(map[string]any)

	for i := range p.Fields {
		d := p.Fields[i]

		switch d.Type {
		case plan.FieldTypeTextList,
			plan.FieldTypeNumberList,
			plan.FieldTypeDecimalList,
			plan.FieldTypeDateTimeList:
			harvested[d.Name] = make([]string, 0)
		}

		c.OnHTML(d.Selector, func(e *colly.HTMLElement) {
			switch d.Type {
			case plan.FieldTypeText,
				plan.FieldTypeNumber,
				plan.FieldTypeDecimal,
				plan.FieldTypeDateTime:
				harvested[d.Name] = e.Text
			case plan.FieldTypeTextList,
				plan.FieldTypeNumberList,
				plan.FieldTypeDecimalList,
				plan.FieldTypeDateTimeList:
				harvested[d.Name] = append(harvested[d.Name].([]string), e.Text) //nolint:forcetypeassert
			}
		})
	}

	if err := c.Visit(p.Source); err != nil {
		return nil, fmt.Errorf("failed to visit site: %w", err)
	}

	return harvested, nil
}

// Transform transforms any harvested data.
func Transform(ctx context.Context, p *plan.Plan, data map[string]any) (any, error) {
	var err error

	sanitizeds := make(map[string]any)

	for name, raw := range data {
		d, found := lo.Find(p.Fields, func(d plan.Field) bool {
			return d.Name == name
		})

		if !found {
			continue
		}

		name = strcase.ToSnake(name)

		switch r := raw.(type) {
		case string:
			sanitizeds[name] = Sanitize(ctx, &d, r)
		case []string:
			sanitizeds[name] = make([]any, 0)
			for i := range r {
				sanitizeds[name] = append(sanitizeds[name].([]any), Sanitize(ctx, &d, r[i])) //nolint:forcetypeassert
			}
		}
	}

	logger.Log.Debugw("sanitized data", "sanitized", sanitizeds)

	d, err := json.Marshal(sanitizeds)
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

// Sanitize sanitizes a value according to a datum.
func Sanitize(ctx context.Context, d *plan.Field, val string) any {
	var err error

	if d.Regex != "" {
		var re *regexp.Regexp
		re, err = regexp.Compile(d.Regex)
		if err != nil {
			logger.Log.Warnw("failed to compile regex", "name", d.Name, "regex", d.Regex, "error", err)

			return val
		}

		matches := re.FindStringSubmatch(val)

		logger.Log.Debugw(
			"regex matches", "name", d.Name, "val", val, "regex", d.Regex, "matches", matches,
		)

		val = matches[1]
	}

	conform := modifiers.New()

	tags := []string{"trim"}
	switch d.Type {
	case plan.FieldTypeNumber, plan.FieldTypeDecimal:
		tags = append(tags, "strip_alpha", "strip_alpha_unicode", "strip_punctuation")
	}

	tags = lo.Uniq(tags)

	if err = conform.Field(ctx, &val, strings.Join(tags, ",")); err != nil {
		logger.Log.ErrorwContext(ctx, "failed to conform field", "error", err, "val", val, "tags", tags)

		return val
	}

	var sanitized any
	switch d.Type {
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
		if d.Format == "" {
			sanitized = carbon.Parse(val).ToIso8601String()
		} else {
			sanitized = carbon.ParseByFormat(val, d.Format).ToIso8601String()
		}
	}

	return sanitized
}

func dynamicStructBuilder(p *plan.Plan) (dynamicstruct.Builder, error) {
	builder := dynamicstruct.NewStruct()

	for i := range p.Fields {
		d := p.Fields[i]

		var (
			typ  interface{}
			tags = map[string][]string{
				"json": {strcase.ToSnake(d.Name)},
			}
		)
		switch d.Type {
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

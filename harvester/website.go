package harvester

import (
	"context"
	"fmt"
	"net/url"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/mgjules/harvit/converter"
	"github.com/mgjules/harvit/logger"
	"github.com/mgjules/harvit/plan"
)

// Website is a harvester that harvests data from a website.
type Website struct{}

// Harvest harvests data from a website using a plan.
//nolint:gocognit
func (Website) Harvest(ctx context.Context, p *plan.Plan) (map[string]any, error) {
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
							if field.Type == converter.TypeRaw {
								html, err := dom.GetOuterHTML().WithNodeID(nodes[i].NodeID).Do(ctx)
								if err != nil {
									logger.Log.ErrorwContext(ctx,
										"failed to get outer HTML",
										"name", field.Name, "selector", field.Selector, "node", nodes[i],
									)

									continue
								}

								harvested[field.Name] = html

								continue
							}

							if nodes[i].ChildNodeCount == 0 || nodes[i].Children[0].NodeType != cdp.NodeTypeText {
								continue
							}

							harvested[field.Name] = append( //nolint:forcetypeassert
								harvested[field.Name].([]string),
								nodes[i].Children[0].NodeValue,
							)
						}
					} else if len(nodes) == 1 {
						if field.Type == converter.TypeRaw {
							html, err := dom.GetOuterHTML().WithNodeID(nodes[0].NodeID).Do(ctx)
							if err != nil {
								logger.Log.ErrorwContext(ctx,
									"failed to get outer HTML",
									"name", field.Name, "selector", field.Selector, "node", nodes[0],
								)
							} else {
								harvested[field.Name] = html
							}
						} else if nodes[0].ChildNodeCount > 0 &&
							nodes[0].Children[0].NodeType == cdp.NodeTypeText {
							harvested[field.Name] = nodes[0].Children[0].NodeValue
						}
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

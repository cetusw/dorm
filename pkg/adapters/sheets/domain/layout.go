package domain

import "dorm/pkg/adapters/sheets/types"

type LayoutEngine struct {
	Margin int64
}

func (e *LayoutEngine) CalculateAnchors(widgets []types.Widget) map[types.Widget]types.Anchor {
	anchors := make(map[types.Widget]types.Anchor)
	currentColumn := int64(0)

	for _, widget := range widgets {
		anchors[widget] = types.Anchor{Row: 0, Col: currentColumn}
		currentColumn += widget.GetWidth() + e.Margin
	}
	return anchors
}

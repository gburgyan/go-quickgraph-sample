package handlers

import (
	"context"
	"errors"
	"github.com/gburgyan/go-quickgraph"
	"sync"
)

type Widget struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type WidgetCreateInput struct {
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

var (
	widgets = []Widget{
		{
			ID:       1,
			Name:     "Widget 1",
			Price:    1.00,
			Quantity: 10,
		},
	}
	widgetsMux sync.RWMutex
)

func RegisterWidgetHandlers(ctx context.Context, graphy *quickgraph.Graphy) {
	graphy.RegisterQuery(ctx, "GetWidget", GetWidget, "id")
	graphy.RegisterQuery(ctx, "GetWidgets", GetWidgets)
	graphy.RegisterMutation(ctx, "CreateWidget", CreateWidget, "widget")
	graphy.RegisterMutation(ctx, "UpdateWidget", UpdateWidget, "widget")
}

func GetWidget(id int) (Widget, error) {
	widgetsMux.RLock()
	defer widgetsMux.RUnlock()

	for _, widget := range widgets {
		if widget.ID == id {
			return widget, nil
		}
	}
	return Widget{}, errors.New("widget not found")
}

func GetWidgets() ([]Widget, error) {
	widgetsMux.RLock()
	defer widgetsMux.RUnlock()

	result := make([]Widget, len(widgets))
	copy(result, widgets)
	return result, nil
}

func CreateWidget(input WidgetCreateInput) (Widget, error) {
	widgetsMux.Lock()
	defer widgetsMux.Unlock()

	widget := Widget{
		ID:       len(widgets) + 1,
		Name:     input.Name,
		Price:    input.Price,
		Quantity: input.Quantity,
	}
	widgets = append(widgets, widget)

	// Broadcast the widget creation
	BroadcastWidgetUpdate(widget, "created")

	return widget, nil
}

func UpdateWidget(widget Widget) (Widget, error) {
	if widget.Quantity < 0 {
		return Widget{}, errors.New("quantity cannot be negative")
	}

	widgetsMux.Lock()
	defer widgetsMux.Unlock()

	for i, w := range widgets {
		if w.ID == widget.ID {
			widgets[i] = widget

			// Broadcast the widget update
			BroadcastWidgetUpdate(widget, "updated")

			return widget, nil
		}
	}
	return Widget{}, errors.New("widget not found")
}

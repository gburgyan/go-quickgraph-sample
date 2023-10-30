package handlers

import (
	"context"
	"errors"
	"github.com/gburgyan/go-quickgraph"
)

type Widget struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

var widgets = []Widget{
	{
		ID:       1,
		Name:     "Widget 1",
		Price:    1.00,
		Quantity: 10,
	},
}

func RegisterWidgetHandlers(ctx context.Context, graphy *quickgraph.Graphy) {
	graphy.RegisterQuery(ctx, "GetWidget", GetWidget, "id")
	graphy.RegisterQuery(ctx, "GetWidgets", GetWidgets)
	graphy.RegisterMutation(ctx, "CreateWidget", CreateWidget, "widget")
	graphy.RegisterMutation(ctx, "UpdateWidget", UpdateWidget, "widget")
}

func GetWidget(id int) (Widget, error) {
	for _, widget := range widgets {
		if widget.ID == id {
			return widget, nil
		}
	}
	return Widget{}, errors.New("widget not found")
}

func GetWidgets() ([]Widget, error) {
	return widgets, nil
}

func CreateWidget(widget Widget) (Widget, error) {
	widget.ID = len(widgets) + 1
	widgets = append(widgets, widget)
	return widget, nil
}

func UpdateWidget(widget Widget) (Widget, error) {
	if widget.Quantity < 0 {
		return Widget{}, errors.New("quantity cannot be negative")
	}
	for i, w := range widgets {
		if w.ID == widget.ID {
			widgets[i] = widget
			return widget, nil
		}
	}
	return Widget{}, errors.New("widget not found")
}

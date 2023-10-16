package handlers

import (
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

func WidgetDefinitions() []quickgraph.FunctionDefinition {
	return []quickgraph.FunctionDefinition{
		{
			Name:           "GetWidget",
			Function:       GetWidget,
			Mode:           quickgraph.ModeQuery,
			ParameterNames: []string{"id"},
		},
		{
			Name:     "GetWidgets",
			Function: GetWidgets,
			Mode:     quickgraph.ModeQuery,
		},
		{
			Name:           "CreateWidget",
			Function:       CreateWidget,
			Mode:           quickgraph.ModeMutation,
			ParameterNames: []string{"widget"},
		},
		{
			Name:           "UpdateWidget",
			Function:       UpdateWidget,
			Mode:           quickgraph.ModeMutation,
			ParameterNames: []string{"widget"},
		},
	}
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

func CreateWidget(widget *Widget) (Widget, error) {
	if widget == nil {
		return Widget{}, errors.New("widget is nil")
	}
	widget.ID = len(widgets) + 1
	widgets = append(widgets, *widget)
	return *widget, nil
}

func UpdateWidget(widget *Widget) (Widget, error) {
	if widget == nil {
		return Widget{}, errors.New("widget is nil")
	}
	for i, w := range widgets {
		if w.ID == widget.ID {
			widgets[i] = *widget
			return *widget, nil
		}
	}
	return Widget{}, errors.New("widget not found")
}

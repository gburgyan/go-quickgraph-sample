package handlers

import (
	"context"
	"errors"
	"github.com/gburgyan/go-quickgraph"
	"sync"
)

type Widget struct {
	ID       int     `graphy:"id,description=Unique widget identifier"`
	Name     string  `graphy:"widgetName,description=Widget name"`
	Price    float64 `graphy:"price,description=Widget price in USD"`
	Quantity int     `graphy:"stock,description=Available quantity in stock"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for Widget
func (Widget) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "A simple widget item with inventory tracking",
	}
}

type WidgetCreateInput struct {
	Name     string  `graphy:"name,description=Widget name"`
	Price    float64 `graphy:"price,description=Widget price in USD"`
	Quantity int     `graphy:"initialStock,description=Initial quantity in stock (must be non-negative)"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for WidgetCreateInput
func (WidgetCreateInput) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Input type for creating a new widget",
	}
}

// Validate implements the Validator interface for WidgetCreateInput
func (w WidgetCreateInput) Validate() error {
	if w.Name == "" {
		return errors.New("widget name is required")
	}
	if len(w.Name) < 2 {
		return errors.New("widget name must be at least 2 characters long")
	}
	if w.Price <= 0 {
		return errors.New("widget price must be greater than 0")
	}
	if w.Price > 10000 {
		return errors.New("widget price cannot exceed $10,000")
	}
	if w.Quantity < 0 {
		return errors.New("widget quantity cannot be negative")
	}
	if w.Quantity > 1000 {
		return errors.New("widget quantity cannot exceed 1000 units")
	}
	return nil
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
	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "GetWidget",
		Function:       GetWidget,
		ParameterNames: []string{"id"},
		Description:    strPtr("Retrieves a single widget by its ID"),
		Mode:           quickgraph.ModeQuery,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:        "GetWidgets",
		Function:    GetWidgets,
		Description: strPtr("Retrieves all widgets in inventory"),
		Mode:        quickgraph.ModeQuery,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "CreateWidget",
		Function:       CreateWidget,
		ParameterNames: []string{"widget"},
		Description:    strPtr("Creates a new widget with the specified properties"),
		Mode:           quickgraph.ModeMutation,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "UpdateWidget",
		Function:       UpdateWidget,
		ParameterNames: []string{"widget"},
		Description:    strPtr("Updates an existing widget's properties"),
		Mode:           quickgraph.ModeMutation,
	})
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

package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gburgyan/go-quickgraph"
)

// ProductUpdate represents a product update event
type ProductUpdate struct {
	Product   Product   `json:"product"`
	Action    string    `json:"action"` // "created", "updated", "deleted"
	Timestamp time.Time `json:"timestamp"`
}

// WidgetUpdate represents a widget update event
type WidgetUpdate struct {
	Widget    Widget    `json:"widget"`
	Action    string    `json:"action"` // "created", "updated", "deleted"
	Timestamp time.Time `json:"timestamp"`
}

// OrderUpdate represents an order update event
type OrderUpdate struct {
	OrderID   string    `json:"orderId"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Global channels for broadcasting updates
var (
	productUpdateChan = make(chan ProductUpdate, 100)
	widgetUpdateChan  = make(chan WidgetUpdate, 100)
	orderUpdateChan   = make(chan OrderUpdate, 100)
)

// BroadcastProductUpdate sends a product update to all subscribers
func BroadcastProductUpdate(product Product, action string) {
	select {
	case productUpdateChan <- ProductUpdate{
		Product:   product,
		Action:    action,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, drop the message (in production, use better buffering)
	}
}

// BroadcastWidgetUpdate sends a widget update to all subscribers
func BroadcastWidgetUpdate(widget Widget, action string) {
	select {
	case widgetUpdateChan <- WidgetUpdate{
		Widget:    widget,
		Action:    action,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, drop the message
	}
}

// BroadcastOrderUpdate sends an order update to subscribers
func BroadcastOrderUpdate(orderID, status, message string) {
	select {
	case orderUpdateChan <- OrderUpdate{
		OrderID:   orderID,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}:
	default:
		// Channel full, drop the message
	}
}

// ProductUpdates subscription - subscribes to all product changes
func ProductUpdates(ctx context.Context) <-chan ProductUpdate {
	ch := make(chan ProductUpdate)

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-productUpdateChan:
				select {
				case ch <- update:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch
}

// ProductUpdatesByCategory subscription - subscribes to product changes in a specific category
func ProductUpdatesByCategory(ctx context.Context, categoryId int) (<-chan ProductUpdate, error) {
	// Validate category exists
	productsMux.RLock()
	found := false
	for _, cat := range categories {
		if cat.ID == categoryId {
			found = true
			break
		}
	}
	productsMux.RUnlock()

	if !found {
		return nil, fmt.Errorf("category with ID %d not found", categoryId)
	}

	ch := make(chan ProductUpdate)

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-productUpdateChan:
				// Filter by category
				if update.Product.CategoryID == categoryId {
					select {
					case ch <- update:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// WidgetUpdates subscription - subscribes to widget updates
func WidgetUpdates(ctx context.Context, widgetId *int) <-chan WidgetUpdate {
	ch := make(chan WidgetUpdate)

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-widgetUpdateChan:
				// If widgetId is specified, filter by it
				if widgetId == nil || update.Widget.ID == *widgetId {
					select {
					case ch <- update:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch
}

// OrderStatusUpdates subscription - subscribes to order status changes
func OrderStatusUpdates(ctx context.Context, orderId string) (<-chan OrderUpdate, error) {
	if orderId == "" {
		return nil, fmt.Errorf("orderId is required")
	}

	ch := make(chan OrderUpdate)

	// Simulate initial order creation
	go func() {
		time.Sleep(100 * time.Millisecond)
		BroadcastOrderUpdate(orderId, "processing", "Order is being processed")
		time.Sleep(2 * time.Second)
		BroadcastOrderUpdate(orderId, "shipped", "Order has been shipped")
		time.Sleep(3 * time.Second)
		BroadcastOrderUpdate(orderId, "delivered", "Order has been delivered")
	}()

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-orderUpdateChan:
				// Filter by order ID
				if update.OrderID == orderId {
					select {
					case ch <- update:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// CurrentTime subscription - emits current time at specified intervals
func CurrentTime(ctx context.Context, intervalMs int) <-chan TimeUpdate {
	if intervalMs < 100 {
		intervalMs = 1000 // Default to 1 second
	}

	ch := make(chan TimeUpdate)

	go func() {
		defer close(ch)
		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				select {
				case ch <- TimeUpdate{
					Timestamp: t.Unix(),
					Formatted: t.Format(time.RFC3339),
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch
}

// TimeUpdate represents a time update message
type TimeUpdate struct {
	Timestamp int64  `json:"timestamp"`
	Formatted string `json:"formatted"`
}

// RegisterSubscriptionHandlers registers all subscription handlers
func RegisterSubscriptionHandlers(ctx context.Context, graph *quickgraph.Graphy) {
	// Product subscriptions
	graph.RegisterSubscription(ctx, "productUpdates", ProductUpdates)
	graph.RegisterSubscription(ctx, "productUpdatesByCategory", ProductUpdatesByCategory, "categoryId")

	// Widget subscriptions
	graph.RegisterSubscription(ctx, "widgetUpdates", WidgetUpdates, "widgetId")

	// Order subscriptions
	graph.RegisterSubscription(ctx, "orderStatusUpdates", OrderStatusUpdates, "orderId")

	// Utility subscriptions
	graph.RegisterSubscription(ctx, "currentTime", CurrentTime, "intervalMs")
}
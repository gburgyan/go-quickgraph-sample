package handlers

import (
	"context"
	"fmt"
	"sync"
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

// Global subscriber registries for broadcasting updates
var (
	productUpdateChan = make(chan ProductUpdate, 100)
	widgetUpdateChan  = make(chan WidgetUpdate, 100)
	orderUpdateChan   = make(chan OrderUpdate, 100)

	// Widget subscription management
	widgetSubscribers = make(map[string]chan WidgetUpdate)
	widgetSubsMutex   sync.RWMutex
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
	update := WidgetUpdate{
		Widget:    widget,
		Action:    action,
		Timestamp: time.Now(),
	}

	widgetSubsMutex.RLock()
	subscriberCount := len(widgetSubscribers)
	widgetSubsMutex.RUnlock()

	fmt.Printf("🔔 Broadcasting widget update: %s widget ID=%d to %d subscribers\n", action, widget.ID, subscriberCount)

	widgetSubsMutex.RLock()
	defer widgetSubsMutex.RUnlock()

	// Send to all registered subscribers
	sent := 0
	for subId, ch := range widgetSubscribers {
		select {
		case ch <- update:
			sent++
			fmt.Printf("  ✅ Sent to subscriber %s\n", subId)
		default:
			fmt.Printf("  ❌ Failed to send to subscriber %s (channel full)\n", subId)
		}
	}
	fmt.Printf("📊 Broadcast complete: %d/%d subscribers received update\n", sent, subscriberCount)
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

// ProductUpdates subscription - subscribes to product changes
// Use -1 for categoryId to get updates for all categories
func ProductUpdates(ctx context.Context, categoryId int) <-chan ProductUpdate {
	ch := make(chan ProductUpdate)

	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-productUpdateChan:
				// If categoryId is -1, send all updates; otherwise filter by specific category
				if categoryId == -1 || update.Product.CategoryID == categoryId {
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

// WidgetUpdates subscription - subscribes to widget updates
// Use -1 for widgetId to get updates for all widgets
func WidgetUpdates(ctx context.Context, widgetId int) <-chan WidgetUpdate {
	ch := make(chan WidgetUpdate, 10)
	subCh := make(chan WidgetUpdate, 10)

	// Generate unique subscription ID
	subId := fmt.Sprintf("widget-sub-%d", time.Now().UnixNano())

	fmt.Printf("🔗 NEW WIDGET SUBSCRIPTION CALLED: %s (filter: widgetId=%v)\n", subId, widgetId)
	fmt.Printf("🔗 Context: %+v\n", ctx)
	fmt.Printf("🔗 WidgetId pointer: %p, value: %v\n", widgetId, widgetId)

	// Register subscriber
	widgetSubsMutex.Lock()
	widgetSubscribers[subId] = subCh
	widgetSubsMutex.Unlock()

	go func() {
		defer close(ch)
		defer func() {
			// Unregister subscriber
			widgetSubsMutex.Lock()
			delete(widgetSubscribers, subId)
			close(subCh)
			widgetSubsMutex.Unlock()
			fmt.Printf("🔌 Widget subscription closed: %s\n", subId)
		}()

		// Send initial connection confirmation
		fmt.Printf("🚀 Widget subscription goroutine started: %s\n", subId)

		// Send initial confirmation update to keep subscription alive
		initialUpdate := WidgetUpdate{
			Widget: Widget{
				ID:       0,
				Name:     "Connection established",
				Price:    0,
				Quantity: 0,
			},
			Action:    "connected",
			Timestamp: time.Now(),
		}

		select {
		case ch <- initialUpdate:
			fmt.Printf("📤 Sent initial connection message to client: %s\n", subId)
		case <-ctx.Done():
			return
		}

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("⏹️ Widget subscription cancelled: %s\n", subId)
				return
			case update := <-subCh:
				fmt.Printf("📨 Widget subscription %s received update: %s widget ID=%d\n", subId, update.Action, update.Widget.ID)
				// If widgetId is -1, send all updates; otherwise filter by specific ID
				if widgetId == -1 || update.Widget.ID == widgetId {
					fmt.Printf("✅ Widget update passed filter, sending to client: %s\n", subId)
					select {
					case ch <- update:
					case <-ctx.Done():
						return
					}
				} else {
					fmt.Printf("🚫 Widget update filtered out (wanted ID=%d, got ID=%d): %s\n", widgetId, update.Widget.ID, subId)
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
	graph.RegisterSubscription(ctx, "productUpdates", ProductUpdates, "categoryId")

	// Widget subscriptions
	graph.RegisterSubscription(ctx, "widgetUpdates", WidgetUpdates, "widgetId")

	// Order subscriptions
	graph.RegisterSubscription(ctx, "orderStatusUpdates", OrderStatusUpdates, "orderId")

	// Utility subscriptions
	graph.RegisterSubscription(ctx, "currentTime", CurrentTime, "intervalMs")
}

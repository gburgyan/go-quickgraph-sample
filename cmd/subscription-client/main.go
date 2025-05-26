package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a message in the graphql-ws protocol
type WebSocketMessage struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// GraphQL-WS message types
const (
	GQLConnectionInit = "connection_init"
	GQLConnectionAck  = "connection_ack"
	GQLSubscribe      = "subscribe"
	GQLNext           = "next"
	GQLComplete       = "complete"
)

func runSubscriptionClient() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Connect to WebSocket
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/graphql"}
	log.Printf("Connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

	done := make(chan struct{})

	// Handle incoming messages
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var msg WebSocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Failed to parse message: %v", err)
				continue
			}

			switch msg.Type {
			case GQLConnectionAck:
				log.Println("Connection acknowledged")
			case GQLNext:
				log.Printf("Subscription data: %s", string(msg.Payload))
			case GQLComplete:
				log.Printf("Subscription %s completed", msg.ID)
			default:
				log.Printf("Received message type: %s", msg.Type)
			}
		}
	}()

	// Send connection init
	initMsg := WebSocketMessage{
		Type: GQLConnectionInit,
	}
	if err := conn.WriteJSON(initMsg); err != nil {
		log.Fatal("write init:", err)
	}

	// Wait for connection ack
	time.Sleep(100 * time.Millisecond)

	// Example 1: Subscribe to current time
	subscribeToTime(conn)

	// Example 2: Subscribe to product updates
	subscribeToProductUpdates(conn)

	// Example 3: Subscribe to order status
	subscribeToOrderStatus(conn, "order-123")

	// Wait for interrupt
	select {
	case <-interrupt:
		log.Println("Interrupt received, closing connection...")

		// Close connection
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("write close:", err)
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
}

func subscribeToTime(conn *websocket.Conn) {
	query := `subscription CurrentTime {
		currentTime(intervalMs: 1000) {
			timestamp
			formatted
		}
	}`

	subMsg := WebSocketMessage{
		ID:   "time-sub",
		Type: GQLSubscribe,
		Payload: json.RawMessage(fmt.Sprintf(`{
			"query": %q
		}`, query)),
	}

	if err := conn.WriteJSON(subMsg); err != nil {
		log.Printf("Failed to subscribe to time: %v", err)
	} else {
		log.Println("Subscribed to current time updates")
	}
}

func subscribeToProductUpdates(conn *websocket.Conn) {
	query := `subscription ProductUpdates {
		productUpdates {
			product {
				id
				name
				price
				status
			}
			action
			timestamp
		}
	}`

	subMsg := WebSocketMessage{
		ID:   "product-sub",
		Type: GQLSubscribe,
		Payload: json.RawMessage(fmt.Sprintf(`{
			"query": %q
		}`, query)),
	}

	if err := conn.WriteJSON(subMsg); err != nil {
		log.Printf("Failed to subscribe to products: %v", err)
	} else {
		log.Println("Subscribed to product updates")
	}
}

func subscribeToOrderStatus(conn *websocket.Conn, orderID string) {
	query := `subscription OrderStatus($orderId: String!) {
		orderStatusUpdates(orderId: $orderId) {
			orderId
			status
			message
			timestamp
		}
	}`

	variables := map[string]string{"orderId": orderID}
	varsJSON, _ := json.Marshal(variables)

	subMsg := WebSocketMessage{
		ID:   "order-sub",
		Type: GQLSubscribe,
		Payload: json.RawMessage(fmt.Sprintf(`{
			"query": %q,
			"variables": %s
		}`, query, varsJSON)),
	}

	if err := conn.WriteJSON(subMsg); err != nil {
		log.Printf("Failed to subscribe to order status: %v", err)
	} else {
		log.Printf("Subscribed to order %s status updates", orderID)
	}
}

// This is a separate program to test subscriptions
// Run with: go run subscription_client_example.go
func main() {
	fmt.Println("GraphQL Subscription Client Example")
	fmt.Println("===================================")
	fmt.Println("This client will:")
	fmt.Println("1. Connect to the GraphQL WebSocket endpoint")
	fmt.Println("2. Subscribe to current time (updates every second)")
	fmt.Println("3. Subscribe to product updates")
	fmt.Println("4. Subscribe to order status for order-123")
	fmt.Println("\nPress Ctrl+C to exit\n")

	runSubscriptionClient()
}

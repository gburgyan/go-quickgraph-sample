# GraphQL Subscriptions in go-quickgraph-sample

This sample now includes GraphQL subscriptions support using WebSockets. Subscriptions allow clients to receive real-time updates when data changes on the server.

## Available Subscriptions

### 1. Current Time
Emits the current time at specified intervals.
```graphql
subscription {
  currentTime(intervalMs: 1000) {
    timestamp
    formatted
  }
}
```

### 2. Product Updates
Receive notifications when products are created, updated, or deleted.
```graphql
# All product updates
subscription {
  productUpdates {
    product { id name price status }
    action
    timestamp
  }
}

# Product updates filtered by category
subscription {
  productUpdatesByCategory(categoryId: 1) {
    product { id name price status }
    action
    timestamp
  }
}
```

### 3. Widget Updates
Monitor widget changes, optionally filtered by widget ID.
```graphql
subscription {
  widgetUpdates(widgetId: 1) {
    widget { id name price quantity }
    action
    timestamp
  }
}
```

### 4. Order Status Updates
Track the status of a specific order as it progresses.
```graphql
subscription {
  orderStatusUpdates(orderId: "order-123") {
    orderId
    status
    message
    timestamp
  }
}
```

## Testing Subscriptions

### 1. Using the HTML Client
Open `subscription_client.html` in a web browser:
```bash
open subscription_client.html
```
This provides an interactive UI to test all subscription types.

### 2. Using the Go Client Example
Run the subscription client:
```bash
go run ./cmd/subscription-client
# Or: make run-client
```
This client automatically subscribes to multiple events and displays updates.

### 3. Triggering Events
To see subscription updates, you need to trigger mutations that cause changes:

Run the trigger example:
```bash
go run ./cmd/trigger-events
# Or: make run-trigger
```

Or manually trigger events using GraphQL mutations (see `SampleCommands.http`).

## Implementation Details

### WebSocket Protocol
The implementation uses the `graphql-ws` protocol over WebSockets:
1. Client connects to `ws://localhost:8080/graphql`
2. Client sends `connection_init` message
3. Server responds with `connection_ack`
4. Client can now send `subscribe` messages
5. Server sends `next` messages with data
6. Either party can send `complete` to end a subscription

### Broadcasting Updates
When mutations modify data, they broadcast updates to all active subscriptions:
- `BroadcastProductUpdate()` - for product changes
- `BroadcastWidgetUpdate()` - for widget changes
- `BroadcastOrderUpdate()` - for order status changes

### Architecture
- `websocket_adapter.go` - Adapts gorilla/websocket to quickgraph interface
- `handlers/subscription.go` - Contains all subscription handlers and broadcast logic
- Mutations in `handlers/product.go` and `handlers/widget.go` call broadcast functions

## Running the Full Example

1. Start the server:
```bash
go run ./cmd/server
# Or: make run-server
```

2. Open the HTML client in a browser or run the Go client:
```bash
# HTML client
open subscription_client.html

# Or Go client
go run ./cmd/subscription-client
# Or: make run-client
```

3. In another terminal, trigger some events:
```bash
go run ./cmd/trigger-events
# Or: make run-trigger
```

4. Watch the real-time updates appear in your subscription client!
# Go Quickgraph Comprehensive Sample

This sample application demonstrates a wide range of features available in the [go-quickgraph](https://github.com/gburgyan/go-quickgraph) library for building GraphQL APIs in Go, including real-time subscriptions over WebSocket.

## Features Demonstrated

### Core GraphQL Features
- **Code-First Approach**: GraphQL schema is automatically generated from Go types and functions
- **Queries and Mutations**: Full support for GraphQL operations
- **Subscriptions**: Real-time updates via WebSocket using graphql-ws protocol
- **Introspection**: Built-in schema introspection support

### Advanced Type System
- **Interfaces**: Employee interface implemented by Developer and Manager types
- **Union Types**: Search results that can return multiple types (Widget, Product, Employee)
- **Enums**: ProductStatus and UserRole enums with validation
- **Optional Fields**: Nullable fields using Go pointers
- **Complex Nested Types**: Products with categories, reviews, and user relationships

### Security & Performance
- **Query Limits**: DoS protection with configurable limits:
  - Max query depth
  - Max fields per level
  - Max aliases (prevents amplification attacks)
  - Max array size
  - Max concurrent resolvers
  - Query complexity scoring
- **Request Caching**: Parsed query caching for performance
- **Context-Based Authentication**: User authentication via context
- **Field-Level Authorization**: Restrict access to specific fields based on user role

### Development Features
- **HTTP Handler**: Ready-to-use HTTP handler with GET (schema) and POST (query) support
- **Timing**: Request timing for performance monitoring
- **Error Handling**: Proper GraphQL-compliant error responses
- **Health Check**: Simple health endpoint at `/health`

## Project Structure

This project uses a modular structure with multiple example programs:

```
cmd/
├── server/           # Main GraphQL server (port 8080)
├── gin-server/       # Gin framework example (port 8081)
├── subscription-client/  # WebSocket subscription client
└── trigger-events/   # Event generator for testing subscriptions

handlers/            # Business logic and GraphQL handlers
├── widget.go        # Basic CRUD operations
├── employee.go      # Interface types demo
├── product.go       # Complex relationships
├── search.go        # Union types
├── auth.go          # Authentication
└── subscription.go  # Real-time subscriptions
```

## Running the Examples

### Quick Start
```bash
# Show available commands
go run .

# Or use make
make help
```

### Main GraphQL Server
```bash
# Run the server (port 8080)
go run ./cmd/server
# Or: make run-server

# Endpoints:
# - GraphQL: http://localhost:8080/graphql
# - WebSocket: ws://localhost:8080/graphql
# - Health: http://localhost:8080/health
# - Schema: GET http://localhost:8080/graphql
```

### Testing Subscriptions
```bash
# Terminal 1: Start the server
make run-server

# Terminal 2: Run subscription client
make run-client

# Terminal 3: Trigger events
make run-trigger

# Or use the HTML client
open subscription_client.html
```

## Testing the API

Use the examples in `SampleCommands.http` with any HTTP client that supports GraphQL syntax. The file includes examples for:

- Basic queries and mutations
- Interface types with fragments
- Union type searches
- Enum usage
- Complex nested queries
- Authenticated requests
- Error cases

## Available Subscriptions

### Real-time Updates
- **Product Updates**: Monitor product creation, updates, and deletions
- **Widget Updates**: Track widget changes with optional filtering
- **Order Status**: Follow order progression through different states
- **Current Time**: Simple time ticker for testing

See [SUBSCRIPTIONS.md](SUBSCRIPTIONS.md) for detailed subscription documentation.

## Authentication

The sample includes a simple authentication middleware. To test authenticated endpoints:

```bash
# Admin user
Authorization: Bearer admin-token

# Regular user  
Authorization: Bearer user-token
```

## Generated Schema

View the complete generated GraphQL schema by visiting:
```
GET http://localhost:8080/graphql
```

## Building

```bash
# Build all examples
go build ./...
# Or: make build

# Binaries will be in ./bin/
```

## Additional Examples

- **Gin Framework Integration**: See `cmd/gin-server/` for using go-quickgraph with Gin (port 8081, WebSocket not implemented in this example)
- **WebSocket Subscriptions**: See `cmd/subscription-client/` for a subscription client example
- **Event Generation**: See `cmd/trigger-events/` for triggering subscription events
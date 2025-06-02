# Go Quickgraph Comprehensive Sample

This sample application demonstrates a wide range of features available in the [go-quickgraph](https://github.com/gburgyan/go-quickgraph) library for building GraphQL APIs in Go, including real-time subscriptions over WebSocket.

## Features Demonstrated

### Core GraphQL Features
- **Code-First Approach**: GraphQL schema is automatically generated from Go types and functions
- **Queries and Mutations**: Full support for GraphQL operations
- **Subscriptions**: Real-time updates via WebSocket using graphql-ws protocol
- **Introspection**: Built-in schema introspection support

### Advanced Type System
- **Custom Scalars**: DateTime, Money, HexColor, EmailAddress, ProductID, EmployeeID, URL with validation
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

### Command-Line Query Execution
You can also execute GraphQL queries directly from the command line without starting the server:

```bash
# Basic query
go run ./cmd/server -query 'query { GetAllEmployees { __typename ID Name } }'

# Query with variables
go run ./cmd/server -query 'query GetEmp($id: Int!) { GetEmployee(id: $id) { Name } }' -variables '{"id": 1}'

# Mutation example
go run ./cmd/server -query 'mutation { CreateWidget(widget: {name: "Test", price: 9.99, quantity: 10}) { id name } }'

# Complex query with fragments
go run ./cmd/server -query 'query { GetEmployee(id: 1) { __typename ... on Developer { Name ProgrammingLanguages } ... on Manager { Name Department } } }'

# Custom Scalar Examples
go run ./cmd/server -query 'query { validateEmail(email: "user@example.com") }'
go run ./cmd/server -query 'mutation { createColoredProduct(name: "Red Widget", price: 2500, color: "#FF0000") { id name price color } }'
go run ./cmd/server -query 'query { getEmployeeByIDScalar(id: "1") { ID Name Email } }'
```

### Custom Scalar Types
The sample demonstrates several custom scalar types with validation:

```bash
# Email validation
go run ./cmd/server -query 'query { validateEmail(email: "test@example.com") }'

# Colored product with Money and HexColor scalars
go run ./cmd/server -query 'mutation { createColoredProduct(name: "Blue Widget", price: 1500, color: "#0000FF") { id name price color } }'

# Invalid color (will fail validation)
go run ./cmd/server -query 'mutation { createColoredProduct(name: "Bad Widget", price: 1000, color: "invalid") { id } }'

# Different color formats
go run ./cmd/server -query 'mutation { createColoredProduct(name: "Green Widget", price: 2000, color: "#0F0") { id color } }'
```

This is particularly useful for:
- Testing queries quickly during development
- CI/CD pipelines that need to verify GraphQL endpoints
- Debugging specific queries without using an HTTP client
- Generating sample responses for documentation

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
- **Custom scalar types** (DateTime, Money, HexColor, EmailAddress, etc.)
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

## ⚠️ **IMPORTANT SECURITY NOTICE** ⚠️

**THIS IS A DEMONSTRATION APPLICATION - NOT PRODUCTION READY**

This sample application is designed to showcase the features of the go-quickgraph library and is **intentionally simplified for educational purposes**. It contains several security vulnerabilities that make it **unsuitable for production deployment**:

- **Hardcoded authentication tokens** for demo purposes
- **No query complexity limits** configured (allows DoS attacks)
- **Introspection enabled** (exposes internal schema)
- **Permissive CORS settings** (allows cross-origin access)
- **No rate limiting** or input validation
- **Development-mode error handling** (may leak sensitive information)

### 🔒 **For Production Use**

Before deploying any GraphQL service based on this sample:

1. **Read the [Security Documentation](../go-quickgraph/docs/SECURITY_API.md)** in the main library
2. **Implement proper authentication** (replace hardcoded tokens with JWT/OAuth)
3. **Configure query limits** and memory protection
4. **Disable introspection** in production environments
5. **Implement proper CORS policies** and security headers
6. **Add rate limiting** and input validation
7. **Enable production mode** for proper error handling

### Authentication (Demo Only)

The sample includes a **demonstration-only** authentication middleware with hardcoded tokens. **DO NOT use these in production**:

```bash
# Demo admin user (INSECURE - hardcoded token)
Authorization: Bearer admin-token

# Demo regular user (INSECURE - hardcoded token)
Authorization: Bearer user-token
```

**⚠️ These tokens are publicly known and provide no security.**

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
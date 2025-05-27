package main

import (
	"context"
	"fmt"
	"github.com/gburgyan/go-quickgraph"
	"github.com/gburgyan/go-quickgraph-sample/handlers"
	"github.com/patrickmn/go-cache"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx := context.Background()

	// Create graph with timing enabled
	graph := quickgraph.Graphy{
		EnableTiming: true,
	}

	// Configure query limits for DoS protection
	graph.QueryLimits = &quickgraph.QueryLimits{
		MaxDepth:               0, // Maximum query nesting depth
		MaxFields:              0, // Maximum fields per level
		MaxAliases:             0, // Prevent alias amplification attacks
		MaxArraySize:           0, // Limit array elements returned
		MaxConcurrentResolvers: 0, // Control parallel execution
		MaxComplexity:          0, // Overall query complexity score
	}

	// Register original handlers
	graph.RegisterQuery(ctx, "greeting", handlers.Greeting, "name")
	handlers.RegisterWidgetHandlers(ctx, &graph)

	// Register new feature handlers
	handlers.RegisterEmployeeHandlers(ctx, &graph)
	handlers.RegisterProductHandlers(ctx, &graph)
	handlers.RegisterSearchHandlers(ctx, &graph)
	handlers.RegisterAuthHandlers(ctx, &graph)
	handlers.RegisterSubscriptionHandlers(ctx, &graph)

	// Explicitly register types that aren't directly returned by any GraphQL function
	// This ensures they appear in the schema and can be used in unions
	graph.RegisterTypes(ctx, handlers.Employee{}, handlers.Developer{}, handlers.Manager{})

	// Enable introspection
	graph.EnableIntrospection(ctx)

	// Generate and save schema to file
	schema := graph.SchemaDefinition(ctx)
	err := os.WriteFile("schema.graphql", []byte(schema), 0644)
	if err != nil {
		log.Printf("Failed to write schema file: %v", err)
	} else {
		log.Println("Schema written to schema.graphql")
	}

	// Set a cache for parsed queries
	graph.RequestCache = &SimpleGraphRequestCache{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	// Create WebSocket upgrader
	upgrader := NewGorillaUpgrader()

	// Create HTTP handler with authentication middleware and WebSocket support
	graphHandler := handlers.AuthMiddleware(graph.HttpHandlerWithWebSocket(upgrader))

	http.Handle("/graphql", graphHandler)

	// Add a health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	log.Println("GraphQL server starting on http://localhost:8080/graphql")
	log.Println("WebSocket endpoint available at ws://localhost:8080/graphql")
	log.Println("Health check available at http://localhost:8080/health")
	log.Println("GraphQL schema available at GET http://localhost:8080/graphql")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

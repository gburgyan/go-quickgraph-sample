package main

import (
	"context"
	"encoding/json"
	"flag"
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
	// Parse command line flags
	queryFlag := flag.String("query", "", "Execute a GraphQL query directly and print the result")
	variablesFlag := flag.String("variables", "{}", "Variables for the query in JSON format")
	flag.Parse()

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

	// Register custom scalar types first
	if err := handlers.RegisterScalarHandlers(ctx, &graph); err != nil {
		log.Fatalf("Failed to register scalar handlers: %v", err)
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

	// Register scalar demo handlers
	handlers.RegisterScalarDemoHandlers(ctx, &graph)

	// Explicitly register types that aren't directly returned by any GraphQL function
	// This ensures they appear in the schema and can be used in unions
	graph.RegisterTypes(ctx, handlers.Employee{}, handlers.Developer{}, handlers.Manager{}, handlers.EmployeeResultUnion{})

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

	// If a query is provided via command line, execute it and exit
	if *queryFlag != "" {
		executeQueryAndExit(ctx, &graph, *queryFlag, *variablesFlag)
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

// executeQueryAndExit executes a GraphQL query and prints the result, then exits
func executeQueryAndExit(ctx context.Context, graph *quickgraph.Graphy, query string, variablesJSON string) {
	// Execute the query
	result, err := graph.ProcessRequest(ctx, query, variablesJSON)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// The result is already a JSON string, but let's parse and re-format it for pretty printing
	var jsonResult interface{}
	if err := json.Unmarshal([]byte(result), &jsonResult); err != nil {
		// If we can't parse it, just print it as-is
		fmt.Println(result)
	} else {
		// Pretty print the result
		output, err := json.MarshalIndent(jsonResult, "", "  ")
		if err != nil {
			fmt.Println(result)
		} else {
			fmt.Println(string(output))
		}
	}

	os.Exit(0)
}

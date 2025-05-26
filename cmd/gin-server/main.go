package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gburgyan/go-quickgraph"
	"github.com/gburgyan/go-quickgraph-sample/handlers"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

func main() {
	ctx := context.Background()

	// Create graph with timing enabled
	graph := quickgraph.Graphy{
		EnableTiming: true,
	}

	// Configure query limits for DoS protection
	graph.QueryLimits = &quickgraph.QueryLimits{
		MaxDepth:               10,   // Maximum query nesting depth
		MaxFields:              100,  // Maximum fields per level
		MaxAliases:             20,   // Prevent alias amplification attacks
		MaxArraySize:           100,  // Limit array elements returned
		MaxConcurrentResolvers: 50,   // Control parallel execution
		MaxComplexity:          1000, // Overall query complexity score
	}

	// Register handlers (same as main server)
	graph.RegisterQuery(ctx, "greeting", handlers.Greeting, "name")
	handlers.RegisterWidgetHandlers(ctx, &graph)
	handlers.RegisterEmployeeHandlers(ctx, &graph)
	handlers.RegisterProductHandlers(ctx, &graph)
	handlers.RegisterSearchHandlers(ctx, &graph)
	handlers.RegisterAuthHandlers(ctx, &graph)
	handlers.RegisterSubscriptionHandlers(ctx, &graph)

	// Enable introspection
	graph.EnableIntrospection(ctx)

	// Optional: Set a cache for parsed queries
	graph.RequestCache = &SimpleGraphRequestCache{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	// Set up Gin server
	server := gin.Default()

	type graphqlRequest struct {
		Query     string          `json:"query"`
		Variables json.RawMessage `json:"variables"`
	}

	// GraphQL endpoint
	server.POST("/graphql", func(c *gin.Context) {
		// Apply authentication middleware logic
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			user := handlers.GetUserFromAuthHeader(authHeader)
			if user != nil {
				c.Set("user", user)
			}
		}

		// Pull the query and variables from the request
		var request graphqlRequest
		err := c.BindJSON(&request)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Process the GraphQL request
		res, err := graph.ProcessRequest(c, request.Query, string(request.Variables))
		if err != nil {
			// Log the error here, but the response still has a GraphQL response that can be returned
			log.Printf("GraphQL processing error: %v", err)
		}

		// Return the response string
		c.Header("Content-Type", "application/json")
		c.String(200, res)
	})

	// Schema endpoint
	server.GET("/graphql", func(c *gin.Context) {
		schema := graph.SchemaDefinition(ctx)
		c.String(200, schema)
	})

	// Health check endpoint
	server.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"server": "gin",
		})
	})

	log.Println("Gin-based GraphQL server starting on http://localhost:8081/graphql")
	log.Println("Health check available at http://localhost:8081/health")
	log.Println("GraphQL schema available at GET http://localhost:8081/graphql")
	log.Println("Note: This example does not implement WebSocket subscriptions (though Gin can support them)")

	err := server.Run(":8081")
	if err != nil {
		log.Fatal("Failed to start Gin server:", err)
	}
}

// SimpleGraphRequestCache is copied from the main server
// In a real application, this would be in a shared package
type SimpleGraphRequestCache struct {
	cache *cache.Cache
}

func (d *SimpleGraphRequestCache) SetRequestStub(ctx context.Context, request string, stub *quickgraph.RequestStub, err error) {
	setErr := d.cache.Add(request, &simpleGraphRequestCacheEntry{
		request: request,
		stub:    stub,
		err:     err,
	}, cache.DefaultExpiration)
	if setErr != nil {
		// Cache might be full or key exists, try replacing
		d.cache.Set(request, &simpleGraphRequestCacheEntry{
			request: request,
			stub:    stub,
			err:     err,
		}, cache.DefaultExpiration)
	}
}

func (d *SimpleGraphRequestCache) GetRequestStub(ctx context.Context, request string) (*quickgraph.RequestStub, error) {
	value, found := d.cache.Get(request)
	if !found {
		return nil, nil
	}
	entry, ok := value.(*simpleGraphRequestCacheEntry)
	if !ok {
		return nil, nil
	}
	return entry.stub, entry.err
}

type simpleGraphRequestCacheEntry struct {
	request string
	stub    *quickgraph.RequestStub
	err     error
}

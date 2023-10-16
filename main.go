package main

import (
	"context"
	"encoding/json"
	"github.com/gburgyan/go-quickgraph"
	"github.com/gburgyan/go-quickgraph-sample/handlers"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"time"
)

func main() {
	server := gin.Default()
	graph := quickgraph.Graphy{}
	ctx := context.Background()

	graph.RegisterProcessorWithParamNames(ctx, "greeting", handlers.Greeting, "name")

	defs := handlers.WidgetDefinitions()
	for _, def := range defs {
		graph.RegisterFunction(ctx, def)
	}

	graph.RequestCache = &SimpleGraphRequestCache{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	type graphqlRequest struct {
		Query     string          `json:"query"`
		Variables json.RawMessage `json:"variables"`
	}

	server.POST("/graphql", func(c *gin.Context) {
		// Pull the query and variables from the request.
		var request graphqlRequest
		err := c.BindJSON(&request)
		if err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		res, err := graph.ProcessRequest(c, request.Query, string(request.Variables))
		if err != nil {
			// Log the error here, but the response still has a GraphQL response that can be returned.
		}
		// Return the response string.
		c.String(200, res)
	})
	server.GET("/graphql", func(c *gin.Context) {
		schema, err := graph.SchemaDefinition(ctx)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.String(200, schema)
	})

	server.Run()
}

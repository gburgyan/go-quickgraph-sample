package main

import (
	"context"
	"encoding/json"
	"github.com/gburgyan/go-quickgraph"
	"github.com/gin-gonic/gin"
)

func exampleGinServer(ctx context.Context, graph *quickgraph.Graphy) {
	server := gin.Default()

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
		c.Header("Content-Type", "application/json")
		c.String(200, res)
	})
	server.GET("/graphql", func(c *gin.Context) {
		schema := graph.SchemaDefinition(ctx)
		c.String(200, schema)
	})

	server.Run(":8081")
}

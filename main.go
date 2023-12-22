package main

import (
	"context"
	"github.com/gburgyan/go-quickgraph"
	"github.com/gburgyan/go-quickgraph-sample/handlers"
	"github.com/patrickmn/go-cache"
	"net/http"
	"time"
)

func main() {

	ctx := context.Background()

	graph := quickgraph.Graphy{EnableTiming: true}

	graph.RegisterQuery(ctx, "greeting", handlers.Greeting, "name")

	handlers.RegisterWidgetHandlers(ctx, &graph)

	graph.EnableIntrospection(ctx)

	// Optional: Set a cache for the graph.
	graph.RequestCache = &SimpleGraphRequestCache{
		cache: cache.New(5*time.Minute, 10*time.Minute),
	}

	http.Handle("/graphql", graph.HttpHandler())
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

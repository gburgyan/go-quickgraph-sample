package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("go-quickgraph-sample")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("This project contains multiple examples. Run them with:")
	fmt.Println()
	fmt.Println("  # Main GraphQL server (port 8080)")
	fmt.Println("  go run ./cmd/server")
	fmt.Println()
	fmt.Println("  # Gin-based GraphQL server (port 8081)")
	fmt.Println("  go run ./cmd/gin-server")
	fmt.Println()
	fmt.Println("  # WebSocket subscription client")
	fmt.Println("  go run ./cmd/subscription-client")
	fmt.Println()
	fmt.Println("  # Trigger events for subscriptions")
	fmt.Println("  go run ./cmd/trigger-events")
	fmt.Println()
	fmt.Println("Or build all examples:")
	fmt.Println("  go build ./...")
	fmt.Println()
	
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("For more information, see README.md and SUBSCRIPTIONS.md")
	} else {
		fmt.Println("Run with --help for more information")
	}
}
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// GraphQL request/response structures
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []interface{}   `json:"errors,omitempty"`
}

// triggerProductUpdate creates or updates a product to trigger subscription events
func triggerProductUpdate() {
	// Create a new product
	createQuery := `
		mutation CreateProduct($input: ProductInput!) {
			CreateProduct(input: $input) {
				id
				name
				price
				status
			}
		}
	`

	createVars := map[string]interface{}{
		"input": map[string]interface{}{
			"name":        fmt.Sprintf("Test Product %d", time.Now().Unix()),
			"description": "Product created to test subscriptions",
			"price":       99.99,
			"categoryId":  1,
		},
	}

	resp := executeGraphQL(createQuery, createVars)
	fmt.Printf("Created product: %s\n", resp.Data)

	// Parse the response to get the product ID
	var createResult struct {
		CreateProduct struct {
			ID int `json:"id"`
		} `json:"CreateProduct"`
	}
	json.Unmarshal(resp.Data, &createResult)

	if createResult.CreateProduct.ID > 0 {
		// Update the product status after a delay
		time.Sleep(2 * time.Second)

		updateQuery := `
			mutation UpdateStatus($id: Int!, $status: ProductStatus!) {
				UpdateProductStatus(id: $id, status: $status) {
					id
					name
					status
				}
			}
		`

		updateVars := map[string]interface{}{
			"id":     createResult.CreateProduct.ID,
			"status": "ACTIVE",
		}

		resp = executeGraphQL(updateQuery, updateVars)
		fmt.Printf("Updated product status: %s\n", resp.Data)
	}
}

// triggerWidgetUpdate creates and updates widgets
func triggerWidgetUpdate() {
	// Create a widget
	createQuery := `
		mutation CreateWidget($widget: Widget!) {
			CreateWidget(widget: $widget) {
				id
				name
				quantity
			}
		}
	`

	createVars := map[string]interface{}{
		"widget": map[string]interface{}{
			"name":     fmt.Sprintf("Test Widget %d", time.Now().Unix()),
			"price":    49.99,
			"quantity": 100,
		},
	}

	resp := executeGraphQL(createQuery, createVars)
	fmt.Printf("Created widget: %s\n", resp.Data)

	// Parse response to get widget ID
	var createResult struct {
		CreateWidget struct {
			ID int `json:"id"`
		} `json:"CreateWidget"`
	}
	json.Unmarshal(resp.Data, &createResult)

	if createResult.CreateWidget.ID > 0 {
		// Update the widget after a delay
		time.Sleep(1 * time.Second)

		updateQuery := `
			mutation UpdateWidget($widget: Widget!) {
				UpdateWidget(widget: $widget) {
					id
					name
					quantity
				}
			}
		`

		updateVars := map[string]interface{}{
			"widget": map[string]interface{}{
				"id":       createResult.CreateWidget.ID,
				"name":     fmt.Sprintf("Updated Widget %d", time.Now().Unix()),
				"price":    59.99,
				"quantity": 150,
			},
		}

		resp = executeGraphQL(updateQuery, updateVars)
		fmt.Printf("Updated widget: %s\n", resp.Data)
	}
}

// executeGraphQL sends a GraphQL request
func executeGraphQL(query string, variables map[string]interface{}) *GraphQLResponse {
	url := "http://localhost:8080/graphql"

	request := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Failed to marshal request: %v", err)
		return &GraphQLResponse{}
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		return &GraphQLResponse{}
	}

	req.Header.Set("Content-Type", "application/json")
	// Add auth token if needed
	// req.Header.Set("Authorization", "Bearer admin-token")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to execute request: %v", err)
		return &GraphQLResponse{}
	}
	defer resp.Body.Close()

	var graphqlResp GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&graphqlResp); err != nil {
		log.Printf("Failed to decode response: %v", err)
		return &GraphQLResponse{}
	}

	if len(graphqlResp.Errors) > 0 {
		log.Printf("GraphQL errors: %v", graphqlResp.Errors)
	}

	return &graphqlResp
}

// This is a separate program to trigger subscription events
// Run with: go run subscription_trigger_example.go
func main() {
	fmt.Println("GraphQL Subscription Trigger Example")
	fmt.Println("====================================")
	fmt.Println("This will trigger events that subscription clients can observe.")
	fmt.Println("Make sure the server is running and you have subscription clients connected.\n")

	// Trigger different types of updates
	for i := 0; i < 3; i++ {
		fmt.Printf("\n--- Trigger round %d ---\n", i+1)
		
		// Trigger product updates
		fmt.Println("Triggering product update...")
		triggerProductUpdate()
		
		time.Sleep(2 * time.Second)
		
		// Trigger widget updates
		fmt.Println("Triggering widget update...")
		triggerWidgetUpdate()
		
		time.Sleep(3 * time.Second)
	}

	fmt.Println("\nDone triggering events!")
}
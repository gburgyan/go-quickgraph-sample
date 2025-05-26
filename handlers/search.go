package handlers

import (
	"context"
	"github.com/gburgyan/go-quickgraph"
	"strings"
)


func RegisterSearchHandlers(ctx context.Context, graphy *quickgraph.Graphy) {
	// Register the search query
	graphy.RegisterQuery(ctx, "Search", Search, "query")
}

// SearchResultUnion explicitly defines the union type
type SearchResultUnion struct {
	Widget   *Widget
	Product  *Product
	Employee *Employee
}

// Search demonstrates union types by returning different types based on search
// Returns SearchResultUnion which explicitly defines possible types
func Search(query string) ([]SearchResultUnion, error) {
	query = strings.ToLower(query)
	var results []SearchResultUnion

	// Search widgets
	widgetsMux.RLock()
	for _, w := range widgets {
		if strings.Contains(strings.ToLower(w.Name), query) {
			// Create a copy to avoid data races
			widget := w
			results = append(results, SearchResultUnion{Widget: &widget})
		}
	}
	widgetsMux.RUnlock()

	// Search products  
	productsMux.RLock()
	for _, p := range products {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Description), query) {
			// Create a copy to avoid data races
			product := p
			results = append(results, SearchResultUnion{Product: &product})
		}
	}
	productsMux.RUnlock()

	// Search employees
	employeeMux.RLock()
	for _, emp := range employees {
		switch e := emp.(type) {
		case *Developer:
			if strings.Contains(strings.ToLower(e.Name), query) ||
				strings.Contains(strings.ToLower(e.Email), query) {
				// Return the base Employee type
				employee := e.Employee
				results = append(results, SearchResultUnion{Employee: &employee})
			}
		case *Manager:
			if strings.Contains(strings.ToLower(e.Name), query) ||
				strings.Contains(strings.ToLower(e.Email), query) ||
				strings.Contains(strings.ToLower(e.Department), query) {
				// Return the base Employee type
				employee := e.Employee
				results = append(results, SearchResultUnion{Employee: &employee})
			}
		}
	}
	employeeMux.RUnlock()

	return results, nil
}

// Alternative approach: Use explicit union type with methods returning different types
// This creates a more explicit union in the GraphQL schema
type SearchResult interface {
	IsSearchResult()
}

// Make our types implement the interface
func (Widget) IsSearchResult()    {}
func (Product) IsSearchResult()   {}
func (Developer) IsSearchResult() {}
func (Manager) IsSearchResult()   {}

// SearchV2 using the interface approach
func SearchV2(query string) ([]SearchResult, error) {
	query = strings.ToLower(query)
	var results []SearchResult

	// Search widgets
	for _, w := range widgets {
		if strings.Contains(strings.ToLower(w.Name), query) {
			widget := w
			results = append(results, widget)
		}
	}

	// Search products
	for _, p := range products {
		if strings.Contains(strings.ToLower(p.Name), query) ||
			strings.Contains(strings.ToLower(p.Description), query) {
			product := p
			results = append(results, product)
		}
	}

	// Note: This approach works but the first approach with []interface{} 
	// is more flexible and idiomatic for go-quickgraph

	return results, nil
}
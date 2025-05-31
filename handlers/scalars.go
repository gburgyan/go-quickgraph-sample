package handlers

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gburgyan/go-quickgraph"
)

// Custom scalar types used in the sample application

// EmployeeID represents a unique identifier for employees
type EmployeeID string

// ProductID represents a unique identifier for products
type ProductID string

// HexColor represents a color in hexadecimal format
type HexColor string

// Money represents a monetary amount with currency
type Money struct {
	Amount   int64  `json:"amount"`   // Amount in cents
	Currency string `json:"currency"` // Currency code (e.g., "USD", "EUR")
}

// EmailAddress represents a validated email address
type EmailAddress string

// RegisterScalarHandlers registers all custom scalar types with the GraphQL engine
func RegisterScalarHandlers(ctx context.Context, graph *quickgraph.Graphy) error {
	// Register DateTime scalar for time.Time
	if err := graph.RegisterDateTimeScalar(ctx); err != nil {
		return fmt.Errorf("failed to register DateTime scalar: %w", err)
	}

	// Register EmployeeID scalar
	if err := graph.RegisterScalar(ctx, quickgraph.ScalarDefinition{
		Name:        "EmployeeID",
		GoType:      reflect.TypeOf(EmployeeID("")),
		Description: "Unique identifier for employees",
		Serialize: func(value interface{}) (interface{}, error) {
			switch v := value.(type) {
			case EmployeeID:
				return string(v), nil
			case *EmployeeID:
				if v != nil {
					return string(*v), nil
				}
				return nil, nil
			default:
				return nil, fmt.Errorf("expected EmployeeID, got %T", value)
			}
		},
		ParseValue: func(value interface{}) (interface{}, error) {
			if str, ok := value.(string); ok {
				if str == "" {
					return nil, fmt.Errorf("EmployeeID cannot be empty")
				}
				return EmployeeID(str), nil
			}
			return nil, fmt.Errorf("expected string for EmployeeID, got %T", value)
		},
	}); err != nil {
		return fmt.Errorf("failed to register EmployeeID scalar: %w", err)
	}

	// Register ProductID scalar
	if err := graph.RegisterScalar(ctx, quickgraph.ScalarDefinition{
		Name:        "ProductID",
		GoType:      reflect.TypeOf(ProductID("")),
		Description: "Unique identifier for products",
		Serialize: func(value interface{}) (interface{}, error) {
			switch v := value.(type) {
			case ProductID:
				return string(v), nil
			case *ProductID:
				if v != nil {
					return string(*v), nil
				}
				return nil, nil
			default:
				return nil, fmt.Errorf("expected ProductID, got %T", value)
			}
		},
		ParseValue: func(value interface{}) (interface{}, error) {
			if str, ok := value.(string); ok {
				if str == "" {
					return nil, fmt.Errorf("ProductID cannot be empty")
				}
				return ProductID(str), nil
			}
			return nil, fmt.Errorf("expected string for ProductID, got %T", value)
		},
	}); err != nil {
		return fmt.Errorf("failed to register ProductID scalar: %w", err)
	}

	// Register HexColor scalar
	if err := graph.RegisterScalar(ctx, quickgraph.ScalarDefinition{
		Name:        "HexColor",
		GoType:      reflect.TypeOf(HexColor("")),
		Description: "Hexadecimal color representation (e.g., #FF0000)",
		Serialize: func(value interface{}) (interface{}, error) {
			switch v := value.(type) {
			case HexColor:
				return string(v), nil
			case *HexColor:
				if v != nil {
					return string(*v), nil
				}
				return nil, nil
			default:
				return nil, fmt.Errorf("expected HexColor, got %T", value)
			}
		},
		ParseValue: func(value interface{}) (interface{}, error) {
			if str, ok := value.(string); ok {
				if !isValidHexColor(str) {
					return nil, fmt.Errorf("invalid hex color format: %s", str)
				}
				return HexColor(str), nil
			}
			return nil, fmt.Errorf("expected string for HexColor, got %T", value)
		},
	}); err != nil {
		return fmt.Errorf("failed to register HexColor scalar: %w", err)
	}

	// Register Money scalar
	if err := graph.RegisterScalar(ctx, quickgraph.ScalarDefinition{
		Name:        "Money",
		GoType:      reflect.TypeOf(Money{}),
		Description: "Monetary amount with currency (e.g., '123.45 USD')",
		Serialize: func(value interface{}) (interface{}, error) {
			switch v := value.(type) {
			case Money:
				return v.String(), nil
			case *Money:
				if v != nil {
					return v.String(), nil
				}
				return nil, nil
			default:
				return nil, fmt.Errorf("expected Money, got %T", value)
			}
		},
		ParseValue: func(value interface{}) (interface{}, error) {
			if str, ok := value.(string); ok {
				return ParseMoney(str)
			}
			return nil, fmt.Errorf("expected string for Money (e.g., '123.45 USD'), got %T", value)
		},
	}); err != nil {
		return fmt.Errorf("failed to register Money scalar: %w", err)
	}

	// Register EmailAddress scalar
	if err := graph.RegisterScalar(ctx, quickgraph.ScalarDefinition{
		Name:        "EmailAddress",
		GoType:      reflect.TypeOf(EmailAddress("")),
		Description: "Valid email address",
		Serialize: func(value interface{}) (interface{}, error) {
			switch v := value.(type) {
			case EmailAddress:
				return string(v), nil
			case *EmailAddress:
				if v != nil {
					return string(*v), nil
				}
				return nil, nil
			default:
				return nil, fmt.Errorf("expected EmailAddress, got %T", value)
			}
		},
		ParseValue: func(value interface{}) (interface{}, error) {
			if str, ok := value.(string); ok {
				if !isValidEmail(str) {
					return nil, fmt.Errorf("invalid email address: %s", str)
				}
				return EmailAddress(str), nil
			}
			return nil, fmt.Errorf("expected string for EmailAddress, got %T", value)
		},
	}); err != nil {
		return fmt.Errorf("failed to register EmailAddress scalar: %w", err)
	}

	// Register URL scalar
	if err := graph.RegisterScalar(ctx, quickgraph.ScalarDefinition{
		Name:        "URL",
		GoType:      reflect.TypeOf(url.URL{}),
		Description: "Valid URL",
		Serialize: func(value interface{}) (interface{}, error) {
			switch v := value.(type) {
			case url.URL:
				return v.String(), nil
			case *url.URL:
				if v != nil {
					return v.String(), nil
				}
				return nil, nil
			default:
				return nil, fmt.Errorf("expected url.URL, got %T", value)
			}
		},
		ParseValue: func(value interface{}) (interface{}, error) {
			if str, ok := value.(string); ok {
				u, err := url.Parse(str)
				if err != nil {
					return nil, fmt.Errorf("invalid URL: %v", err)
				}
				return *u, nil
			}
			return nil, fmt.Errorf("expected string for URL, got %T", value)
		},
	}); err != nil {
		return fmt.Errorf("failed to register URL scalar: %w", err)
	}

	// Register JSON scalar for arbitrary JSON data
	if err := graph.RegisterJSONScalar(ctx); err != nil {
		return fmt.Errorf("failed to register JSON scalar: %w", err)
	}

	return nil
}

// Helper functions for validation

// isValidHexColor validates hex color format
func isValidHexColor(color string) bool {
	if len(color) != 7 && len(color) != 4 {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for i := 1; i < len(color); i++ {
		c := color[i]
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

// isValidEmail provides basic email validation
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	// Basic validation - must contain @ and have text before and after
	atIndex := strings.Index(email, "@")
	if atIndex <= 0 || atIndex >= len(email)-1 {
		return false
	}

	// Check for multiple @ symbols
	if strings.Count(email, "@") != 1 {
		return false
	}

	// Basic domain validation - must contain a dot after @
	domain := email[atIndex+1:]
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

// Utility functions for working with custom scalars

// String formats a Money value as "amount currency" string (e.g., "123.45 USD")
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", float64(m.Amount)/100, m.Currency)
}

// Dollars returns the Money amount as dollars (float64)
func (m Money) Dollars() float64 {
	return float64(m.Amount) / 100
}

// IsZero returns true if the Money amount is zero
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// Add adds two Money values (must have same currency)
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Subtract subtracts two Money values (must have same currency)
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.Currency, other.Currency)
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

// Multiply multiplies a Money value by a factor
func (m Money) Multiply(factor float64) Money {
	return Money{Amount: int64(float64(m.Amount) * factor), Currency: m.Currency}
}

// ParseMoney parses a money string in format "123.45 USD" into Money struct
func ParseMoney(s string) (Money, error) {
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return Money{}, fmt.Errorf("invalid money format: expected 'amount currency' (e.g., '123.45 USD'), got '%s'", s)
	}

	amountStr := parts[0]
	currency := parts[1]

	// Parse the amount
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return Money{}, fmt.Errorf("invalid amount in money string '%s': %v", s, err)
	}

	// Convert to cents
	cents := int64(amount * 100)

	// Validate currency (basic check for 3-letter codes)
	if len(currency) != 3 {
		return Money{}, fmt.Errorf("invalid currency code '%s': must be 3 letters", currency)
	}

	return Money{Amount: cents, Currency: strings.ToUpper(currency)}, nil
}

// NewMoney creates a new Money value from dollars and currency
func NewMoney(dollars float64, currency string) Money {
	return Money{
		Amount:   int64(dollars * 100),
		Currency: strings.ToUpper(currency),
	}
}

// Sample functions demonstrating custom scalar usage

// GetEmployeeByID demonstrates EmployeeID scalar usage
func GetEmployeeByIDScalar(id EmployeeID) (*Employee, error) {
	// Convert EmployeeID to int for lookup
	idInt, err := strconv.Atoi(string(id))
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID format: %v", err)
	}
	return GetEmployee(idInt)
}

// GetCurrentDateTime demonstrates DateTime scalar usage
func GetCurrentDateTime() time.Time {
	return time.Now()
}

// GetServerStartTime demonstrates DateTime scalar with fixed value
func GetServerStartTime() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

// ColoredProduct represents a product with color information
type ColoredProduct struct {
	ID    ProductID `json:"id"`
	Name  string    `json:"name"`
	Price Money     `json:"price"`
	Color HexColor  `json:"color"`
}

// CreateColoredProduct demonstrates HexColor and Money scalar usage
func CreateColoredProduct(name string, price Money, color HexColor) ColoredProduct {
	return ColoredProduct{
		ID:    ProductID(fmt.Sprintf("prod_%d", time.Now().Unix())),
		Name:  name,
		Price: price,
		Color: color,
	}
}

// ValidateEmail demonstrates EmailAddress scalar usage
func ValidateEmail(email EmailAddress) bool {
	return isValidEmail(string(email))
}

// ProductWithMetadata demonstrates JSON scalar usage
type ProductWithMetadata struct {
	ID       ProductID              `json:"id"`
	Name     string                 `json:"name"`
	Price    Money                  `json:"price"`
	Metadata map[string]interface{} `json:"metadata"`
}

// CreateProductWithMetadata demonstrates JSON scalar usage
func CreateProductWithMetadata(name string, price Money, metadata map[string]interface{}) ProductWithMetadata {
	return ProductWithMetadata{
		ID:       ProductID(fmt.Sprintf("prod_meta_%d", time.Now().Unix())),
		Name:     name,
		Price:    price,
		Metadata: metadata,
	}
}

// GetSampleJSONData returns sample JSON data for demonstration
func GetSampleJSONData() map[string]interface{} {
	return map[string]interface{}{
		"features": []string{"waterproof", "durable", "lightweight"},
		"dimensions": map[string]interface{}{
			"width":  10.5,
			"height": 20.3,
			"depth":  5.0,
			"unit":   "cm",
		},
		"specs": map[string]interface{}{
			"material":    "aluminum",
			"weight":      "250g",
			"color":       "#FF0000",
			"warranty":    "2 years",
			"energy_star": true,
		},
		"tags": []interface{}{"electronics", "gadget", "premium"},
	}
}

// ProcessJSONMetadata processes arbitrary JSON metadata and returns summary
func ProcessJSONMetadata(metadata map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Count different types of values
	typeCount := make(map[string]int)
	var keyCount int

	for _, value := range metadata {
		keyCount++
		switch value.(type) {
		case string:
			typeCount["string"]++
		case float64, int, int64:
			typeCount["number"]++
		case bool:
			typeCount["boolean"]++
		case []interface{}:
			typeCount["array"]++
		case map[string]interface{}:
			typeCount["object"]++
		default:
			typeCount["other"]++
		}
	}

	result["totalKeys"] = keyCount
	result["typeBreakdown"] = typeCount
	result["hasNestedObjects"] = typeCount["object"] > 0
	result["hasArrays"] = typeCount["array"] > 0

	return result
}

// RegisterScalarDemoHandlers registers additional demo functions that use custom scalars
func RegisterScalarDemoHandlers(ctx context.Context, graph *quickgraph.Graphy) {
	// Query functions demonstrating scalar usage
	graph.RegisterQuery(ctx, "getEmployeeByIDScalar", GetEmployeeByIDScalar, "id")
	graph.RegisterQuery(ctx, "getCurrentDateTime", GetCurrentDateTime)
	graph.RegisterQuery(ctx, "getServerStartTime", GetServerStartTime)
	graph.RegisterQuery(ctx, "validateEmail", ValidateEmail, "email")
	graph.RegisterQuery(ctx, "getSampleJSONData", GetSampleJSONData)
	graph.RegisterQuery(ctx, "processJSONMetadata", ProcessJSONMetadata, "metadata")

	// Mutations demonstrating multiple scalar types
	graph.RegisterMutation(ctx, "createColoredProduct", CreateColoredProduct, "name", "price", "color")
	graph.RegisterMutation(ctx, "createProductWithMetadata", CreateProductWithMetadata, "name", "price", "metadata")
}

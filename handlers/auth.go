package handlers

import (
	"context"
	"errors"
	"github.com/gburgyan/go-quickgraph"
	"net/http"
)

// Context key type for type safety
type contextKey string

const (
	// UserContextKey is used to store the current user in context
	UserContextKey contextKey = "currentUser"
)

func RegisterAuthHandlers(ctx context.Context, graphy *quickgraph.Graphy) {
	// Register query that uses context
	graphy.RegisterQuery(ctx, "GetCurrentUser", GetCurrentUser)
	
	// Register the PersonalDetails method on Employee interface
	// Note: This registration might not be necessary as methods are usually auto-discovered
	// graphy.RegisterFunction(ctx, (*Employee).PersonalDetails)
}

// GetCurrentUser demonstrates using context to get authenticated user
// Handlers can accept context as first parameter when they need it
func GetCurrentUser(ctx context.Context) (*User, error) {
	// Extract user from context
	userValue := ctx.Value(UserContextKey)
	if userValue == nil {
		return nil, errors.New("not authenticated")
	}

	// Type assert to User pointer
	user, ok := userValue.(*User)
	if !ok || user == nil {
		return nil, errors.New("invalid user in context")
	}

	return user, nil
}

// AuthMiddleware is an example HTTP middleware that could be used
// to inject user into context before GraphQL processing
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In a real app, you'd validate JWT token or session here
		// For demo, we'll simulate authentication based on a header
		authHeader := r.Header.Get("Authorization")
		
		if authHeader != "" {
			// Simulate user lookup
			var user *User
			switch authHeader {
			case "Bearer admin-token":
				// Use existing user from product.go users slice
				if len(users) > 0 {
					user = &users[0]
				}
			case "Bearer user-token":
				// Use existing user from product.go users slice
				if len(users) > 1 {
					user = &users[1]
				}
			}
			
			if user != nil {
				// Add user to context
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				r = r.WithContext(ctx)
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// GetUserFromAuthHeader extracts user based on authorization header
// This is useful for non-middleware based servers like Gin
func GetUserFromAuthHeader(authHeader string) *User {
	if authHeader == "" {
		return nil
	}
	
	switch authHeader {
	case "Bearer admin-token":
		if len(users) > 0 {
			return &users[0]
		}
	case "Bearer user-token":
		if len(users) > 1 {
			return &users[1]
		}
	}
	
	return nil
}

// Example of a mutation that uses context for authorization
func CreateProductWithAuth(ctx context.Context, input ProductInput) (*Product, error) {
	// Check if user is authenticated and has admin role
	userValue := ctx.Value(UserContextKey)
	if userValue == nil {
		return nil, errors.New("authentication required")
	}

	user, ok := userValue.(*User)
	if !ok {
		return nil, errors.New("invalid user in context")
	}

	if user.Role != UserRoleAdmin {
		return nil, errors.New("admin role required")
	}

	// Proceed with normal creation
	return CreateProduct(input)
}

// PersonalDetails method demonstrates field-level authorization
// Returns sensitive employee information only to authorized users
func (e *Employee) PersonalDetails(ctx context.Context) (*PersonalInfo, error) {
	// Extract current user from context
	currentUser := ctx.Value(UserContextKey)
	if currentUser == nil {
		return nil, errors.New("authentication required to view personal details")
	}

	user, ok := currentUser.(*User)
	if !ok {
		return nil, errors.New("invalid user in context")
	}

	// Allow if admin or viewing own data
	if user.Role == UserRoleAdmin || user.Email == e.Email {
		return &PersonalInfo{
			Salary:      e.Salary,
			Email:       e.Email,
			PhoneNumber: "+1-555-0123", // Mock data
			Address:     "123 Main St, Anytown, USA", // Mock data
		}, nil
	}

	return nil, errors.New("not authorized to view personal details")
}

// PersonalInfo contains sensitive employee information
type PersonalInfo struct {
	Salary      float64 `json:"salary"`
	Email       string  `json:"email"`
	PhoneNumber string  `json:"phoneNumber"`
	Address     string  `json:"address"`
}
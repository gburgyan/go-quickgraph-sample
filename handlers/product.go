package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gburgyan/go-quickgraph"
	"sync"
	"time"
)

// Enum types
type ProductStatus string

const (
	ProductStatusDraft        ProductStatus = "DRAFT"
	ProductStatusActive       ProductStatus = "ACTIVE"
	ProductStatusDiscontinued ProductStatus = "DISCONTINUED"
	ProductStatusOutOfStock   ProductStatus = "OUT_OF_STOCK"
)

// EnumValues implements the StringEnumValues interface for schema generation
func (ProductStatus) EnumValues() []string {
	return []string{"DRAFT", "ACTIVE", "DISCONTINUED", "OUT_OF_STOCK"}
}

type UserRole string

const (
	UserRoleAdmin    UserRole = "ADMIN"
	UserRoleCustomer UserRole = "CUSTOMER"
	UserRoleGuest    UserRole = "GUEST"
)

// EnumValues implements the StringEnumValues interface for schema generation
func (UserRole) EnumValues() []string {
	return []string{"ADMIN", "CUSTOMER", "GUEST"}
}

// Domain types
type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Status      ProductStatus
	CategoryID  int
	InStock     bool
}

type Category struct {
	ID          int
	Name        string
	Description *string // Optional field
}

type Review struct {
	ID        int
	ProductID int
	UserID    int
	Rating    int
	Comment   string
	CreatedAt string
}

type User struct {
	ID       int
	Username string
	Email    string
	Role     UserRole
}

// Input types
type ProductInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CategoryID  int     `json:"categoryId"`
}

type ProductFilter struct {
	CategoryID *int           `json:"categoryId"`
	MinPrice   *float64       `json:"minPrice"`
	MaxPrice   *float64       `json:"maxPrice"`
	Status     *ProductStatus `json:"status"`
	InStock    *bool          `json:"inStock"`
}

type ReviewInput struct {
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
}

// Storage
var (
	products    []Product
	categories  []Category
	reviews     []Review
	users       []User
	productsMux sync.RWMutex
	nextProdID  = 1
	nextRevID   = 1
)

func init() {
	// Initialize sample data
	categories = []Category{
		{ID: 1, Name: "Electronics", Description: strPtr("Electronic devices and accessories")},
		{ID: 2, Name: "Books", Description: strPtr("Physical and digital books")},
		{ID: 3, Name: "Clothing", Description: nil},
	}

	products = []Product{
		{ID: 1, Name: "Laptop", Description: "High-performance laptop", Price: 999.99, Status: ProductStatusActive, CategoryID: 1, InStock: true},
		{ID: 2, Name: "Go Programming Book", Description: "Learn Go in 30 days", Price: 39.99, Status: ProductStatusActive, CategoryID: 2, InStock: true},
		{ID: 3, Name: "Vintage T-Shirt", Description: "Retro design", Price: 24.99, Status: ProductStatusOutOfStock, CategoryID: 3, InStock: false},
		{ID: 4, Name: "Smartphone", Description: "Latest model", Price: 699.99, Status: ProductStatusActive, CategoryID: 1, InStock: true},
	}

	users = []User{
		{ID: 1, Username: "admin", Email: "admin@example.com", Role: UserRoleAdmin},
		{ID: 2, Username: "john_customer", Email: "john@example.com", Role: UserRoleCustomer},
		{ID: 3, Username: "jane_customer", Email: "jane@example.com", Role: UserRoleCustomer},
	}

	reviews = []Review{
		{ID: 1, ProductID: 1, UserID: 2, Rating: 5, Comment: "Excellent laptop!", CreatedAt: "2024-01-15T10:00:00Z"},
		{ID: 2, ProductID: 1, UserID: 3, Rating: 4, Comment: "Good value for money", CreatedAt: "2024-01-16T14:30:00Z"},
		{ID: 3, ProductID: 2, UserID: 2, Rating: 5, Comment: "Great book for beginners", CreatedAt: "2024-01-17T09:15:00Z"},
	}

	nextProdID = 5
	nextRevID = 4
}

func RegisterProductHandlers(ctx context.Context, graphy *quickgraph.Graphy) {
	// Query registrations
	graphy.RegisterQuery(ctx, "GetProduct", GetProduct, "id")
	graphy.RegisterQuery(ctx, "GetProducts", GetProducts, "filter")
	graphy.RegisterQuery(ctx, "GetCategories", GetCategories)
	
	// Mutation registrations
	graphy.RegisterMutation(ctx, "CreateProduct", CreateProduct, "input")
	graphy.RegisterMutation(ctx, "UpdateProductStatus", UpdateProductStatus, "id", "status")
	graphy.RegisterMutation(ctx, "AddProductReview", AddProductReview, "productId", "review")
	
	// Note: Methods on Product, Category, Review, and User types will be automatically
	// exposed as fields when those objects are returned from queries
}

// Query handlers
func GetProduct(id int) (*Product, error) {
	productsMux.RLock()
	defer productsMux.RUnlock()

	for _, p := range products {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("product with id %d not found", id)
}

func GetProducts(filter *ProductFilter) ([]Product, error) {
	productsMux.RLock()
	defer productsMux.RUnlock()

	var result []Product
	for _, p := range products {
		if filter != nil {
			// Apply filters
			if filter.CategoryID != nil && p.CategoryID != *filter.CategoryID {
				continue
			}
			if filter.MinPrice != nil && p.Price < *filter.MinPrice {
				continue
			}
			if filter.MaxPrice != nil && p.Price > *filter.MaxPrice {
				continue
			}
			if filter.Status != nil && p.Status != *filter.Status {
				continue
			}
			if filter.InStock != nil && p.InStock != *filter.InStock {
				continue
			}
		}
		result = append(result, p)
	}
	return result, nil
}

func GetCategories() ([]Category, error) {
	productsMux.RLock()
	defer productsMux.RUnlock()
	
	result := make([]Category, len(categories))
	copy(result, categories)
	return result, nil
}

// Mutation handlers
func CreateProduct(input ProductInput) (*Product, error) {
	// Validate category exists
	categoryExists := false
	for _, c := range categories {
		if c.ID == input.CategoryID {
			categoryExists = true
			break
		}
	}
	if !categoryExists {
		return nil, fmt.Errorf("category with id %d not found", input.CategoryID)
	}

	productsMux.Lock()
	defer productsMux.Unlock()

	product := Product{
		ID:          nextProdID,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Status:      ProductStatusDraft,
		CategoryID:  input.CategoryID,
		InStock:     false,
	}
	nextProdID++

	products = append(products, product)
	
	// Broadcast the product creation
	BroadcastProductUpdate(product, "created")
	
	return &product, nil
}

func UpdateProductStatus(id int, status ProductStatus) (*Product, error) {
	// Validate status
	switch status {
	case ProductStatusDraft, ProductStatusActive, ProductStatusDiscontinued, ProductStatusOutOfStock:
		// Valid status
	default:
		return nil, fmt.Errorf("invalid product status: %s", status)
	}

	productsMux.Lock()
	defer productsMux.Unlock()

	for i, p := range products {
		if p.ID == id {
			products[i].Status = status
			if status == ProductStatusOutOfStock {
				products[i].InStock = false
			} else if status == ProductStatusActive {
				products[i].InStock = true
			}
					// Broadcast the product update
				BroadcastProductUpdate(products[i], "updated")
				
			return &products[i], nil
		}
	}
	return nil, fmt.Errorf("product with id %d not found", id)
}

func AddProductReview(productId int, review ReviewInput) (*Review, error) {
	// Validate rating
	if review.Rating < 1 || review.Rating > 5 {
		return nil, errors.New("rating must be between 1 and 5")
	}

	// Validate product exists
	productExists := false
	for _, p := range products {
		if p.ID == productId {
			productExists = true
			break
		}
	}
	if !productExists {
		return nil, fmt.Errorf("product with id %d not found", productId)
	}

	productsMux.Lock()
	defer productsMux.Unlock()

	// In a real app, we'd get the user from context
	r := Review{
		ID:        nextRevID,
		ProductID: productId,
		UserID:    2, // Hardcoded for demo
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	nextRevID++

	reviews = append(reviews, r)
	return &r, nil
}

// Field resolvers
func (p *Product) Category() (*Category, error) {
	for _, c := range categories {
		if c.ID == p.CategoryID {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("category with id %d not found", p.CategoryID)
}

func (p *Product) Reviews() ([]Review, error) {
	var productReviews []Review
	for _, r := range reviews {
		if r.ProductID == p.ID {
			productReviews = append(productReviews, r)
		}
	}
	return productReviews, nil
}

func (p *Product) AverageRating() (*float64, error) {
	var total, count int
	for _, r := range reviews {
		if r.ProductID == p.ID {
			total += r.Rating
			count++
		}
	}
	if count == 0 {
		return nil, nil // No reviews
	}
	avg := float64(total) / float64(count)
	return &avg, nil
}

func (c *Category) Products() ([]Product, error) {
	var categoryProducts []Product
	for _, p := range products {
		if p.CategoryID == c.ID {
			categoryProducts = append(categoryProducts, p)
		}
	}
	return categoryProducts, nil
}

func (r *Review) User() (*User, error) {
	for _, u := range users {
		if u.ID == r.UserID {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user with id %d not found", r.UserID)
}

func (u *User) Reviews() ([]Review, error) {
	var userReviews []Review
	for _, r := range reviews {
		if r.UserID == u.ID {
			userReviews = append(userReviews, r)
		}
	}
	return userReviews, nil
}

// Helper function
func strPtr(s string) *string {
	return &s
}
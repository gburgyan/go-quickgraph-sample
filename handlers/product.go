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
func (ProductStatus) EnumValues() []quickgraph.EnumValue {
	return []quickgraph.EnumValue{
		{Name: "DRAFT", Description: "Product is in draft state and not yet available"},
		{Name: "ACTIVE", Description: "Product is active and available for purchase"},
		{Name: "DISCONTINUED", Description: "Product has been discontinued", IsDeprecated: true, DeprecationReason: "Use OUT_OF_STOCK for temporary unavailability"},
		{Name: "OUT_OF_STOCK", Description: "Product is temporarily out of stock"},
	}
}

type UserRole string

const (
	UserRoleAdmin    UserRole = "ADMIN"
	UserRoleCustomer UserRole = "CUSTOMER"
	UserRoleGuest    UserRole = "GUEST"
)

// EnumValues implements the StringEnumValues interface for schema generation
func (UserRole) EnumValues() []quickgraph.EnumValue {
	return []quickgraph.EnumValue{
		{Name: "ADMIN", Description: "Administrator with full access"},
		{Name: "CUSTOMER", Description: "Registered customer with purchase history"},
		{Name: "GUEST", Description: "Guest user with limited access"},
	}
}

// Domain types
type Product struct {
	ID          int           `graphy:"id,description=Unique product identifier"`
	Name        string        `graphy:"name,description=Product name"`
	Description string        `graphy:"description,description=Detailed product description"`
	Price       float64       `graphy:"price,description=Current price in USD"`
	Status      ProductStatus `graphy:"status,description=Current product status"`
	CategoryID  int           `graphy:"categoryId,description=ID of the category this product belongs to"`
	InStock     bool          `graphy:"inStock,description=Whether the product is currently in stock"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for Product
func (Product) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "A product available in the catalog",
	}
}

type Category struct {
	ID          int     `graphy:"id,description=Unique category identifier"`
	Name        string  `graphy:"name,description=Category name"`
	Description *string `graphy:"description,description=Optional category description"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for Category
func (Category) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Product category for organizing products",
	}
}

type Review struct {
	ID        int    `graphy:"id,description=Unique review identifier"`
	ProductID int    `graphy:"productId,description=ID of the reviewed product"`
	UserID    int    `graphy:"userId,description=ID of the user who wrote the review"`
	Rating    int    `graphy:"rating,description=Rating from 1 to 5 stars"`
	Comment   string `graphy:"comment,description=Review comment text"`
	CreatedAt string `graphy:"createdAt,description=ISO 8601 timestamp when the review was created"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for Review
func (Review) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Customer review for a product",
	}
}

type User struct {
	ID       int      `graphy:"id,description=Unique user identifier"`
	Username string   `graphy:"username,description=User's display name"`
	Email    string   `graphy:"email,description=User's email address"`
	Role     UserRole `graphy:"role,description=User's role in the system"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for User
func (User) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "System user who can view products and write reviews",
	}
}

// Input types
type ProductInput struct {
	Name        string  `graphy:"name,description=Product name"`
	Description string  `graphy:"description,description=Detailed product description"`
	Price       float64 `graphy:"price,description=Product price in USD"`
	CategoryID  int     `graphy:"categoryId,description=ID of the category for this product"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for ProductInput
func (ProductInput) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Input type for creating a new product",
	}
}

// Validate implements the Validator interface for ProductInput
func (p ProductInput) Validate() error {
	if p.Name == "" {
		return errors.New("product name is required")
	}
	if len(p.Name) < 3 {
		return errors.New("product name must be at least 3 characters long")
	}
	if p.Description == "" {
		return errors.New("product description is required")
	}
	if p.Price <= 0 {
		return errors.New("product price must be greater than 0")
	}
	if p.Price > 1000000 {
		return errors.New("product price cannot exceed $1,000,000")
	}
	// Category validation is done in the handler since it requires DB lookup
	return nil
}

type ProductFilter struct {
	CategoryID *int           `graphy:"categoryId,description=Filter by category ID"`
	MinPrice   *float64       `graphy:"minPrice,description=Minimum price filter"`
	MaxPrice   *float64       `graphy:"maxPrice,description=Maximum price filter"`
	Status     *ProductStatus `graphy:"status,description=Filter by product status"`
	InStock    *bool          `graphy:"inStock,description=Filter by stock availability"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for ProductFilter
func (ProductFilter) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Filter criteria for querying products",
	}
}

type ReviewInput struct {
	Rating  int    `graphy:"rating,description=Rating from 1 to 5 stars"`
	Comment string `graphy:"comment,description=Review comment text"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for ReviewInput
func (ReviewInput) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Input type for adding a product review",
	}
}

// Validate implements the Validator interface for ReviewInput
func (r ReviewInput) Validate() error {
	if r.Rating < 1 || r.Rating > 5 {
		return errors.New("rating must be between 1 and 5")
	}
	if r.Comment == "" {
		return errors.New("review comment is required")
	}
	if len(r.Comment) < 10 {
		return errors.New("review comment must be at least 10 characters long")
	}
	if len(r.Comment) > 1000 {
		return errors.New("review comment cannot exceed 1000 characters")
	}
	return nil
}

// ProductStatusUpdateInput is used for updating product status with admin validation
type ProductStatusUpdateInput struct {
	ID     int           `graphy:"id,description=Product ID to update"`
	Status ProductStatus `graphy:"status,description=New product status"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for ProductStatusUpdateInput
func (ProductStatusUpdateInput) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Input type for updating product status",
	}
}

// Validate implements the Validator interface for ProductStatusUpdateInput
func (p ProductStatusUpdateInput) Validate() error {
	// Validate the data fields only
	if p.ID <= 0 {
		return errors.New("invalid product ID")
	}

	switch p.Status {
	case ProductStatusDraft, ProductStatusActive, ProductStatusDiscontinued, ProductStatusOutOfStock:
		// Valid status
	default:
		return fmt.Errorf("invalid product status: %s", p.Status)
	}

	return nil
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
	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "GetProduct",
		Function:       GetProduct,
		ParameterNames: []string{"id"},
		Description:    strPtr("Retrieves a single product by its ID"),
		Mode:           quickgraph.ModeQuery,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "GetProducts",
		Function:       GetProducts,
		ParameterNames: []string{"filter"},
		Description:    strPtr("Retrieves products matching the specified filter criteria"),
		Mode:           quickgraph.ModeQuery,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:        "GetCategories",
		Function:    GetCategories,
		Description: strPtr("Retrieves all product categories"),
		Mode:        quickgraph.ModeQuery,
	})

	// Mutation registrations
	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "CreateProduct",
		Function:       CreateProduct,
		ParameterNames: []string{"input"},
		Description:    strPtr("Creates a new product in DRAFT status"),
		Mode:           quickgraph.ModeMutation,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "UpdateProductStatus",
		Function:       UpdateProductStatus,
		ParameterNames: []string{"input"},
		Description:    strPtr("Updates the status of an existing product"),
		Mode:           quickgraph.ModeMutation,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "AddProductReview",
		Function:       AddProductReview,
		ParameterNames: []string{"productId", "review"},
		Description:    strPtr("Adds a customer review to a product"),
		Mode:           quickgraph.ModeMutation,
	})

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
func CreateProduct(ctx context.Context, input ProductInput) (*Product, error) {
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

func UpdateProductStatus(ctx context.Context, input ProductStatusUpdateInput) (*Product, error) {
	// Check authentication
	userRole, ok := ctx.Value("userRole").(string)
	if !ok || userRole == "" {
		return nil, errors.New("authentication required to update product status")
	}

	// Check authorization - only admins can change product status
	if userRole != "admin" {
		return nil, errors.New("only administrators can update product status")
	}

	productsMux.Lock()
	defer productsMux.Unlock()

	for i, p := range products {
		if p.ID == input.ID {
			products[i].Status = input.Status
			if input.Status == ProductStatusOutOfStock {
				products[i].InStock = false
			} else if input.Status == ProductStatusActive {
				products[i].InStock = true
			}
			// Broadcast the product update
			BroadcastProductUpdate(products[i], "updated")

			return &products[i], nil
		}
	}
	return nil, fmt.Errorf("product with id %d not found", input.ID)
}

func AddProductReview(productId int, review ReviewInput) (*Review, error) {
	// Note: Rating validation is now handled by the Validator interface
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

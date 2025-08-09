package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gburgyan/go-quickgraph"
	"strings"
	"sync"
	"time"
)

// Employee base type - will be used as interface in GraphQL
type Employee struct {
	ID       int     `graphy:"id,description=Unique identifier for the employee"`
	Name     string  `graphy:"name,description=Full name of the employee"`
	Email    string  `graphy:"email,description=Work email address"`
	Salary   float64 `graphy:"salary,description=Annual salary in USD"`
	HireDate string  `graphy:"hireDate,description=Date when the employee was hired (ISO 8601 format)"`

	// Field for type discovery - allows runtime resolution of actual type
	actualType interface{} `graphy:"-"`
}

// ActualType implements the TypeDiscoverable interface for Employee
func (e *Employee) ActualType() interface{} {
	if e.actualType != nil {
		return e.actualType
	}
	return e
}

// GraphTypeExtension implements the GraphTypeExtension interface for Employee
func (Employee) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Base employee type representing common attributes for all employees",
	}
}

// Developer implements Employee interface via anonymous embedding
type Developer struct {
	Employee                      // Anonymous embedding for interface
	ProgrammingLanguages []string `graphy:"programmingLanguages,description=List of programming languages the developer is proficient in"`
	GithubUsername       *string  `graphy:"githubUsername,description=GitHub username for the developer (optional)"`
}

// NewDeveloper creates a new Developer with type discovery enabled
func NewDeveloper(id int, name, email string, salary float64, hireDate string, languages []string, github *string) *Developer {
	d := &Developer{
		Employee: Employee{
			ID:       id,
			Name:     name,
			Email:    email,
			Salary:   salary,
			HireDate: hireDate,
		},
		ProgrammingLanguages: languages,
		GithubUsername:       github,
	}
	d.Employee.actualType = d // Enable type discovery
	return d
}

// GraphTypeExtension implements the GraphTypeExtension interface for Developer
func (Developer) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "A software developer employee with programming skills and technical expertise",
	}
}

// Manager implements Employee interface via anonymous embedding
type Manager struct {
	Employee          // Anonymous embedding for interface
	Department string `graphy:"department,description=The department this manager oversees"`
	TeamSize   int    `graphy:"teamSize,description=Number of direct reports"`
}

// NewManager creates a new Manager with type discovery enabled
func NewManager(id int, name, email string, salary float64, hireDate string, department string, teamSize int) *Manager {
	m := &Manager{
		Employee: Employee{
			ID:       id,
			Name:     name,
			Email:    email,
			Salary:   salary,
			HireDate: hireDate,
		},
		Department: department,
		TeamSize:   teamSize,
	}
	m.Employee.actualType = m // Enable type discovery
	return m
}

// GraphTypeExtension implements the GraphTypeExtension interface for Manager
func (Manager) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "A management employee responsible for overseeing teams and departments",
	}
}

// Enums are represented as string types with constants
type EmployeeType string

const (
	EmployeeTypeDeveloper EmployeeType = "DEVELOPER"
	EmployeeTypeManager   EmployeeType = "MANAGER"
)

// EnumValues implements the StringEnumValues interface for schema generation
func (EmployeeType) EnumValues() []quickgraph.EnumValue {
	return []quickgraph.EnumValue{
		{Name: "DEVELOPER", Description: "Software developer employee type"},
		{Name: "MANAGER", Description: "Management employee type"},
	}
}

// Input types
type EmployeeInput struct {
	Name                 string       `graphy:"name,description=Full name of the employee"`
	Email                string       `graphy:"email,description=Work email address"`
	Salary               float64      `graphy:"salary,description=Annual salary in USD"`
	Type                 EmployeeType `graphy:"type,description=Type of employee (DEVELOPER or MANAGER)"`
	ProgrammingLanguages []string     `graphy:"programmingLanguages,description=Required for developers: list of programming languages"`
	GithubUsername       *string      `graphy:"githubUsername,description=Optional GitHub username for developers"`
	Department           *string      `graphy:"department,description=Required for managers: department name"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for EmployeeInput
func (EmployeeInput) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Input type for creating a new employee",
	}
}

// Validate implements the Validator interface for EmployeeInput
func (e EmployeeInput) Validate() error {
	// Validate required fields
	if e.Name == "" {
		return errors.New("name is required")
	}
	if len(e.Name) < 2 {
		return errors.New("name must be at least 2 characters long")
	}
	if e.Email == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(e.Email, "@") || !strings.Contains(e.Email, ".") {
		return errors.New("invalid email format")
	}
	if e.Salary <= 0 {
		return errors.New("salary must be greater than 0")
	}

	// Type-specific validation
	switch e.Type {
	case EmployeeTypeDeveloper:
		if len(e.ProgrammingLanguages) == 0 {
			return errors.New("developers must have at least one programming language")
		}
		for _, lang := range e.ProgrammingLanguages {
			if lang == "" {
				return errors.New("programming language cannot be empty")
			}
		}
	case EmployeeTypeManager:
		if e.Department == nil || *e.Department == "" {
			return errors.New("managers must have a department")
		}
	default:
		return fmt.Errorf("invalid employee type: %s", e.Type)
	}

	return nil
}

// EmployeeResultUnion represents the possible results when creating an employee
// The "Union" suffix tells the library this is a GraphQL union type
type EmployeeResultUnion struct {
	Developer *Developer
	Manager   *Manager
}

// GraphTypeExtension implements the GraphTypeExtension interface for EmployeeResultUnion
func (EmployeeResultUnion) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Union type representing either a Developer or Manager result",
	}
}

var (
	employees   []interface{} // Stores both Developer and Manager
	employeeMux sync.RWMutex
	nextEmpID   = 1
)

func init() {
	// Initialize with sample data
	github := "johndoe"
	employees = []interface{}{
		NewDeveloper(
			1,
			"John Doe",
			"john@example.com",
			120000,
			"2020-01-15",
			[]string{"Go", "Python", "JavaScript"},
			&github,
		),
		NewManager(
			2,
			"Jane Smith",
			"jane@example.com",
			150000,
			"2019-06-01",
			"Engineering",
			5,
		),
		NewDeveloper(
			3,
			"Bob Wilson",
			"bob@example.com",
			110000,
			"2021-03-20",
			[]string{"Go", "Rust"},
			nil,
		),
	}
	nextEmpID = 4
}

func RegisterEmployeeHandlers(ctx context.Context, graphy *quickgraph.Graphy) {
	// Query registrations
	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "GetEmployee",
		Function:       GetEmployee,
		ParameterNames: []string{"id"},
		Description:    strPtr("Retrieves a single employee by their unique ID"),
		Mode:           quickgraph.ModeQuery,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:        "GetAllEmployees",
		Function:    GetAllEmployees,
		Description: strPtr("Retrieves all employees in the system"),
		Mode:        quickgraph.ModeQuery,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:        "GetManagers",
		Function:    GetManagers,
		Description: strPtr("Retrieves all employees who are managers"),
		Mode:        quickgraph.ModeQuery,
	})

	// Mutation registrations
	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "CreateEmployee",
		Function:       CreateEmployee,
		ParameterNames: []string{"input"},
		Description:    strPtr("Creates a new employee (Developer or Manager) based on the input type"),
		Mode:           quickgraph.ModeMutation,
	})

	graphy.RegisterFunction(ctx, quickgraph.FunctionDefinition{
		Name:           "PromoteToManager",
		Function:       PromoteToManager,
		ParameterNames: []string{"input"},
		Description:    strPtr("Promotes a developer to a manager role with a 20% salary increase"),
		Mode:           quickgraph.ModeMutation,
	})

	// Note: The Reports() method on Manager will be automatically exposed as a field
	// when a Manager object is returned from a query
}

// GetEmployee returns a single employee by ID
// This demonstrates type discovery - we return *Employee but the actual type
// (Developer or Manager) is discoverable at runtime
func GetEmployee(id int) (*Employee, error) {
	employeeMux.RLock()
	defer employeeMux.RUnlock()

	for _, emp := range employees {
		switch e := emp.(type) {
		case *Developer:
			if e.ID == id {
				return &e.Employee, nil
			}
		case *Manager:
			if e.ID == id {
				return &e.Employee, nil
			}
		}
	}

	return nil, fmt.Errorf("employee with id %d not found", id)
}

// GetAllEmployees returns all employees
// This also demonstrates type discovery with slices - we return []*Employee
// but each element's actual type (Developer or Manager) is discoverable
func GetAllEmployees() ([]*Employee, error) {
	employeeMux.RLock()
	defer employeeMux.RUnlock()

	result := make([]*Employee, 0, len(employees))
	for _, emp := range employees {
		switch e := emp.(type) {
		case *Developer:
			result = append(result, &e.Employee)
		case *Manager:
			result = append(result, &e.Employee)
		}
	}

	return result, nil
}

// GetManagers returns only managers
func GetManagers() ([]*Manager, error) {
	employeeMux.RLock()
	defer employeeMux.RUnlock()

	var managers []*Manager
	for _, emp := range employees {
		if mgr, ok := emp.(*Manager); ok {
			managers = append(managers, mgr)
		}
	}

	return managers, nil
}

// Reports method for Manager - demonstrates field resolution with type discovery
func (m *Manager) Reports() ([]*Employee, error) {
	employeeMux.RLock()
	defer employeeMux.RUnlock()

	var reports []*Employee
	// In a real app, this would query by manager ID
	// For demo, return some developers
	for _, emp := range employees {
		if dev, ok := emp.(*Developer); ok && len(reports) < m.TeamSize {
			reports = append(reports, &dev.Employee)
		}
	}

	return reports, nil
}

// Helper function
func strPtr(s string) *string {
	return &s
}

// CreateEmployee mutation - returns a union of Developer or Manager
func CreateEmployee(input EmployeeInput) (EmployeeResultUnion, error) {
	// Note: Validation is now handled automatically by the Validator interface
	employeeMux.Lock()
	defer employeeMux.Unlock()

	nextID := nextEmpID
	nextEmpID++

	switch input.Type {
	case EmployeeTypeDeveloper:
		dev := NewDeveloper(
			nextID,
			input.Name,
			input.Email,
			input.Salary,
			time.Now().Format("2006-01-02"),
			input.ProgrammingLanguages,
			input.GithubUsername,
		)
		employees = append(employees, dev)
		return EmployeeResultUnion{Developer: dev}, nil
	case EmployeeTypeManager:
		mgr := NewManager(
			nextID,
			input.Name,
			input.Email,
			input.Salary,
			time.Now().Format("2006-01-02"),
			*input.Department,
			0, // Start with no reports
		)
		employees = append(employees, mgr)
		return EmployeeResultUnion{Manager: mgr}, nil
	default:
		return EmployeeResultUnion{}, fmt.Errorf("invalid employee type: %s", input.Type)
	}
}

// PromotionInput represents the input for promoting an employee
type PromotionInput struct {
	EmployeeID int    `graphy:"employeeId,description=ID of the employee to promote"`
	Department string `graphy:"department,description=Department the new manager will oversee"`
}

// GraphTypeExtension implements the GraphTypeExtension interface for PromotionInput
func (PromotionInput) GraphTypeExtension() quickgraph.GraphTypeInfo {
	return quickgraph.GraphTypeInfo{
		Description: "Input type for promoting an employee to manager",
	}
}

// Validate implements the Validator interface for PromotionInput
func (p PromotionInput) Validate() error {
	// Validate input fields only
	if p.EmployeeID <= 0 {
		return errors.New("invalid employee ID")
	}
	if p.Department == "" {
		return errors.New("department is required")
	}
	if len(p.Department) < 2 {
		return errors.New("department name must be at least 2 characters long")
	}

	return nil
}

// PromoteToManager mutation - demonstrates type transformation with auth validation
func PromoteToManager(ctx context.Context, input PromotionInput) (*Manager, error) {
	// Check authentication
	userRole, ok := ctx.Value("userRole").(string)
	if !ok || userRole == "" {
		return nil, errors.New("authentication required for promotions")
	}

	// Check authorization
	if userRole != "admin" {
		return nil, errors.New("only administrators can promote employees")
	}

	employeeMux.Lock()
	defer employeeMux.Unlock()

	for i, emp := range employees {
		if dev, ok := emp.(*Developer); ok && dev.ID == input.EmployeeID {
			// Create new manager from developer
			mgr := &Manager{
				Employee:   dev.Employee,
				Department: input.Department,
				TeamSize:   0,
			}
			mgr.Salary *= 1.2 // 20% raise with promotion

			employees[i] = mgr
			return mgr, nil
		}
	}

	return nil, fmt.Errorf("developer with id %d not found", input.EmployeeID)
}

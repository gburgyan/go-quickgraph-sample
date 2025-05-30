package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gburgyan/go-quickgraph"
	"sync"
	"time"
)

// Employee base type - will be used as interface in GraphQL
type Employee struct {
	ID       int
	Name     string
	Email    string
	Salary   float64
	HireDate string

	// Field for type discovery - allows runtime resolution of actual type
	actualType interface{} `json:"-" graphy:"-"`
}

// ActualType implements the TypeDiscoverable interface for Employee
func (e *Employee) ActualType() interface{} {
	if e.actualType != nil {
		return e.actualType
	}
	return e
}

// Developer implements Employee interface via anonymous embedding
type Developer struct {
	Employee             // Anonymous embedding for interface
	ProgrammingLanguages []string
	GithubUsername       *string // Optional field
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

// Manager implements Employee interface via anonymous embedding
type Manager struct {
	Employee   // Anonymous embedding for interface
	Department string
	TeamSize   int
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

// Enums are represented as string types with constants
type EmployeeType string

const (
	EmployeeTypeDeveloper EmployeeType = "DEVELOPER"
	EmployeeTypeManager   EmployeeType = "MANAGER"
)

// EnumValues implements the StringEnumValues interface for schema generation
func (EmployeeType) EnumValues() []string {
	return []string{"DEVELOPER", "MANAGER"}
}

// Input types
type EmployeeInput struct {
	Name                 string       `json:"name"`
	Email                string       `json:"email"`
	Salary               float64      `json:"salary"`
	Type                 EmployeeType `json:"type"`
	ProgrammingLanguages []string     `json:"programmingLanguages"`
	GithubUsername       *string      `json:"githubUsername"`
	Department           *string      `json:"department"`
}

// EmployeeResultUnion represents the possible results when creating an employee
// The "Union" suffix tells the library this is a GraphQL union type
type EmployeeResultUnion struct {
	Developer *Developer
	Manager   *Manager
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
	graphy.RegisterQuery(ctx, "GetEmployee", GetEmployee, "id")
	graphy.RegisterQuery(ctx, "GetAllEmployees", GetAllEmployees)
	graphy.RegisterQuery(ctx, "GetManagers", GetManagers)

	// Mutation registrations
	graphy.RegisterMutation(ctx, "CreateEmployee", CreateEmployee, "input")
	graphy.RegisterMutation(ctx, "PromoteToManager", PromoteToManager, "employeeId", "department")

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

// CreateEmployee mutation - returns a union of Developer or Manager
func CreateEmployee(input EmployeeInput) (EmployeeResultUnion, error) {
	// Validate input
	if input.Type == EmployeeTypeDeveloper && len(input.ProgrammingLanguages) == 0 {
		return EmployeeResultUnion{}, errors.New("developers must have at least one programming language")
	}
	if input.Type == EmployeeTypeManager && (input.Department == nil || *input.Department == "") {
		return EmployeeResultUnion{}, errors.New("managers must have a department")
	}

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

// PromoteToManager mutation - demonstrates type transformation
func PromoteToManager(employeeId int, department string) (*Manager, error) {
	employeeMux.Lock()
	defer employeeMux.Unlock()

	for i, emp := range employees {
		if dev, ok := emp.(*Developer); ok && dev.ID == employeeId {
			// Create new manager from developer
			mgr := &Manager{
				Employee:   dev.Employee,
				Department: department,
				TeamSize:   0,
			}
			mgr.Salary *= 1.2 // 20% raise with promotion

			employees[i] = mgr
			return mgr, nil
		}
	}

	return nil, fmt.Errorf("developer with id %d not found", employeeId)
}

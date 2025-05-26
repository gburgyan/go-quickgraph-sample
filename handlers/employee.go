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
}

// Developer implements Employee interface via anonymous embedding
type Developer struct {
	Employee             // Anonymous embedding for interface
	ProgrammingLanguages []string
	GithubUsername       *string // Optional field
}

// Manager implements Employee interface via anonymous embedding
type Manager struct {
	Employee   // Anonymous embedding for interface
	Department string
	TeamSize   int
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

var (
	employees   []interface{} // Stores both Developer and Manager
	employeeMux sync.RWMutex
	nextEmpID   = 1
)

func init() {
	// Initialize with sample data
	github := "johndoe"
	employees = []interface{}{
		&Developer{
			Employee: Employee{
				ID:       1,
				Name:     "John Doe",
				Email:    "john@example.com",
				Salary:   120000,
				HireDate: "2020-01-15",
			},
			ProgrammingLanguages: []string{"Go", "Python", "JavaScript"},
			GithubUsername:       &github,
		},
		&Manager{
			Employee: Employee{
				ID:       2,
				Name:     "Jane Smith",
				Email:    "jane@example.com",
				Salary:   150000,
				HireDate: "2019-06-01",
			},
			Department: "Engineering",
			TeamSize:   5,
		},
		&Developer{
			Employee: Employee{
				ID:       3,
				Name:     "Bob Wilson",
				Email:    "bob@example.com",
				Salary:   110000,
				HireDate: "2021-03-20",
			},
			ProgrammingLanguages: []string{"Go", "Rust"},
			GithubUsername:       nil,
		},
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
func GetEmployee(id int) (Employee, error) {
	employeeMux.RLock()
	defer employeeMux.RUnlock()

	for _, emp := range employees {
		switch e := emp.(type) {
		case *Developer:
			if e.ID == id {
				return e.Employee, nil
			}
		case *Manager:
			if e.ID == id {
				return e.Employee, nil
			}
		}
	}

	return Employee{}, fmt.Errorf("employee with id %d not found", id)
}

// GetAllEmployees returns all employees
func GetAllEmployees() ([]Employee, error) {
	employeeMux.RLock()
	defer employeeMux.RUnlock()

	result := make([]Employee, 0, len(employees))
	for _, emp := range employees {
		switch e := emp.(type) {
		case *Developer:
			result = append(result, e.Employee)
		case *Manager:
			result = append(result, e.Employee)
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

// Reports method for Manager - demonstrates field resolution
func (m *Manager) Reports() ([]Employee, error) {
	employeeMux.RLock()
	defer employeeMux.RUnlock()

	var reports []Employee
	// In a real app, this would query by manager ID
	// For demo, return some developers
	for _, emp := range employees {
		if dev, ok := emp.(*Developer); ok && len(reports) < m.TeamSize {
			reports = append(reports, dev.Employee)
		}
	}

	return reports, nil
}

// CreateEmployee mutation - returns multiple pointers (only one non-nil) for implicit union
func CreateEmployee(input EmployeeInput) (*Developer, *Manager, error) {
	// Validate input
	if input.Type == EmployeeTypeDeveloper && len(input.ProgrammingLanguages) == 0 {
		return nil, nil, errors.New("developers must have at least one programming language")
	}
	if input.Type == EmployeeTypeManager && (input.Department == nil || *input.Department == "") {
		return nil, nil, errors.New("managers must have a department")
	}

	employeeMux.Lock()
	defer employeeMux.Unlock()

	emp := Employee{
		ID:       nextEmpID,
		Name:     input.Name,
		Email:    input.Email,
		Salary:   input.Salary,
		HireDate: time.Now().Format("2006-01-02"),
	}
	nextEmpID++

	switch input.Type {
	case EmployeeTypeDeveloper:
		dev := &Developer{
			Employee:             emp,
			ProgrammingLanguages: input.ProgrammingLanguages,
			GithubUsername:       input.GithubUsername,
		}
		employees = append(employees, dev)
		return dev, nil, nil
	case EmployeeTypeManager:
		mgr := &Manager{
			Employee:   emp,
			Department: *input.Department,
			TeamSize:   0, // Start with no reports
		}
		employees = append(employees, mgr)
		return nil, mgr, nil
	default:
		return nil, nil, fmt.Errorf("invalid employee type: %s", input.Type)
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

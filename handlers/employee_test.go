package handlers

import (
	"testing"

	"github.com/gburgyan/go-quickgraph"
)

func TestEmployeeTypeDiscovery(t *testing.T) {
	// Create test employees
	github := "testuser"
	dev := NewDeveloper(
		1,
		"John Developer",
		"john@example.com",
		100000,
		"2023-01-01",
		[]string{"Go", "Python"},
		&github,
	)

	mgr := NewManager(
		2,
		"Jane Manager",
		"jane@example.com",
		120000,
		"2022-01-01",
		"Engineering",
		5,
	)

	// Test Developer type discovery
	t.Run("Developer type discovery", func(t *testing.T) {
		empPtr := &dev.Employee

		// Should implement TypeDiscoverable
		if _, ok := interface{}(empPtr).(quickgraph.TypeDiscoverable); !ok {
			t.Error("Employee should implement TypeDiscoverable")
		}

		// Should discover as Developer
		discovered, ok := quickgraph.Discover[*Developer](empPtr)
		if !ok {
			t.Error("Should be able to discover Developer from Employee pointer")
		}
		if discovered == nil {
			t.Error("Discovered Developer should not be nil")
		}
		if discovered != dev {
			t.Error("Discovered Developer should be the same instance")
		}

		// Should have access to Developer fields
		if len(discovered.ProgrammingLanguages) != 2 {
			t.Errorf("Expected 2 programming languages, got %d", len(discovered.ProgrammingLanguages))
		}
		if discovered.GithubUsername == nil || *discovered.GithubUsername != "testuser" {
			t.Error("GithubUsername not preserved correctly")
		}
	})

	// Test Manager type discovery
	t.Run("Manager type discovery", func(t *testing.T) {
		empPtr := &mgr.Employee

		// Should discover as Manager
		discovered, ok := quickgraph.Discover[*Manager](empPtr)
		if !ok {
			t.Error("Should be able to discover Manager from Employee pointer")
		}
		if discovered == nil {
			t.Error("Discovered Manager should not be nil")
		}
		if discovered != mgr {
			t.Error("Discovered Manager should be the same instance")
		}

		// Should have access to Manager fields
		if discovered.Department != "Engineering" {
			t.Errorf("Expected Department to be Engineering, got %s", discovered.Department)
		}
		if discovered.TeamSize != 5 {
			t.Errorf("Expected TeamSize to be 5, got %d", discovered.TeamSize)
		}
	})

	// Test wrong type discovery
	t.Run("Wrong type discovery", func(t *testing.T) {
		empPtr := &dev.Employee

		// Should not discover Developer as Manager
		_, ok := quickgraph.Discover[*Manager](empPtr)
		if ok {
			t.Error("Should not be able to discover Developer as Manager")
		}
	})

	// Test slice of employees
	t.Run("Slice type discovery", func(t *testing.T) {
		employees := []*Employee{
			&dev.Employee,
			&mgr.Employee,
		}

		// First should be Developer
		if dev, ok := quickgraph.Discover[*Developer](employees[0]); ok {
			if dev.Name != "John Developer" {
				t.Errorf("Expected name John Developer, got %s", dev.Name)
			}
		} else {
			t.Error("First employee should be discoverable as Developer")
		}

		// Second should be Manager
		if mgr, ok := quickgraph.Discover[*Manager](employees[1]); ok {
			if mgr.Name != "Jane Manager" {
				t.Errorf("Expected name Jane Manager, got %s", mgr.Name)
			}
		} else {
			t.Error("Second employee should be discoverable as Manager")
		}
	})
}

func TestEmployeeActualType(t *testing.T) {
	dev := NewDeveloper(1, "Test", "test@example.com", 100000, "2023-01-01", []string{"Go"}, nil)
	emp := &dev.Employee

	// ActualType should return the Developer instance
	actual := emp.ActualType()
	if actual != dev {
		t.Error("ActualType should return the Developer instance")
	}

	// Test with base Employee (no actualType set)
	baseEmp := &Employee{
		ID:    1,
		Name:  "Base",
		Email: "base@example.com",
	}
	actual = baseEmp.ActualType()
	if actual != baseEmp {
		t.Error("ActualType should return self when actualType is nil")
	}
}

func TestCreateEmployeeWithTypeDiscovery(t *testing.T) {
	// Test creating Developer
	t.Run("Create Developer", func(t *testing.T) {
		input := EmployeeInput{
			Name:                 "New Developer",
			Email:                "newdev@example.com",
			Salary:               90000,
			Type:                 EmployeeTypeDeveloper,
			ProgrammingLanguages: []string{"Go", "Rust"},
		}

		result, err := CreateEmployee(input)
		if err != nil {
			t.Fatalf("CreateEmployee failed: %v", err)
		}

		if result.Developer == nil {
			t.Fatal("Expected Developer in result")
		}
		if result.Manager != nil {
			t.Fatal("Expected no Manager in result")
		}

		// Verify type discovery works on the created employee
		emp := &result.Developer.Employee
		discovered, ok := quickgraph.Discover[*Developer](emp)
		if !ok {
			t.Error("Should be able to discover created Developer")
		}
		if discovered.Name != "New Developer" {
			t.Errorf("Expected name New Developer, got %s", discovered.Name)
		}
	})

	// Test creating Manager
	t.Run("Create Manager", func(t *testing.T) {
		dept := "Sales"
		input := EmployeeInput{
			Name:       "New Manager",
			Email:      "newmgr@example.com",
			Salary:     110000,
			Type:       EmployeeTypeManager,
			Department: &dept,
		}

		result, err := CreateEmployee(input)
		if err != nil {
			t.Fatalf("CreateEmployee failed: %v", err)
		}

		if result.Manager == nil {
			t.Fatal("Expected Manager in result")
		}
		if result.Developer != nil {
			t.Fatal("Expected no Developer in result")
		}

		// Verify type discovery works on the created employee
		emp := &result.Manager.Employee
		discovered, ok := quickgraph.Discover[*Manager](emp)
		if !ok {
			t.Error("Should be able to discover created Manager")
		}
		if discovered.Department != "Sales" {
			t.Errorf("Expected department Sales, got %s", discovered.Department)
		}
	})
}

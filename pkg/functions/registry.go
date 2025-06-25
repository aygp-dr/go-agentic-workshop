package functions

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Function represents a callable function
type Function struct {
	Name        string
	Description string
	Parameters  json.RawMessage
	Handler     interface{}
}

// Registry manages available functions
type Registry struct {
	functions map[string]*Function
}

// NewRegistry creates a new function registry
func NewRegistry() *Registry {
	return &Registry{
		functions: make(map[string]*Function),
	}
}

// Register adds a function to the registry
func (r *Registry) Register(fn *Function) error {
	if _, exists := r.functions[fn.Name]; exists {
		return fmt.Errorf("function %s already registered", fn.Name)
	}

	// Validate handler is a function
	if reflect.TypeOf(fn.Handler).Kind() != reflect.Func {
		return fmt.Errorf("handler must be a function")
	}

	r.functions[fn.Name] = fn
	return nil
}

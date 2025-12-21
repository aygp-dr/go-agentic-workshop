package functions

import (
	"context"
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

// FunctionManifest represents a function's metadata for LLM consumption
type FunctionManifest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// GetManifest returns the manifest of all registered functions
func (r *Registry) GetManifest() []FunctionManifest {
	manifests := make([]FunctionManifest, 0, len(r.functions))
	for _, fn := range r.functions {
		manifests = append(manifests, FunctionManifest{
			Name:        fn.Name,
			Description: fn.Description,
			Parameters:  fn.Parameters,
		})
	}
	return manifests
}

// Execute calls the named function with the given arguments
func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	fn, exists := r.functions[name]
	if !exists {
		return nil, fmt.Errorf("function %s not found", name)
	}

	// Call the handler function with the args
	handler, ok := fn.Handler.(func(map[string]interface{}) (interface{}, error))
	if !ok {
		return nil, fmt.Errorf("handler for %s has wrong signature", name)
	}

	_ = ctx // Reserved for future context propagation
	return handler(args)
}

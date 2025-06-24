package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MockLLMClient simulates LLM responses for testing
type MockLLMClient struct {
	mu            sync.RWMutex
	responses     map[string]string
	errors        map[string]error
	callCount     map[string]int
	latency       time.Duration
	tokenCount    int
	costPerToken  float64
	failureRate   float64
	StreamingMode bool
}

// NewMockLLMClient creates a new mock LLM client
func NewMockLLMClient() *MockLLMClient {
	return &MockLLMClient{
		responses:    make(map[string]string),
		errors:       make(map[string]error),
		callCount:    make(map[string]int),
		latency:      100 * time.Millisecond,
		costPerToken: 0.00002, // $0.02 per 1K tokens
	}
}

// SetResponse sets a canned response for a prompt pattern
func (m *MockLLMClient) SetResponse(promptPattern, response string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[promptPattern] = response
}

// SetError sets an error response for a prompt pattern
func (m *MockLLMClient) SetError(promptPattern string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[promptPattern] = err
}

// SetLatency sets the simulated response latency
func (m *MockLLMClient) SetLatency(latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latency = latency
}

// SetFailureRate sets the random failure rate (0.0 to 1.0)
func (m *MockLLMClient) SetFailureRate(rate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failureRate = rate
}

// SetCostPerToken sets the cost per token for cost estimation
func (m *MockLLMClient) SetCostPerToken(cost float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.costPerToken = cost
}

// Call simulates an LLM API call
func (m *MockLLMClient) Call(ctx context.Context, prompt string) (string, error) {
	m.mu.Lock()
	m.callCount[prompt]++
	latency := m.latency
	m.mu.Unlock()

	// Simulate latency
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(latency):
	}

	// Check for specific errors
	m.mu.RLock()
	if err, ok := m.errors[prompt]; ok {
		m.mu.RUnlock()
		return "", err
	}

	// Check for specific responses
	for pattern, response := range m.responses {
		if strings.Contains(prompt, pattern) {
			m.mu.RUnlock()
			m.updateTokenCount(len(prompt) + len(response))
			return response, nil
		}
	}
	m.mu.RUnlock()

	// Default response based on prompt content
	response := m.generateDefaultResponse(prompt)
	m.updateTokenCount(len(prompt) + len(response))

	return response, nil
}

// GetCallCount returns the number of calls for a prompt
func (m *MockLLMClient) GetCallCount(prompt string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount[prompt]
}

// GetTotalCalls returns total number of calls
func (m *MockLLMClient) GetTotalCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	total := 0
	for _, count := range m.callCount {
		total += count
	}
	return total
}

// GetEstimatedCost returns the estimated cost based on token usage
func (m *MockLLMClient) GetEstimatedCost() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return float64(m.tokenCount) * m.costPerToken
}

// Reset clears all mock data
func (m *MockLLMClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses = make(map[string]string)
	m.errors = make(map[string]error)
	m.callCount = make(map[string]int)
	m.tokenCount = 0
}

func (m *MockLLMClient) updateTokenCount(chars int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Rough approximation: 1 token ≈ 4 characters
	m.tokenCount += chars / 4
}

func (m *MockLLMClient) generateDefaultResponse(prompt string) string {
	// Generate contextual responses based on prompt content
	switch {
	case strings.Contains(prompt, "calculate"):
		return `{"action": "calculate", "expression": "2+2", "result": 4}`
	case strings.Contains(prompt, "search"):
		return `{"action": "search", "query": "test", "results": ["result1", "result2"]}`
	case strings.Contains(prompt, "plan"):
		return `{"plan": ["step1: analyze", "step2: execute", "step3: verify"]}`
	case strings.Contains(prompt, "error"):
		return `{"error": "simulated error", "suggestion": "check configuration"}`
	default:
		return `{"response": "I understand your request. This is a mock response."}`
	}
}

// MockFunctionRegistry simulates function execution for testing
type MockFunctionRegistry struct {
	mu        sync.RWMutex
	functions map[string]MockFunction
	callLog   []FunctionCall
}

// MockFunction represents a mock function
type MockFunction struct {
	Name        string
	Handler     func(args map[string]interface{}) (interface{}, error)
	CallCount   int
	LastArgs    map[string]interface{}
	ShouldError bool
	ErrorMsg    string
	Latency     time.Duration
}

// FunctionCall logs function execution
type FunctionCall struct {
	Name      string
	Args      map[string]interface{}
	Result    interface{}
	Error     error
	Timestamp time.Time
	Duration  time.Duration
}

// NewMockFunctionRegistry creates a new mock function registry
func NewMockFunctionRegistry() *MockFunctionRegistry {
	return &MockFunctionRegistry{
		functions: make(map[string]MockFunction),
		callLog:   make([]FunctionCall, 0),
	}
}

// Register adds a mock function
func (r *MockFunctionRegistry) Register(name string, handler func(args map[string]interface{}) (interface{}, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.functions[name] = MockFunction{
		Name:    name,
		Handler: handler,
		Latency: 10 * time.Millisecond,
	}
}

// Execute runs a mock function
func (r *MockFunctionRegistry) Execute(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	r.mu.Lock()
	fn, exists := r.functions[name]
	if !exists {
		r.mu.Unlock()
		return nil, fmt.Errorf("function not found: %s", name)
	}

	fn.CallCount++
	fn.LastArgs = args
	r.functions[name] = fn
	r.mu.Unlock()

	start := time.Now()

	// Simulate latency
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(fn.Latency):
	}

	// Check if should error
	if fn.ShouldError {
		err := fmt.Errorf(fn.ErrorMsg)
		r.logCall(name, args, nil, err, time.Since(start))
		return nil, err
	}

	// Execute handler
	result, err := fn.Handler(args)
	r.logCall(name, args, result, err, time.Since(start))

	return result, err
}

// GetCallLog returns the function call log
func (r *MockFunctionRegistry) GetCallLog() []FunctionCall {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]FunctionCall{}, r.callLog...)
}

// SetError configures a function to error
func (r *MockFunctionRegistry) SetError(name string, errorMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if fn, ok := r.functions[name]; ok {
		fn.ShouldError = true
		fn.ErrorMsg = errorMsg
		r.functions[name] = fn
	}
}

// Reset clears all mock data
func (r *MockFunctionRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.functions = make(map[string]MockFunction)
	r.callLog = make([]FunctionCall, 0)
}

func (r *MockFunctionRegistry) logCall(name string, args map[string]interface{}, result interface{}, err error, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.callLog = append(r.callLog, FunctionCall{
		Name:      name,
		Args:      args,
		Result:    result,
		Error:     err,
		Timestamp: time.Now(),
		Duration:  duration,
	})
}

// MockWorkflowState provides test workflow state management
type MockWorkflowState struct {
	mu      sync.RWMutex
	states  map[string]interface{}
	history []StateChange
}

// StateChange represents a state transition
type StateChange struct {
	Key       string
	OldValue  interface{}
	NewValue  interface{}
	Timestamp time.Time
}

// NewMockWorkflowState creates a new mock workflow state
func NewMockWorkflowState() *MockWorkflowState {
	return &MockWorkflowState{
		states:  make(map[string]interface{}),
		history: make([]StateChange, 0),
	}
}

// Get retrieves a state value
func (s *MockWorkflowState) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.states[key]
	return val, ok
}

// Set updates a state value
func (s *MockWorkflowState) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldValue := s.states[key]
	s.states[key] = value

	s.history = append(s.history, StateChange{
		Key:       key,
		OldValue:  oldValue,
		NewValue:  value,
		Timestamp: time.Now(),
	})
}

// GetHistory returns state change history
func (s *MockWorkflowState) GetHistory() []StateChange {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]StateChange{}, s.history...)
}

// ToJSON exports state as JSON
func (s *MockWorkflowState) ToJSON() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, err := json.Marshal(s.states)
	return string(data), err
}

// Reset clears all state
func (s *MockWorkflowState) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states = make(map[string]interface{})
	s.history = make([]StateChange, 0)
}

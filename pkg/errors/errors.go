package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorType represents categories of errors
type ErrorType string

const (
	// Infrastructure errors
	ErrorTypeNetwork  ErrorType = "NETWORK"
	ErrorTypeAWS      ErrorType = "AWS"
	ErrorTypeDatabase ErrorType = "DATABASE"
	ErrorTypeDocker   ErrorType = "DOCKER"

	// LLM errors
	ErrorTypeLLMTimeout   ErrorType = "LLM_TIMEOUT"
	ErrorTypeLLMRateLimit ErrorType = "LLM_RATE_LIMIT"
	ErrorTypeLLMInvalid   ErrorType = "LLM_INVALID"
	ErrorTypeLLMCost      ErrorType = "LLM_COST"

	// Agent errors
	ErrorTypeAgentState  ErrorType = "AGENT_STATE"
	ErrorTypeAgentLoop   ErrorType = "AGENT_LOOP"
	ErrorTypeAgentMemory ErrorType = "AGENT_MEMORY"

	// Function errors
	ErrorTypeFunctionNotFound ErrorType = "FUNCTION_NOT_FOUND"
	ErrorTypeFunctionExec     ErrorType = "FUNCTION_EXEC"
	ErrorTypeFunctionTimeout  ErrorType = "FUNCTION_TIMEOUT"

	// Validation errors
	ErrorTypeValidation ErrorType = "VALIDATION"
	ErrorTypePermission ErrorType = "PERMISSION"
)

// AgentError provides rich error context for debugging
type AgentError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Suggestion string                 `json:"suggestion,omitempty"`
	Retryable  bool                   `json:"retryable"`
	Stack      string                 `json:"stack,omitempty"`
	Wrapped    error                  `json:"-"`
}

// Error implements the error interface
func (e *AgentError) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s: %s (suggestion: %s)", e.Type, e.Message, e.Suggestion)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *AgentError) Unwrap() error {
	return e.Wrapped
}

// WithContext adds context to the error
func (e *AgentError) WithContext(key string, value interface{}) *AgentError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// ToJSON converts error to JSON for logging
func (e *AgentError) ToJSON() string {
	data, _ := json.Marshal(e)
	return string(data)
}

// New creates a new AgentError
func New(errorType ErrorType, message string) *AgentError {
	return &AgentError{
		Type:      errorType,
		Message:   message,
		Timestamp: time.Now(),
		Stack:     getStackTrace(),
		Retryable: isRetryable(errorType),
	}
}

// Wrap wraps an existing error with agent context
func Wrap(err error, errorType ErrorType, message string) *AgentError {
	if err == nil {
		return nil
	}

	agentErr := New(errorType, message)
	agentErr.Wrapped = err

	// If wrapping another AgentError, preserve context
	if ae, ok := err.(*AgentError); ok {
		for k, v := range ae.Context {
			_ = agentErr.WithContext(k, v)
		}
	}

	return agentErr
}

// NetworkError creates a network-related error
func NetworkError(endpoint string, err error) *AgentError {
	return Wrap(err, ErrorTypeNetwork, "Network request failed").
		WithContext("endpoint", endpoint).
		WithContext("suggestion", "Check network connectivity and firewall settings")
}

// AWSError creates an AWS-related error
func AWSError(service string, operation string, err error) *AgentError {
	e := Wrap(err, ErrorTypeAWS, fmt.Sprintf("AWS %s operation failed", service)).
		WithContext("service", service).
		WithContext("operation", operation)

	// Add specific suggestions based on error
	if strings.Contains(err.Error(), "credentials") {
		e.Suggestion = "Check AWS credentials: aws configure list"
	} else if strings.Contains(err.Error(), "throttl") {
		e.Suggestion = "Request throttled, implement exponential backoff"
		e.Retryable = true
	}

	return e
}

// LLMError creates an LLM-related error
func LLMError(errorType ErrorType, model string, err error) *AgentError {
	e := Wrap(err, errorType, "LLM request failed").
		WithContext("model", model)

	switch errorType {
	case ErrorTypeLLMTimeout:
		e.Suggestion = "Increase timeout or reduce prompt size"
	case ErrorTypeLLMRateLimit:
		e.Suggestion = "Implement rate limiting or use exponential backoff"
		e.Retryable = true
	case ErrorTypeLLMCost:
		e.Suggestion = "Consider using a smaller model or caching responses"
	}

	return e
}

// FunctionError creates a function execution error
func FunctionError(functionName string, err error) *AgentError {
	return Wrap(err, ErrorTypeFunctionExec, fmt.Sprintf("Function '%s' failed", functionName)).
		WithContext("function", functionName).
		WithContext("suggestion", "Check function parameters and permissions")
}

// ValidationError creates a validation error
func ValidationError(field string, value interface{}, reason string) *AgentError {
	return New(ErrorTypeValidation, fmt.Sprintf("Validation failed for '%s': %s", field, reason)).
		WithContext("field", field).
		WithContext("value", value).
		WithContext("suggestion", "Check input data against schema requirements")
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	OnError    func(context.Context, *AgentError)
	MaxRetries int
	RetryDelay time.Duration
}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		MaxRetries: 3,
		RetryDelay: time.Second,
		OnError: func(ctx context.Context, err *AgentError) {
			// Default logging
			fmt.Printf("[ERROR] %s\n", err.ToJSON())
		},
	}
}

// Handle processes an error with retry logic
func (h *ErrorHandler) Handle(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	agentErr, ok := err.(*AgentError)
	if !ok {
		agentErr = Wrap(err, ErrorTypeNetwork, err.Error())
	}

	// Call error callback
	if h.OnError != nil {
		h.OnError(ctx, agentErr)
	}

	return agentErr
}

// Retry executes a function with retry logic
func (h *ErrorHandler) Retry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= h.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(h.RetryDelay * time.Duration(attempt)):
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if agentErr, ok := err.(*AgentError); ok && !agentErr.Retryable {
			return err
		}
	}

	return Wrap(lastErr, ErrorTypeNetwork, "Max retries exceeded").
		WithContext("max_retries", h.MaxRetries)
}

// Helper functions

func getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func isRetryable(errorType ErrorType) bool {
	switch errorType {
	case ErrorTypeNetwork, ErrorTypeLLMRateLimit, ErrorTypeLLMTimeout:
		return true
	default:
		return false
	}
}

// IsRetryable checks if an error should be retried
func IsRetryable(err error) bool {
	if agentErr, ok := err.(*AgentError); ok {
		return agentErr.Retryable
	}
	return false
}

// GetType returns the error type
func GetType(err error) ErrorType {
	if agentErr, ok := err.(*AgentError); ok {
		return agentErr.Type
	}
	return ErrorTypeNetwork
}

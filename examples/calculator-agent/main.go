package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aygp-dr/go-agentic-workshop/pkg/bedrock"
	"github.com/aygp-dr/go-agentic-workshop/pkg/cost"
	"github.com/aygp-dr/go-agentic-workshop/pkg/errors"
	"github.com/aygp-dr/go-agentic-workshop/pkg/functions"
)

// CalculatorAgent demonstrates a simple tool-calling agent
type CalculatorAgent struct {
	llm      *bedrock.Client
	registry *functions.Registry
	tracker  *cost.CostTracker
}

func main() {
	fmt.Println("🤖 Calculator Agent Demo")
	fmt.Println("Type mathematical questions or 'quit' to exit")
	fmt.Println("Example: 'What is 25 * 4 + 10?'")
	fmt.Println()

	ctx := context.Background()
	
	// Initialize components
	agent, err := setupAgent(ctx)
	if err != nil {
		fmt.Printf("❌ Setup failed: %v\n", err)
		os.Exit(1)
	}

	// Set up cost budget
	agent.tracker.SetBudget("demo", &cost.Budget{
		Limit:        1.00, // $1 limit for demo
		Period:       24 * time.Hour,
		WarningLevel: 0.5,
	})

	// Interactive loop
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "quit" || input == "exit" {
			break
		}
		
		if input == "cost" {
			showCostReport(agent.tracker)
			continue
		}
		
		// Process the query
		response, err := agent.Process(ctx, input)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		
		fmt.Printf("🔢 %s\n\n", response)
	}
	
	fmt.Println("\n👋 Goodbye!")
	showCostReport(agent.tracker)
}

func setupAgent(ctx context.Context) (*CalculatorAgent, error) {
	// Initialize LLM client
	llmClient, err := bedrock.NewClient(ctx, "us-east-1")
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeAWS, "Failed to create Bedrock client")
	}
	
	// Initialize function registry
	registry := functions.NewRegistry()
	
	// Register calculator functions
	registerCalculatorFunctions(registry)
	
	// Initialize cost tracker
	tracker := cost.NewCostTracker()
	
	return &CalculatorAgent{
		llm:      llmClient,
		registry: registry,
		tracker:  tracker,
	}, nil
}

func registerCalculatorFunctions(registry *functions.Registry) {
	// Basic arithmetic
	registry.Register(&functions.Function{
		Name:        "add",
		Description: "Add two numbers",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"a": {"type": "number", "description": "First number"},
				"b": {"type": "number", "description": "Second number"}
			},
			"required": ["a", "b"]
		}`),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			a, _ := args["a"].(float64)
			b, _ := args["b"].(float64)
			return a + b, nil
		},
	})
	
	registry.Register(&functions.Function{
		Name:        "subtract",
		Description: "Subtract two numbers",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"a": {"type": "number", "description": "First number"},
				"b": {"type": "number", "description": "Second number"}
			},
			"required": ["a", "b"]
		}`),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			a, _ := args["a"].(float64)
			b, _ := args["b"].(float64)
			return a - b, nil
		},
	})
	
	registry.Register(&functions.Function{
		Name:        "multiply",
		Description: "Multiply two numbers",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"a": {"type": "number", "description": "First number"},
				"b": {"type": "number", "description": "Second number"}
			},
			"required": ["a", "b"]
		}`),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			a, _ := args["a"].(float64)
			b, _ := args["b"].(float64)
			return a * b, nil
		},
	})
	
	registry.Register(&functions.Function{
		Name:        "divide",
		Description: "Divide two numbers",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"a": {"type": "number", "description": "Dividend"},
				"b": {"type": "number", "description": "Divisor"}
			},
			"required": ["a", "b"]
		}`),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			a, _ := args["a"].(float64)
			b, _ := args["b"].(float64)
			if b == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return a / b, nil
		},
	})
}

// Process handles a user query
func (a *CalculatorAgent) Process(ctx context.Context, query string) (string, error) {
	// Build prompt with available functions
	prompt := a.buildPrompt(query)
	
	// Call LLM
	start := time.Now()
	response, err := a.llm.Invoke(ctx, &bedrock.InvokeRequest{
		ModelID: "anthropic.claude-3-haiku",
		Prompt:  prompt,
		MaxTokens: 500,
	})
	if err != nil {
		return "", errors.LLMError(errors.ErrorTypeLLMInvalid, "claude-3-haiku", err)
	}
	
	// Track usage
	usage := cost.Usage{
		ModelID:      "anthropic.claude-3-haiku",
		InputTokens:  len(prompt) / 4,  // Rough estimate
		OutputTokens: len(response.Content) / 4,
		WorkflowID:   "calculator-demo",
	}
	a.tracker.Track(ctx, usage)
	
	// Parse LLM response for function calls
	functionCalls, explanation := a.parseLLMResponse(response.Content)
	
	// Execute function calls
	results := make([]string, 0)
	for _, call := range functionCalls {
		result, err := a.executeFunction(ctx, call)
		if err != nil {
			return "", errors.FunctionError(call.Name, err)
		}
		results = append(results, fmt.Sprintf("%s = %v", call.Display, result))
	}
	
	// Format final response
	if len(results) > 0 {
		return fmt.Sprintf("%s\n\nCalculations:\n%s", 
			explanation, 
			strings.Join(results, "\n")), nil
	}
	
	return explanation, nil
}

// FunctionCall represents a parsed function call
type FunctionCall struct {
	Name     string
	Args     map[string]interface{}
	Display  string
}

func (a *CalculatorAgent) buildPrompt(query string) string {
	functionsJSON, _ := json.MarshalIndent(a.registry.GetManifest(), "", "  ")
	
	return fmt.Sprintf(`You are a helpful calculator assistant. 
You have access to the following mathematical functions:

%s

User query: %s

Please solve this step by step. For each calculation, specify:
1. The function to call
2. The arguments
3. A clear explanation

Respond in JSON format:
{
  "explanation": "Step by step explanation",
  "function_calls": [
    {
      "name": "function_name",
      "args": {"a": 1, "b": 2},
      "display": "1 + 2"
    }
  ]
}`, functionsJSON, query)
}

func (a *CalculatorAgent) parseLLMResponse(content string) ([]FunctionCall, string) {
	var response struct {
		Explanation   string         `json:"explanation"`
		FunctionCalls []FunctionCall `json:"function_calls"`
	}
	
	// Try to parse JSON response
	if err := json.Unmarshal([]byte(content), &response); err != nil {
		// Fallback to plain text
		return nil, content
	}
	
	return response.FunctionCalls, response.Explanation
}

func (a *CalculatorAgent) executeFunction(ctx context.Context, call FunctionCall) (interface{}, error) {
	return a.registry.Execute(ctx, call.Name, call.Args)
}

func showCostReport(tracker *cost.CostTracker) {
	costs := tracker.GetCurrentCosts()
	fmt.Println("\n💰 Cost Report:")
	fmt.Println("─────────────────")
	
	for model, cost := range costs {
		if model != "total" {
			fmt.Printf("%-30s: $%.4f\n", model, cost)
		}
	}
	
	fmt.Printf("%-30s: $%.4f\n", "TOTAL", costs["total"])
	
	// Show optimization suggestions
	recommendations := tracker.GetOptimizations()
	if len(recommendations) > 0 {
		fmt.Println("\n💡 Cost Optimization Tips:")
		for i, rec := range recommendations[:3] { // Show top 3
			fmt.Printf("%d. %s\n   %s\n", i+1, rec.Title, rec.Impact)
		}
	}
}
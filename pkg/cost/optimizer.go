package cost

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// OptimizationEngine analyzes usage patterns and suggests optimizations
type OptimizationEngine struct {
	patterns      []UsagePattern
	modelAnalysis map[string]*ModelAnalysis
}

// UsagePattern represents a detected usage pattern
type UsagePattern struct {
	Type        string
	Description string
	Impact      float64 // Potential cost reduction
	Frequency   int
}

// ModelAnalysis tracks model-specific metrics
type ModelAnalysis struct {
	ModelID            string
	AvgPromptLength    int
	AvgResponseLength  int
	RepetitivePrompts  int
	LongPrompts        int
	ShortResponses     int
	ErrorRate          float64
	AvgLatency         time.Duration
}

// Recommendation represents an optimization suggestion
type Recommendation struct {
	Priority    string  // high, medium, low
	Type        string  // model_switch, caching, batching, prompt_optimization
	Title       string
	Description string
	Impact      string  // Estimated savings
	Effort      string  // easy, medium, hard
	Example     string
}

// NewOptimizationEngine creates a new optimization engine
func NewOptimizationEngine() *OptimizationEngine {
	return &OptimizationEngine{
		patterns:      make([]UsagePattern, 0),
		modelAnalysis: make(map[string]*ModelAnalysis),
	}
}

// Update processes new usage data
func (e *OptimizationEngine) Update(usage Usage) {
	if _, ok := e.modelAnalysis[usage.ModelID]; !ok {
		e.modelAnalysis[usage.ModelID] = &ModelAnalysis{
			ModelID: usage.ModelID,
		}
	}
	
	analysis := e.modelAnalysis[usage.ModelID]
	
	// Update averages (simplified - in production use proper running averages)
	promptLength := usage.InputTokens * 4 // Approximate chars
	responseLength := usage.OutputTokens * 4
	
	analysis.AvgPromptLength = (analysis.AvgPromptLength + promptLength) / 2
	analysis.AvgResponseLength = (analysis.AvgResponseLength + responseLength) / 2
	
	// Detect patterns
	if promptLength > 8000 {
		analysis.LongPrompts++
	}
	if responseLength < 200 && promptLength > 1000 {
		analysis.ShortResponses++
	}
}

// GetRecommendations returns optimization recommendations
func (e *OptimizationEngine) GetRecommendations() []Recommendation {
	recommendations := make([]Recommendation, 0)
	
	// Analyze each model's usage
	for modelID, analysis := range e.modelAnalysis {
		// Check for expensive model with simple use case
		if strings.Contains(modelID, "claude-3-sonnet") && analysis.AvgResponseLength < 500 {
			recommendations = append(recommendations, Recommendation{
				Priority: "high",
				Type:     "model_switch",
				Title:    "Switch to Cheaper Model",
				Description: fmt.Sprintf(
					"Model %s is being used for short responses (avg %d chars). "+
					"Consider using Claude Haiku or Titan Express for 80%% cost reduction.",
					modelID, analysis.AvgResponseLength,
				),
				Impact: "Save $0.012 per 1K tokens",
				Effort: "easy",
				Example: `// Replace:
client.Call(ctx, prompt, WithModel("claude-3-sonnet"))

// With:
client.Call(ctx, prompt, WithModel("claude-3-haiku"))`,
			})
		}
		
		// Check for caching opportunities
		if analysis.RepetitivePrompts > 10 {
			recommendations = append(recommendations, Recommendation{
				Priority: "high",
				Type:     "caching",
				Title:    "Implement Response Caching",
				Description: fmt.Sprintf(
					"Detected %d repetitive prompts. Implement caching to eliminate redundant API calls.",
					analysis.RepetitivePrompts,
				),
				Impact: "Save 90% on repeated queries",
				Effort: "medium",
				Example: `// Add caching layer:
cache := NewLRUCache(1000)
if cached, ok := cache.Get(promptHash); ok {
    return cached, nil
}
response := client.Call(ctx, prompt)
cache.Set(promptHash, response, 1*time.Hour)`,
			})
		}
		
		// Check for prompt optimization
		if analysis.LongPrompts > 5 && analysis.ShortResponses > 5 {
			recommendations = append(recommendations, Recommendation{
				Priority: "medium",
				Type:     "prompt_optimization",
				Title:    "Optimize Prompt Length",
				Description: "Long prompts are generating short responses. Consider prompt compression or summarization.",
				Impact:  "Save 30-50% on input tokens",
				Effort:  "medium",
				Example: `// Use prompt templates:
template := "Summarize in 50 words: {{.Content}}"
// Instead of including full context every time`,
			})
		}
	}
	
	// Add general recommendations
	recommendations = append(recommendations, e.getGeneralRecommendations()...)
	
	// Sort by priority
	sort.Slice(recommendations, func(i, j int) bool {
		priority := map[string]int{"high": 3, "medium": 2, "low": 1}
		return priority[recommendations[i].Priority] > priority[recommendations[j].Priority]
	})
	
	return recommendations
}

// getGeneralRecommendations returns universally applicable optimizations
func (e *OptimizationEngine) getGeneralRecommendations() []Recommendation {
	return []Recommendation{
		{
			Priority: "medium",
			Type:     "batching",
			Title:    "Batch Similar Requests",
			Description: "Group similar prompts into batches to reduce overhead and improve throughput.",
			Impact:   "10-20% latency reduction",
			Effort:   "medium",
			Example: `// Batch process multiple items:
prompts := []string{prompt1, prompt2, prompt3}
responses := client.BatchCall(ctx, prompts)`,
		},
		{
			Priority: "low",
			Type:     "streaming",
			Title:    "Enable Response Streaming",
			Description: "Use streaming for long responses to improve perceived latency.",
			Impact:   "50% faster time-to-first-token",
			Effort:   "easy",
			Example: `// Enable streaming:
stream := client.CallStream(ctx, prompt)
for chunk := range stream {
    process(chunk)
}`,
		},
		{
			Priority: "medium",
			Type:     "local_model",
			Title:    "Use Local Models for Development",
			Description: "Run Ollama locally for development and testing to eliminate API costs.",
			Impact:   "100% cost reduction in dev",
			Effort:   "easy",
			Example: `// Development config:
if os.Getenv("ENV") == "development" {
    client = ollama.NewClient("llama3:8b")
}`,
		},
		{
			Priority: "high",
			Type:     "monitoring",
			Title:    "Implement Cost Alerts",
			Description: "Set up automated alerts for unusual spending patterns.",
			Impact:   "Prevent bill shock",
			Effort:   "easy",
			Example: `// Set daily budget:
tracker.SetBudget("daily", &Budget{
    Limit: 10.00,
    Period: 24 * time.Hour,
    WarningLevel: 0.8,
})`,
		},
	}
}

// CostOptimizer provides high-level optimization strategies
type CostOptimizer struct {
	tracker *CostTracker
}

// NewCostOptimizer creates a new cost optimizer
func NewCostOptimizer(tracker *CostTracker) *CostOptimizer {
	return &CostOptimizer{
		tracker: tracker,
	}
}

// SuggestModelForUseCase recommends the most cost-effective model
func (o *CostOptimizer) SuggestModelForUseCase(useCase string) ModelSuggestion {
	suggestions := map[string]ModelSuggestion{
		"simple_qa": {
			ModelID:     "amazon.titan-text-express-v1",
			Reason:      "Fast and cheap for simple questions",
			CostSavings: "90% cheaper than Claude 3",
		},
		"code_generation": {
			ModelID:     "anthropic.claude-3-sonnet",
			Reason:      "Best performance for complex code tasks",
			CostSavings: "Worth the cost for accuracy",
		},
		"summarization": {
			ModelID:     "anthropic.claude-3-haiku",
			Reason:      "Good balance of quality and cost",
			CostSavings: "80% cheaper than Sonnet",
		},
		"translation": {
			ModelID:     "meta.llama3-8b-instruct",
			Reason:      "Sufficient for most translation tasks",
			CostSavings: "85% cheaper than Claude",
		},
		"development": {
			ModelID:     "llama3:8b",
			Reason:      "Free local model for development",
			CostSavings: "100% cost reduction",
		},
	}
	
	if suggestion, ok := suggestions[useCase]; ok {
		return suggestion
	}
	
	return ModelSuggestion{
		ModelID:     "anthropic.claude-3-haiku",
		Reason:      "Good default balance",
		CostSavings: "Reasonable for most tasks",
	}
}

// ModelSuggestion contains model recommendation details
type ModelSuggestion struct {
	ModelID     string
	Reason      string
	CostSavings string
}

// EstimateMonthlyCost projects monthly costs based on usage
func (o *CostOptimizer) EstimateMonthlyCost(dailyRequests int, avgTokensPerRequest int, modelID string) CostEstimate {
	pricing, ok := ModelPricing[modelID]
	if !ok {
		return CostEstimate{Error: "Unknown model"}
	}
	
	// Assume 60/40 input/output split
	inputTokens := avgTokensPerRequest * 60 / 100
	outputTokens := avgTokensPerRequest * 40 / 100
	
	dailyCost := float64(dailyRequests) * (
		(float64(inputTokens)/1000)*pricing.InputPer1K +
		(float64(outputTokens)/1000)*pricing.OutputPer1K)
	
	monthlyCost := dailyCost * 30
	
	return CostEstimate{
		DailyCost:      dailyCost,
		MonthlyCost:    monthlyCost,
		YearlyCost:     monthlyCost * 12,
		BreakdownInput: monthlyCost * 0.3, // Input typically 30% of cost
		BreakdownOutput: monthlyCost * 0.7, // Output typically 70% of cost
	}
}

// CostEstimate contains cost projections
type CostEstimate struct {
	DailyCost       float64
	MonthlyCost     float64
	YearlyCost      float64
	BreakdownInput  float64
	BreakdownOutput float64
	Error           string
}
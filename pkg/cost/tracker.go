package cost

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Model pricing information (as of 2024)
var ModelPricing = map[string]ModelCost{
	// AWS Bedrock pricing
	"anthropic.claude-3-sonnet": {
		InputPer1K:  0.003,
		OutputPer1K: 0.015,
		Provider:    "bedrock",
	},
	"anthropic.claude-3-haiku": {
		InputPer1K:  0.00025,
		OutputPer1K: 0.00125,
		Provider:    "bedrock",
	},
	"anthropic.claude-instant-v1": {
		InputPer1K:  0.00163,
		OutputPer1K: 0.00551,
		Provider:    "bedrock",
	},
	"amazon.titan-text-express-v1": {
		InputPer1K:  0.0002,
		OutputPer1K: 0.0006,
		Provider:    "bedrock",
	},
	"meta.llama3-8b-instruct": {
		InputPer1K:  0.0003,
		OutputPer1K: 0.0006,
		Provider:    "bedrock",
	},
	// Ollama (local) - no cost but track for comparison
	"llama3:8b": {
		InputPer1K:  0,
		OutputPer1K: 0,
		Provider:    "ollama",
	},
	"mistral:7b": {
		InputPer1K:  0,
		OutputPer1K: 0,
		Provider:    "ollama",
	},
}

// ModelCost represents pricing for a model
type ModelCost struct {
	InputPer1K  float64 // Cost per 1K input tokens
	OutputPer1K float64 // Cost per 1K output tokens
	Provider    string  // Service provider
}

// Usage represents token usage for a request
type Usage struct {
	ModelID      string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Cost         float64
	Timestamp    time.Time
	RequestID    string
	WorkflowID   string
	Cached       bool
}

// CostTracker tracks LLM usage and costs
type CostTracker struct {
	mu           sync.RWMutex
	usage        []Usage
	budgets      map[string]*Budget
	alerts       []Alert
	optimization *OptimizationEngine
}

// Budget represents spending limits
type Budget struct {
	Name          string
	Limit         float64
	Period        time.Duration
	CurrentSpend  float64
	PeriodStart   time.Time
	WarningLevel  float64 // Percentage (0.8 = 80%)
	WorkflowLimit float64 // Per-workflow limit
}

// Alert represents a cost alert
type Alert struct {
	Type      string
	Message   string
	Timestamp time.Time
	Budget    string
	Value     float64
}

// NewCostTracker creates a new cost tracker
func NewCostTracker() *CostTracker {
	return &CostTracker{
		usage:        make([]Usage, 0),
		budgets:      make(map[string]*Budget),
		alerts:       make([]Alert, 0),
		optimization: NewOptimizationEngine(),
	}
}

// Track records usage and calculates cost
func (t *CostTracker) Track(ctx context.Context, usage Usage) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Calculate cost if not provided
	if usage.Cost == 0 {
		if pricing, ok := ModelPricing[usage.ModelID]; ok {
			usage.Cost = (float64(usage.InputTokens)/1000)*pricing.InputPer1K +
				(float64(usage.OutputTokens)/1000)*pricing.OutputPer1K
		}
	}

	usage.Timestamp = time.Now()
	t.usage = append(t.usage, usage)

	// Check budgets
	t.checkBudgets(usage)

	// Update optimization recommendations
	t.optimization.Update(usage)

	return nil
}

// SetBudget configures a spending budget
func (t *CostTracker) SetBudget(name string, budget *Budget) {
	t.mu.Lock()
	defer t.mu.Unlock()

	budget.Name = name
	budget.PeriodStart = time.Now()
	t.budgets[name] = budget
}

// GetCurrentCosts returns costs for current period
func (t *CostTracker) GetCurrentCosts() map[string]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	costs := make(map[string]float64)
	now := time.Now()

	// Aggregate by model
	for _, u := range t.usage {
		// Only count recent usage based on budget periods
		if now.Sub(u.Timestamp) < 24*time.Hour {
			costs[u.ModelID] += u.Cost
		}
	}

	costs["total"] = 0
	for _, cost := range costs {
		costs["total"] += cost
	}

	return costs
}

// GetUsageReport generates a detailed usage report
func (t *CostTracker) GetUsageReport(start, end time.Time) *UsageReport {
	t.mu.RLock()
	defer t.mu.RUnlock()

	report := &UsageReport{
		StartTime:     start,
		EndTime:       end,
		ModelUsage:    make(map[string]*ModelStats),
		WorkflowUsage: make(map[string]*WorkflowStats),
	}

	for _, u := range t.usage {
		if u.Timestamp.Before(start) || u.Timestamp.After(end) {
			continue
		}

		// Model stats
		if _, ok := report.ModelUsage[u.ModelID]; !ok {
			report.ModelUsage[u.ModelID] = &ModelStats{
				ModelID: u.ModelID,
			}
		}
		stats := report.ModelUsage[u.ModelID]
		stats.Requests++
		stats.InputTokens += u.InputTokens
		stats.OutputTokens += u.OutputTokens
		stats.TotalCost += u.Cost
		if u.Cached {
			stats.CacheHits++
		}

		// Workflow stats
		if u.WorkflowID != "" {
			if _, ok := report.WorkflowUsage[u.WorkflowID]; !ok {
				report.WorkflowUsage[u.WorkflowID] = &WorkflowStats{
					WorkflowID: u.WorkflowID,
				}
			}
			wstats := report.WorkflowUsage[u.WorkflowID]
			wstats.Requests++
			wstats.TotalCost += u.Cost
			wstats.TotalTokens += u.TotalTokens
		}

		report.TotalCost += u.Cost
		report.TotalRequests++
		report.TotalTokens += u.TotalTokens
	}

	// Calculate averages and rates
	for _, stats := range report.ModelUsage {
		if stats.Requests > 0 {
			stats.AvgInputTokens = stats.InputTokens / stats.Requests
			stats.AvgOutputTokens = stats.OutputTokens / stats.Requests
			stats.AvgCostPerRequest = stats.TotalCost / float64(stats.Requests)
			stats.CacheHitRate = float64(stats.CacheHits) / float64(stats.Requests)
		}
	}

	return report
}

// GetOptimizations returns cost optimization recommendations
func (t *CostTracker) GetOptimizations() []Recommendation {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.optimization.GetRecommendations()
}

// checkBudgets verifies spending against configured budgets
func (t *CostTracker) checkBudgets(usage Usage) {
	for name, budget := range t.budgets {
		// Check if budget period expired
		if time.Since(budget.PeriodStart) > budget.Period {
			budget.CurrentSpend = 0
			budget.PeriodStart = time.Now()
		}

		budget.CurrentSpend += usage.Cost

		// Check limits
		if budget.CurrentSpend > budget.Limit {
			t.alerts = append(t.alerts, Alert{
				Type:      "budget_exceeded",
				Message:   fmt.Sprintf("Budget '%s' exceeded: $%.2f > $%.2f", name, budget.CurrentSpend, budget.Limit),
				Timestamp: time.Now(),
				Budget:    name,
				Value:     budget.CurrentSpend,
			})
		} else if budget.CurrentSpend > budget.Limit*budget.WarningLevel {
			t.alerts = append(t.alerts, Alert{
				Type:      "budget_warning",
				Message:   fmt.Sprintf("Budget '%s' at %.0f%%: $%.2f of $%.2f", name, (budget.CurrentSpend/budget.Limit)*100, budget.CurrentSpend, budget.Limit),
				Timestamp: time.Now(),
				Budget:    name,
				Value:     budget.CurrentSpend,
			})
		}

		// Check per-workflow limit
		if usage.WorkflowID != "" && budget.WorkflowLimit > 0 {
			workflowCost := t.getWorkflowCost(usage.WorkflowID)
			if workflowCost > budget.WorkflowLimit {
				t.alerts = append(t.alerts, Alert{
					Type:      "workflow_limit_exceeded",
					Message:   fmt.Sprintf("Workflow '%s' exceeded limit: $%.2f > $%.2f", usage.WorkflowID, workflowCost, budget.WorkflowLimit),
					Timestamp: time.Now(),
					Budget:    name,
					Value:     workflowCost,
				})
			}
		}
	}
}

// getWorkflowCost calculates total cost for a workflow
func (t *CostTracker) getWorkflowCost(workflowID string) float64 {
	total := 0.0
	for _, u := range t.usage {
		if u.WorkflowID == workflowID {
			total += u.Cost
		}
	}
	return total
}

// GetAlerts returns recent alerts
func (t *CostTracker) GetAlerts(since time.Time) []Alert {
	t.mu.RLock()
	defer t.mu.RUnlock()

	alerts := make([]Alert, 0)
	for _, alert := range t.alerts {
		if alert.Timestamp.After(since) {
			alerts = append(alerts, alert)
		}
	}
	return alerts
}

// ExportMetrics exports metrics in Prometheus format
func (t *CostTracker) ExportMetrics() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	costs := t.GetCurrentCosts()
	metrics := "# HELP llm_cost_total Total cost by model\n"
	metrics += "# TYPE llm_cost_total counter\n"

	for model, cost := range costs {
		if model != "total" {
			metrics += fmt.Sprintf("llm_cost_total{model=\"%s\"} %.4f\n", model, cost)
		}
	}

	metrics += "\n# HELP llm_requests_total Total requests by model\n"
	metrics += "# TYPE llm_requests_total counter\n"

	requestCounts := make(map[string]int)
	for _, u := range t.usage {
		requestCounts[u.ModelID]++
	}

	for model, count := range requestCounts {
		metrics += fmt.Sprintf("llm_requests_total{model=\"%s\"} %d\n", model, count)
	}

	return metrics
}

// UsageReport contains detailed usage statistics
type UsageReport struct {
	StartTime     time.Time
	EndTime       time.Time
	TotalCost     float64
	TotalRequests int
	TotalTokens   int
	ModelUsage    map[string]*ModelStats
	WorkflowUsage map[string]*WorkflowStats
}

// ModelStats contains per-model statistics
type ModelStats struct {
	ModelID           string
	Requests          int
	InputTokens       int
	OutputTokens      int
	TotalCost         float64
	AvgInputTokens    int
	AvgOutputTokens   int
	AvgCostPerRequest float64
	CacheHits         int
	CacheHitRate      float64
}

// WorkflowStats contains per-workflow statistics
type WorkflowStats struct {
	WorkflowID  string
	Requests    int
	TotalCost   float64
	TotalTokens int
}

// ToJSON converts report to JSON
func (r *UsageReport) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	return string(data), err
}

package observability

import (
    "context"
    "time"
    
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/metric"
)

// Metrics holds all application metrics
type Metrics struct {
    // LLM metrics
    llmRequestDuration metric.Float64Histogram
    llmTokensUsed      metric.Int64Counter
    llmErrors          metric.Int64Counter
    
    // Workflow metrics
    workflowDuration   metric.Float64Histogram
    workflowSteps      metric.Int64Counter
    activeWorkflows    metric.Int64UpDownCounter
}

// NewMetrics creates a new metrics instance
func NewMetrics(meter metric.Meter) (*Metrics, error) {
    llmRequestDuration, err := meter.Float64Histogram(
        "llm.request.duration",
        metric.WithDescription("LLM request duration in seconds"),
        metric.WithUnit("s"),
    )
    if err != nil {
        return nil, err
    }
    
    llmTokensUsed, err := meter.Int64Counter(
        "llm.tokens.used",
        metric.WithDescription("Total tokens used"),
    )
    if err != nil {
        return nil, err
    }
    
    return &Metrics{
        llmRequestDuration: llmRequestDuration,
        llmTokensUsed:      llmTokensUsed,
    }, nil
}

// RecordLLMRequest records metrics for an LLM request
func (m *Metrics) RecordLLMRequest(ctx context.Context, model string, duration time.Duration, tokens int64) {
    m.llmRequestDuration.Record(ctx, duration.Seconds(),
        metric.WithAttributes(
            attribute.String("model", model),
        ),
    )
    
    m.llmTokensUsed.Add(ctx, tokens,
        metric.WithAttributes(
            attribute.String("model", model),
        ),
    )
}

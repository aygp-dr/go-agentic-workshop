package workflow

import (
    "context"
    "encoding/json"
    "time"
)

// WorkflowState represents the current state of a workflow
type WorkflowState struct {
    ID          string                 `json:"id"`
    Status      Status                 `json:"status"`
    CurrentStep string                 `json:"current_step"`
    Context     map[string]interface{} `json:"context"`
    History     []Step                 `json:"history"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// Status represents workflow status
type Status string

const (
    StatusPending   Status = "pending"
    StatusRunning   Status = "running"
    StatusCompleted Status = "completed"
    StatusFailed    Status = "failed"
)

// Step represents a workflow step
type Step struct {
    Name      string                 `json:"name"`
    Action    string                 `json:"action"`
    Input     map[string]interface{} `json:"input"`
    Output    map[string]interface{} `json:"output"`
    StartedAt time.Time              `json:"started_at"`
    EndedAt   time.Time              `json:"ended_at"`
}

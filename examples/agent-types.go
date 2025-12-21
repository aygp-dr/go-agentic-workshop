package examples

// AgentType represents different categories of AI agents
type AgentType string

const (
	// ReActAgent uses Reasoning and Acting pattern
	ReActAgent AgentType = "react"

	// PlanExecuteAgent creates plans then executes them
	PlanExecuteAgent AgentType = "plan-execute"

	// ToolCallingAgent focuses on function execution
	ToolCallingAgent AgentType = "tool-calling"

	// ConversationalAgent maintains dialogue context
	ConversationalAgent AgentType = "conversational"
)

// AgentCapabilities defines what an agent can do
type AgentCapabilities struct {
	CanPlan         bool
	CanExecuteTools bool
	HasMemory       bool
	SupportsAsync   bool
}

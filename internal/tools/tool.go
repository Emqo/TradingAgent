package tools

import "context"

// Tool defines the interface for agent tools.
type Tool interface {
	// Name returns the tool name.
	Name() string

	// Description returns a description of what the tool does.
	Description() string

	// Parameters returns the JSON schema for the tool's parameters.
	Parameters() map[string]any

	// Execute runs the tool with the given arguments.
	Execute(ctx context.Context, args map[string]any) (*Result, error)
}

// Result represents the result of a tool execution.
type Result struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NewSuccessResult creates a successful result.
func NewSuccessResult(data any) *Result {
	return &Result{Success: true, Data: data}
}

// NewErrorResult creates an error result.
func NewErrorResult(err string) *Result {
	return &Result{Success: false, Error: err}
}

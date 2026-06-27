package llm

import "context"

// Provider is the interface for all LLM providers.
type Provider interface {
	// Chat sends a conversation and returns a response.
	Chat(ctx context.Context, messages []Message, opts ...Option) (*Response, error)

	// ChatWithTools sends a conversation with tool definitions and returns a response.
	ChatWithTools(ctx context.Context, messages []Message, tools []Tool, opts ...Option) (*Response, error)

	// Name returns the provider name (e.g., "claude", "openai").
	Name() string
}

// Message represents a conversation message.
type Message struct {
	Role    string `json:"role"`    // "system", "user", "assistant"
	Content string `json:"content"`
}

// Tool represents a tool definition for function calling.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// Response represents an LLM response.
type Response struct {
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	TokenUsage TokenUsage `json:"token_usage"`
}

// ToolCall represents a tool call from the LLM.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// TokenUsage represents token consumption.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Option represents optional parameters for Chat calls.
type Option struct {
	MaxTokens   *int
	Temperature *float64
}

// WithMaxTokens sets the max tokens option.
func WithMaxTokens(n int) Option {
	return Option{MaxTokens: &n}
}

// WithTemperature sets the temperature option.
func WithTemperature(t float64) Option {
	return Option{Temperature: &t}
}

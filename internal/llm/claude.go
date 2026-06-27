package llm

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// ClaudeProvider implements the Provider interface for Claude/Anthropic API.
type ClaudeProvider struct {
	client anthropic.Client
	model  string
}

// NewClaudeProvider creates a new Claude provider with custom base URL and API key.
func NewClaudeProvider(baseURL, apiKey, model string) *ClaudeProvider {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}

	client := anthropic.NewClient(opts...)

	return &ClaudeProvider{
		client: client,
		model:  model,
	}
}

// Name returns the provider name.
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// Chat sends a conversation and returns a response.
func (p *ClaudeProvider) Chat(ctx context.Context, messages []Message, opts ...Option) (*Response, error) {
	return p.ChatWithTools(ctx, messages, nil, opts...)
}

// ChatWithTools sends a conversation with tool definitions and returns a response.
func (p *ClaudeProvider) ChatWithTools(ctx context.Context, messages []Message, tools []Tool, opts ...Option) (*Response, error) {
	// Build the request
	req := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: 4096,
	}

	// Apply options
	for _, opt := range opts {
		if opt.MaxTokens != nil {
			req.MaxTokens = int64(*opt.MaxTokens)
		}
	}

	// Convert messages
	var systemPrompt string
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			req.Messages = append(req.Messages, anthropic.MessageParam{
				Role: anthropic.MessageParamRoleUser,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewTextBlock(msg.Content),
				},
			})
		case "assistant":
			req.Messages = append(req.Messages, anthropic.MessageParam{
				Role: anthropic.MessageParamRoleAssistant,
				Content: []anthropic.ContentBlockParamUnion{
					anthropic.NewTextBlock(msg.Content),
				},
			})
		}
	}

	if systemPrompt != "" {
		req.System = []anthropic.TextBlockParam{
			{Text: systemPrompt},
		}
	}

	// Convert tools
	for _, tool := range tools {
		// Extract properties from the JSON schema
		var properties any
		if params, ok := tool.Parameters.(map[string]any); ok {
			if props, exists := params["properties"]; exists {
				properties = props
			}
		}

		req.Tools = append(req.Tools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: anthropic.ToolInputSchemaParam{
					Properties: properties,
				},
			},
		})
	}

	// Make the API call
	resp, err := p.client.Messages.New(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("claude api call: %w", err)
	}

	// Convert response
	result := &Response{
		TokenUsage: TokenUsage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
	}

	for _, block := range resp.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			result.Content += b.Text
		case anthropic.ToolUseBlock:
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:        b.ID,
				Name:      b.Name,
				Arguments: string(b.Input),
			})
		}
	}

	return result, nil
}

package tools

import (
	"context"
	"fmt"

	"github.com/Emqo/TradingAgent/internal/memory"
	"github.com/Emqo/TradingAgent/internal/strategy"
)

// GenerateStrategyTool generates a new trading strategy.
type GenerateStrategyTool struct {
	engine *strategy.Engine
}

// NewGenerateStrategyTool creates a new GenerateStrategyTool.
func NewGenerateStrategyTool(engine *strategy.Engine) *GenerateStrategyTool {
	return &GenerateStrategyTool{engine: engine}
}

// Name returns the tool name.
func (t *GenerateStrategyTool) Name() string {
	return "generate_strategy"
}

// Description returns the tool description.
func (t *GenerateStrategyTool) Description() string {
	return "Generate a new trading strategy based on current market conditions."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GenerateStrategyTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"market_context": map[string]any{
				"type":        "string",
				"description": "Current market context and conditions",
			},
		},
		"required": []string{"market_context"},
	}
}

// Execute runs the tool.
func (t *GenerateStrategyTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	marketContext, _ := args["market_context"].(string)
	if marketContext == "" {
		return NewErrorResult("market_context is required"), nil
	}

	strat, err := t.engine.GenerateStrategy(ctx, marketContext)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to generate strategy: %v", err)), nil
	}

	return NewSuccessResult(map[string]any{
		"strategy_id": strat.ID,
		"name":        strat.Name,
		"description": strat.Description,
		"config":      strat.Config,
		"expires_at":  strat.ExpiresAt,
	}), nil
}

// GetStrategyStatusTool returns the current strategy status.
type GetStrategyStatusTool struct {
	engine *strategy.Engine
}

// NewGetStrategyStatusTool creates a new GetStrategyStatusTool.
func NewGetStrategyStatusTool(engine *strategy.Engine) *GetStrategyStatusTool {
	return &GetStrategyStatusTool{engine: engine}
}

// Name returns the tool name.
func (t *GetStrategyStatusTool) Name() string {
	return "get_strategy_status"
}

// Description returns the tool description.
func (t *GetStrategyStatusTool) Description() string {
	return "Get the current active strategy and its status."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GetStrategyStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

// Execute runs the tool.
func (t *GetStrategyStatusTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	strat := t.engine.GetActiveStrategy()
	if strat == nil {
		return NewSuccessResult(map[string]any{
			"active": false,
			"message": "No active strategy",
		}), nil
	}

	return NewSuccessResult(map[string]any{
		"active":      true,
		"strategy_id": strat.ID,
		"name":        strat.Name,
		"description": strat.Description,
		"config":      strat.Config,
		"expires_at":  strat.ExpiresAt,
		"is_expired":  t.engine.IsStrategyExpired(),
		"performance": strat.Performance,
	}), nil
}

// AddMemoryTool adds an entry to memory.
type AddMemoryTool struct {
	memory *memory.Manager
}

// NewAddMemoryTool creates a new AddMemoryTool.
func NewAddMemoryTool(mem *memory.Manager) *AddMemoryTool {
	return &AddMemoryTool{memory: mem}
}

// Name returns the tool name.
func (t *AddMemoryTool) Name() string {
	return "add_memory"
}

// Description returns the tool description.
func (t *AddMemoryTool) Description() string {
	return "Add an entry to the agent's memory for future reference."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *AddMemoryTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"type": map[string]any{
				"type":        "string",
				"description": "Memory type: DECISION, OBSERVATION, ANALYSIS, TRADE, RISK, STRATEGY, REFLECTION",
				"enum":        []string{"DECISION", "OBSERVATION", "ANALYSIS", "TRADE", "RISK", "STRATEGY", "REFLECTION"},
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The memory content",
			},
			"tags": map[string]any{
				"type":        "array",
				"description": "Tags for categorization",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"type", "content"},
	}
}

// Execute runs the tool.
func (t *AddMemoryTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	memType, _ := args["type"].(string)
	content, _ := args["content"].(string)

	if memType == "" || content == "" {
		return NewErrorResult("type and content are required"), nil
	}

	// Parse tags
	var tags []string
	if tagsRaw, ok := args["tags"].([]any); ok {
		for _, tag := range tagsRaw {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	entry := memory.MemoryEntry{
		Type:    memory.MemoryType(memType),
		Content: content,
		Tags:    tags,
	}

	// Add to short-term by default
	t.memory.AddToShortTerm(entry)

	return NewSuccessResult(map[string]any{
		"message": "Memory added successfully",
		"type":    memType,
	}), nil
}

// GetMemoryContextTool returns the current memory context.
type GetMemoryContextTool struct {
	memory *memory.Manager
}

// NewGetMemoryContextTool creates a new GetMemoryContextTool.
func NewGetMemoryContextTool(mem *memory.Manager) *GetMemoryContextTool {
	return &GetMemoryContextTool{memory: mem}
}

// Name returns the tool name.
func (t *GetMemoryContextTool) Name() string {
	return "get_memory_context"
}

// Description returns the tool description.
func (t *GetMemoryContextTool) Description() string {
	return "Get the current memory context including recent decisions and working memory."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GetMemoryContextTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

// Execute runs the tool.
func (t *GetMemoryContextTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	context := t.memory.GetContext()
	stats := t.memory.GetStats()

	return NewSuccessResult(map[string]any{
		"context": context,
		"stats":   stats,
	}), nil
}

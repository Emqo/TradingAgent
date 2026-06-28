package tools

import (
	"context"
	"fmt"

	"github.com/Emqo/TradingAgent/internal/arbitrage"
	"github.com/Emqo/TradingAgent/internal/exchange"
)

// ExecuteArbitrageTool executes an arbitrage trade.
type ExecuteArbitrageTool struct {
	exchange exchange.Exchange
	manager  *arbitrage.Manager
}

// NewExecuteArbitrageTool creates a new ExecuteArbitrageTool.
func NewExecuteArbitrageTool(exchange exchange.Exchange, manager *arbitrage.Manager) *ExecuteArbitrageTool {
	return &ExecuteArbitrageTool{
		exchange: exchange,
		manager:  manager,
	}
}

// Name returns the tool name.
func (t *ExecuteArbitrageTool) Name() string {
	return "execute_arbitrage"
}

// Description returns the tool description.
func (t *ExecuteArbitrageTool) Description() string {
	return "Execute an arbitrage trade. Use this after analyzing an arbitrage opportunity and deciding to execute."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *ExecuteArbitrageTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol": map[string]any{
				"type":        "string",
				"description": "Primary trading pair (e.g., BTCUSDT)",
			},
			"amount_usdt": map[string]any{
				"type":        "number",
				"description": "Amount in USDT to use for arbitrage",
			},
		},
		"required": []string{"symbol", "amount_usdt"},
	}
}

// Execute runs the tool.
func (t *ExecuteArbitrageTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	symbol, _ := args["symbol"].(string)
	amountUSDT, _ := args["amount_usdt"].(float64)

	if symbol == "" || amountUSDT <= 0 {
		return NewErrorResult("symbol and amount_usdt are required"), nil
	}

	// TODO: Execute the arbitrage trade
	// This would involve:
	// 1. Calculate exact quantities for each leg
	// 2. Execute trades in sequence
	// 3. Handle partial fills
	// 4. Calculate actual profit

	return NewSuccessResult(map[string]any{
		"message":     "套利执行功能尚未实现",
		"symbol":      symbol,
		"amount_usdt": amountUSDT,
	}), nil
}

// GetArbitrageOpportunitiesTool returns current arbitrage opportunities.
type GetArbitrageOpportunitiesTool struct {
	manager *arbitrage.Manager
}

// NewGetArbitrageOpportunitiesTool creates a new GetArbitrageOpportunitiesTool.
func NewGetArbitrageOpportunitiesTool(manager *arbitrage.Manager) *GetArbitrageOpportunitiesTool {
	return &GetArbitrageOpportunitiesTool{manager: manager}
}

// Name returns the tool name.
func (t *GetArbitrageOpportunitiesTool) Name() string {
	return "get_arbitrage_opportunities"
}

// Description returns the tool description.
func (t *GetArbitrageOpportunitiesTool) Description() string {
	return "Get current arbitrage opportunities detected by the system."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GetArbitrageOpportunitiesTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

// Execute runs the tool.
func (t *GetArbitrageOpportunitiesTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	// Scan for opportunities
	result, err := t.manager.Scan(ctx)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("扫描套利机会失败: %v", err)), nil
	}

	totalOpportunities := len(result.TriangularOpportunities) + len(result.CashAndCarryOpportunities)

	if totalOpportunities == 0 {
		return NewSuccessResult(map[string]any{
			"message":       "暂无套利机会",
			"opportunities": []any{},
		}), nil
	}

	// Format opportunities
	type FormattedOpportunity struct {
		Type      string  `json:"type"`
		Path      string  `json:"path"`
		SpreadBps float64 `json:"spread_bps"`
		Profit    float64 `json:"profit"`
	}

	var formatted []FormattedOpportunity

	for _, opp := range result.TriangularOpportunities {
		formatted = append(formatted, FormattedOpportunity{
			Type:      "三角套利",
			Path:      opp.Path.Name,
			SpreadBps: opp.Spread,
			Profit:    opp.Profit,
		})
	}

	for _, opp := range result.CashAndCarryOpportunities {
		formatted = append(formatted, FormattedOpportunity{
			Type:      "期现套利",
			Path:      opp.Symbol,
			SpreadBps: opp.BasisPercent * 10000,
			Profit:    0,
		})
	}

	return NewSuccessResult(map[string]any{
		"message":       fmt.Sprintf("发现 %d 个套利机会", totalOpportunities),
		"opportunities": formatted,
	}), nil
}

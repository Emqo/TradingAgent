package tools

import (
	"context"
	"fmt"

	"github.com/Emqo/TradingAgent/internal/exchange"
)

// GetBalanceTool returns the account balance.
type GetBalanceTool struct {
	exchange exchange.Exchange
}

// NewGetBalanceTool creates a new GetBalanceTool.
func NewGetBalanceTool(exchange exchange.Exchange) *GetBalanceTool {
	return &GetBalanceTool{exchange: exchange}
}

// Name returns the tool name.
func (t *GetBalanceTool) Name() string {
	return "get_balance"
}

// Description returns the tool description.
func (t *GetBalanceTool) Description() string {
	return "Get the account balance. Returns all assets with free and locked amounts."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GetBalanceTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

// Execute runs the tool.
func (t *GetBalanceTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	balances, err := t.exchange.GetBalance(ctx)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to get balance: %v", err)), nil
	}

	// Convert to a list for better readability
	type BalanceInfo struct {
		Asset  string  `json:"asset"`
		Free   float64 `json:"free"`
		Locked float64 `json:"locked"`
		Total  float64 `json:"total"`
	}

	var balanceList []BalanceInfo
	for _, b := range balances {
		balanceList = append(balanceList, BalanceInfo{
			Asset:  b.Asset,
			Free:   b.Free,
			Locked: b.Locked,
			Total:  b.Total,
		})
	}

	return NewSuccessResult(map[string]any{
		"balances": balanceList,
		"count":    len(balanceList),
	}), nil
}

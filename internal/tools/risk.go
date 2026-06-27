package tools

import (
	"context"
	"fmt"

	"github.com/Emqo/TradingAgent/internal/risk"
)

// CheckRiskTool checks if a trade is allowed by the risk manager.
type CheckRiskTool struct {
	riskManager *risk.Manager
}

// NewCheckRiskTool creates a new CheckRiskTool.
func NewCheckRiskTool(riskManager *risk.Manager) *CheckRiskTool {
	return &CheckRiskTool{riskManager: riskManager}
}

// Name returns the tool name.
func (t *CheckRiskTool) Name() string {
	return "check_risk"
}

// Description returns the tool description.
func (t *CheckRiskTool) Description() string {
	return "Check if a trade is allowed by the risk management system. Use this before placing any trade."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *CheckRiskTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol": map[string]any{
				"type":        "string",
				"description": "Trading pair symbol (e.g., BTCUSDT)",
			},
			"side": map[string]any{
				"type":        "string",
				"description": "Trade side: BUY or SELL",
				"enum":        []string{"BUY", "SELL"},
			},
			"size_usdt": map[string]any{
				"type":        "number",
				"description": "Trade size in USDT",
			},
			"leverage": map[string]any{
				"type":        "number",
				"description": "Leverage multiplier (default: 1.0)",
			},
		},
		"required": []string{"symbol", "side", "size_usdt"},
	}
}

// Execute runs the tool.
func (t *CheckRiskTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	symbol, _ := args["symbol"].(string)
	side, _ := args["side"].(string)
	sizeUSDT, _ := args["size_usdt"].(float64)
	leverage := 1.0
	if l, ok := args["leverage"].(float64); ok {
		leverage = l
	}

	if symbol == "" || side == "" || sizeUSDT == 0 {
		return NewErrorResult("symbol, side, and size_usdt are required"), nil
	}

	result, err := t.riskManager.CheckTrade(symbol, side, sizeUSDT, leverage)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("risk check failed: %v", err)), nil
	}

	return NewSuccessResult(map[string]any{
		"allowed": result.Allowed,
		"reason":  result.Reason,
		"checks":  result.Checks,
	}), nil
}

// GetRiskStatusTool returns the current risk status.
type GetRiskStatusTool struct {
	riskManager *risk.Manager
}

// NewGetRiskStatusTool creates a new GetRiskStatusTool.
func NewGetRiskStatusTool(riskManager *risk.Manager) *GetRiskStatusTool {
	return &GetRiskStatusTool{riskManager: riskManager}
}

// Name returns the tool name.
func (t *GetRiskStatusTool) Name() string {
	return "get_risk_status"
}

// Description returns the tool description.
func (t *GetRiskStatusTool) Description() string {
	return "Get the current risk management status including positions, PnL, and alerts."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GetRiskStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

// Execute runs the tool.
func (t *GetRiskStatusTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	state := t.riskManager.GetState()
	recentAlerts := t.riskManager.GetAlerts(10)

	return NewSuccessResult(map[string]any{
		"total_position_usdt": state.TotalPositionUSDT,
		"daily_pnl":           state.DailyPnL,
		"daily_pnl_percent":   state.DailyPnLPercent,
		"drawdown_percent":    state.DrawdownPercent,
		"peak_value":          state.PeakValue,
		"current_value":       state.CurrentValue,
		"is_paused":           state.IsPaused,
		"pause_reason":        state.PauseReason,
		"positions":           state.Positions,
		"recent_alerts":       recentAlerts,
	}), nil
}

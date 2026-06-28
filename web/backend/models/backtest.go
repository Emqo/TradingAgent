package models

import "time"

// Backtest represents a backtest configuration and result.
type Backtest struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Strategy    string    `json:"strategy"`
	Symbol      string    `json:"symbol"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	InitialUSDT float64   `json:"initial_usdt"`
	Status      string    `json:"status"` // "pending", "running", "completed", "failed"
	Result      *BacktestResult `json:"result,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BacktestResult represents the result of a backtest.
type BacktestResult struct {
	TotalReturn    float64 `json:"total_return"`
	TotalReturnPct float64 `json:"total_return_pct"`
	SharpeRatio    float64 `json:"sharpe_ratio"`
	MaxDrawdown    float64 `json:"max_drawdown"`
	MaxDrawdownPct float64 `json:"max_drawdown_pct"`
	WinRate        float64 `json:"win_rate"`
	TotalTrades    int     `json:"total_trades"`
	WinningTrades  int     `json:"winning_trades"`
	LosingTrades   int     `json:"losing_trades"`
	ProfitFactor   float64 `json:"profit_factor"`
	AvgTradePnL    float64 `json:"avg_trade_pnl"`
	LLMCost        float64 `json:"llm_cost"`
}

// CreateBacktestRequest represents a request to create a backtest.
type CreateBacktestRequest struct {
	Name        string  `json:"name" binding:"required"`
	Strategy    string  `json:"strategy" binding:"required"`
	Symbol      string  `json:"symbol" binding:"required"`
	StartTime   string  `json:"start_time" binding:"required"`
	EndTime     string  `json:"end_time" binding:"required"`
	InitialUSDT float64 `json:"initial_usdt" binding:"required,gt=0"`
}

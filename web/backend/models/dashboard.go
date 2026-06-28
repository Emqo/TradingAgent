package models

// DashboardStats represents the dashboard overview statistics.
type DashboardStats struct {
	TotalBalance    float64            `json:"total_balance"`
	DailyPnL        float64            `json:"daily_pnl"`
	DailyPnLPct     float64            `json:"daily_pnl_pct"`
	TotalPnL        float64            `json:"total_pnl"`
	TotalPnLPct     float64            `json:"total_pnl_pct"`
	OpenPositions   int                `json:"open_positions"`
	TodayTrades     int                `json:"today_trades"`
	WinRate         float64            `json:"win_rate"`
	Positions       []PositionInfo     `json:"positions"`
	RecentTrades    []TradeInfo        `json:"recent_trades"`
	ArbitrageStats  ArbitrageStats     `json:"arbitrage_stats"`
	RiskStatus      RiskStatus         `json:"risk_status"`
}

// PositionInfo represents a position for display.
type PositionInfo struct {
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`
	Size        float64 `json:"size"`
	EntryPrice  float64 `json:"entry_price"`
	CurrentPrice float64 `json:"current_price"`
	PnL         float64 `json:"pnl"`
	PnLPct      float64 `json:"pnl_pct"`
}

// TradeInfo represents a trade for display.
type TradeInfo struct {
	ID        int64   `json:"id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Type      string  `json:"type"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	PnL       float64 `json:"pnl"`
	Timestamp string  `json:"timestamp"`
}

// ArbitrageStats represents arbitrage statistics.
type ArbitrageStats struct {
	TriangularOpportunities   int     `json:"triangular_opportunities"`
	CashAndCarryOpportunities int     `json:"cash_and_carry_opportunities"`
	TotalProfit               float64 `json:"total_profit"`
	AvgSpreadBps              float64 `json:"avg_spread_bps"`
}

// RiskStatus represents the current risk status.
type RiskStatus struct {
	IsPaused        bool    `json:"is_paused"`
	PauseReason     string  `json:"pause_reason,omitempty"`
	DrawdownPct     float64 `json:"drawdown_pct"`
	DailyLossPct    float64 `json:"daily_loss_pct"`
	PositionsCount  int     `json:"positions_count"`
}

package handlers

import (
	"net/http"

	"github.com/Emqo/TradingAgent/internal/agent"
	"github.com/Emqo/TradingAgent/internal/arbitrage"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/risk"
	"github.com/Emqo/TradingAgent/web/backend/models"
	"github.com/gin-gonic/gin"
)

// DashboardHandler handles dashboard requests.
type DashboardHandler struct {
	exchange  exchange.Exchange
	riskMgr   *risk.Manager
	arbMgr    *arbitrage.Manager
	agent     *agent.Agent
}

// NewDashboardHandler creates a new dashboard handler.
func NewDashboardHandler(
	exchange exchange.Exchange,
	riskMgr *risk.Manager,
	arbMgr *arbitrage.Manager,
	agent *agent.Agent,
) *DashboardHandler {
	return &DashboardHandler{
		exchange: exchange,
		riskMgr:  riskMgr,
		arbMgr:   arbMgr,
		agent:    agent,
	}
}

// GetStats returns dashboard statistics.
func (h *DashboardHandler) GetStats(c *gin.Context) {
	// Get risk status
	riskState := h.riskMgr.GetState()

	// Get balance
	ctx := c.Request.Context()
	balances, err := h.exchange.GetBalance(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balance"})
		return
	}

	// Calculate total balance
	totalBalance := 0.0
	for _, b := range balances {
		totalBalance += b.Total
	}

	// Build response
	stats := models.DashboardStats{
		TotalBalance:  totalBalance,
		DailyPnL:      riskState.DailyPnL,
		DailyPnLPct:   riskState.DailyPnLPercent,
		TotalPnL:      0, // TODO: Calculate from trade history
		TotalPnLPct:   0, // TODO: Calculate from trade history
		OpenPositions: len(riskState.Positions),
		TodayTrades:   0, // TODO: Count from trade history
		WinRate:       0, // TODO: Calculate from trade history
		RiskStatus: models.RiskStatus{
			IsPaused:       riskState.IsPaused,
			PauseReason:    riskState.PauseReason,
			DrawdownPct:    riskState.DrawdownPercent,
			DailyLossPct:   riskState.DailyPnLPercent,
			PositionsCount: len(riskState.Positions),
		},
	}

	// Add positions
	for _, pos := range riskState.Positions {
		stats.Positions = append(stats.Positions, models.PositionInfo{
			Symbol:       pos.Symbol,
			Side:         pos.Side,
			Size:         pos.Size,
			EntryPrice:   pos.EntryPrice,
			CurrentPrice: pos.CurrentPrice,
			PnL:          pos.PnL,
			PnLPct:       pos.PnLPercent,
		})
	}

	c.JSON(http.StatusOK, stats)
}

// GetPositions returns current positions.
func (h *DashboardHandler) GetPositions(c *gin.Context) {
	riskState := h.riskMgr.GetState()

	var positions []models.PositionInfo
	for _, pos := range riskState.Positions {
		positions = append(positions, models.PositionInfo{
			Symbol:       pos.Symbol,
			Side:         pos.Side,
			Size:         pos.Size,
			EntryPrice:   pos.EntryPrice,
			CurrentPrice: pos.CurrentPrice,
			PnL:          pos.PnL,
			PnLPct:       pos.PnLPercent,
		})
	}

	c.JSON(http.StatusOK, positions)
}

// GetBalance returns account balance.
func (h *DashboardHandler) GetBalance(c *gin.Context) {
	ctx := c.Request.Context()
	balances, err := h.exchange.GetBalance(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balance"})
		return
	}

	c.JSON(http.StatusOK, balances)
}

// GetRiskStatus returns risk status.
func (h *DashboardHandler) GetRiskStatus(c *gin.Context) {
	riskState := h.riskMgr.GetState()

	c.JSON(http.StatusOK, models.RiskStatus{
		IsPaused:       riskState.IsPaused,
		PauseReason:    riskState.PauseReason,
		DrawdownPct:    riskState.DrawdownPercent,
		DailyLossPct:   riskState.DailyPnLPercent,
		PositionsCount: len(riskState.Positions),
	})
}

// PauseTrading pauses trading.
func (h *DashboardHandler) PauseTrading(c *gin.Context) {
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = "Paused by user"
	}

	h.riskMgr.PauseTrading(req.Reason)
	c.JSON(http.StatusOK, gin.H{"message": "Trading paused"})
}

// ResumeTrading resumes trading.
func (h *DashboardHandler) ResumeTrading(c *gin.Context) {
	h.riskMgr.ResumeTrading()
	c.JSON(http.StatusOK, gin.H{"message": "Trading resumed"})
}

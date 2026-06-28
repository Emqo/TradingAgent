package handlers

import (
	"net/http"

	"github.com/Emqo/TradingAgent/internal/arbitrage"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/gin-gonic/gin"
)

// ArbitrageHandler handles arbitrage-related requests.
type ArbitrageHandler struct {
	exchange  exchange.Exchange
	arbMgr    *arbitrage.Manager
}

// NewArbitrageHandler creates a new arbitrage handler.
func NewArbitrageHandler(exchange exchange.Exchange, arbMgr *arbitrage.Manager) *ArbitrageHandler {
	return &ArbitrageHandler{
		exchange: exchange,
		arbMgr:   arbMgr,
	}
}

// GetOpportunities returns current arbitrage opportunities.
func (h *ArbitrageHandler) GetOpportunities(c *gin.Context) {
	ctx := c.Request.Context()

	// Scan for opportunities
	result, err := h.arbMgr.Scan(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan arbitrage opportunities"})
		return
	}

	// Format triangular opportunities
	type Opportunity struct {
		Type      string  `json:"type"`
		Path      string  `json:"path"`
		SpreadBps float64 `json:"spread_bps"`
		ProfitUSDT float64 `json:"profit_usdt"`
		Timestamp string  `json:"timestamp"`
	}

	var opportunities []Opportunity

	// Add triangular opportunities
	for _, opp := range result.TriangularOpportunities {
		opportunities = append(opportunities, Opportunity{
			Type:       "三角套利",
			Path:       opp.Path.Name,
			SpreadBps:  opp.Spread,
			ProfitUSDT: opp.Profit,
			Timestamp:  opp.Timestamp.Format("2006-01-02T15:04:05-07:00"),
		})
	}

	// Add cash-and-carry opportunities
	for _, opp := range result.CashAndCarryOpportunities {
		opportunities = append(opportunities, Opportunity{
			Type:       "期现套利",
			Path:       opp.Symbol + " 永续合约",
			SpreadBps:  opp.BasisPercent * 10000, // Convert to bps
			ProfitUSDT: 0, // Cash-and-carry profit depends on position size
			Timestamp:  opp.Timestamp.Format("2006-01-02T15:04:05-07:00"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"opportunities": opportunities,
	})
}

// GetStats returns arbitrage statistics.
func (h *ArbitrageHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Scan for opportunities to get current stats
	result, err := h.arbMgr.Scan(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get arbitrage stats"})
		return
	}

	// Calculate stats
	totalOpportunities := len(result.TriangularOpportunities) + len(result.CashAndCarryOpportunities)
	totalProfit := 0.0
	totalSpread := 0.0

	for _, opp := range result.TriangularOpportunities {
		totalProfit += opp.Profit
		totalSpread += opp.Spread
	}

	avgSpread := 0.0
	if len(result.TriangularOpportunities) > 0 {
		avgSpread = totalSpread / float64(len(result.TriangularOpportunities))
	}

	c.JSON(http.StatusOK, gin.H{
		"total_opportunities": totalOpportunities,
		"total_profit":       totalProfit,
		"avg_spread":         avgSpread,
		"success_rate":       85, // TODO: Calculate from trade history
	})
}

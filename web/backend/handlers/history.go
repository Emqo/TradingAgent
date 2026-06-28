package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/Emqo/TradingAgent/internal/database"
	"github.com/gin-gonic/gin"
)

// HistoryHandler handles historical data requests.
type HistoryHandler struct {
	db *database.DB
}

// NewHistoryHandler creates a new history handler.
func NewHistoryHandler(db *database.DB) *HistoryHandler {
	return &HistoryHandler{db: db}
}

// GetEquityCurve returns the equity curve data for charts.
func (h *HistoryHandler) GetEquityCurve(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	// Get query parameters
	days := 30
	if d := c.Query("days"); d != "" {
		// Parse days parameter
		if parsed, err := time.ParseDuration(d + "h"); err == nil {
			days = int(parsed.Hours() / 24)
		}
	}

	ctx := c.Request.Context()
	from := time.Now().AddDate(0, 0, -days)
	to := time.Now()

	snapshots, err := h.db.GetEquitySnapshots(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get equity curve"})
		return
	}

	// Format for chart
	type DataPoint struct {
		Timestamp string  `json:"timestamp"`
		Value     float64 `json:"value"`
	}

	data := make([]DataPoint, len(snapshots))
	for i, s := range snapshots {
		data[i] = DataPoint{
			Timestamp: s.Timestamp.Format("2006-01-02T15:04:05"),
			Value:     s.TotalValue,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

// GetTradeHistory returns trade history.
func (h *HistoryHandler) GetTradeHistory(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusOK, gin.H{"trades": []interface{}{}})
		return
	}

	// Get query parameters
	limit := 100
	strategy := c.Query("strategy")

	ctx := c.Request.Context()
	trades, err := h.db.GetTrades(ctx, limit, strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trade history"})
		return
	}

	// Format for response
	type TradeResponse struct {
		ID        int64   `json:"id"`
		Time      string  `json:"time"`
		Symbol    string  `json:"symbol"`
		Side      string  `json:"side"`
		Type      string  `json:"type"`
		Price     float64 `json:"price"`
		Quantity  float64 `json:"quantity"`
		Total     float64 `json:"total"`
		PnL       float64 `json:"pnl"`
		Strategy  string  `json:"strategy"`
		Status    string  `json:"status"`
	}

	response := make([]TradeResponse, len(trades))
	for i, t := range trades {
		response[i] = TradeResponse{
			ID:       t.ID,
			Time:     t.CreatedAt.Format("2006-01-02T15:04:05"),
			Symbol:   t.Symbol,
			Side:     t.Side,
			Type:     t.Type,
			Price:    t.Price,
			Quantity: t.Quantity,
			Total:    t.Total,
			PnL:      t.PnL,
			Strategy: t.Strategy,
			Status:   t.Status,
		}
	}

	c.JSON(http.StatusOK, gin.H{"trades": response})
}

// GetDailyStats returns daily statistics.
func (h *HistoryHandler) GetDailyStats(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusOK, gin.H{"stats": []interface{}{}})
		return
	}

	// Get query parameters
	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := time.ParseDuration(d + "h"); err == nil {
			days = int(parsed.Hours() / 24)
		}
	}

	ctx := c.Request.Context()
	from := time.Now().AddDate(0, 0, -days)
	to := time.Now()

	stats, err := h.db.GetDailyStats(ctx, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get daily stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// RecordEquitySnapshot records the current equity value.
func (h *HistoryHandler) RecordEquitySnapshot(totalValue float64) error {
	if h.db == nil {
		return nil
	}

	ctx := context.Background()
	snapshot := &database.EquitySnapshot{
		Timestamp:  time.Now(),
		TotalValue: totalValue,
	}

	return h.db.InsertEquitySnapshot(ctx, snapshot)
}

// RecordTrade records a trade.
func (h *HistoryHandler) RecordTrade(trade *database.Trade) error {
	if h.db == nil {
		return nil
	}

	ctx := context.Background()
	return h.db.InsertTrade(ctx, trade)
}

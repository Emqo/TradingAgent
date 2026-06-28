package handlers

import (
	"net/http"
	"time"

	"github.com/Emqo/TradingAgent/internal/backtest"
	"github.com/Emqo/TradingAgent/internal/database"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/gin-gonic/gin"
)

// BacktestHandler handles backtest requests.
type BacktestHandler struct {
	engine *backtest.Engine
	db     *database.DB
}

// NewBacktestHandler creates a new backtest handler.
func NewBacktestHandler(exchange exchange.Exchange, db *database.DB) *BacktestHandler {
	engine := backtest.NewEngine(exchange, db)
	return &BacktestHandler{
		engine: engine,
		db:     db,
	}
}

// RunBacktest runs a backtest with the given configuration.
func (h *BacktestHandler) RunBacktest(c *gin.Context) {
	var request struct {
		Strategy    string `json:"strategy" binding:"required"`
		Symbol      string `json:"symbol" binding:"required"`
		StartTime   string `json:"start_time" binding:"required"`
		EndTime     string `json:"end_time" binding:"required"`
		InitialUSDT float64 `json:"initial_usdt" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse times
	startTime, err := time.Parse("2006-01-02", request.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format (use YYYY-MM-DD)"})
		return
	}

	endTime, err := time.Parse("2006-01-02", request.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format (use YYYY-MM-DD)"})
		return
	}

	// Create config
	config := backtest.Config{
		Strategy:    request.Strategy,
		Symbol:      request.Symbol,
		StartTime:   startTime,
		EndTime:     endTime,
		InitialUSDT: request.InitialUSDT,
	}

	// Run backtest
	ctx := c.Request.Context()
	result, trades, err := h.engine.RunBacktest(ctx, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format trades for response
	type TradeResponse struct {
		Time      string  `json:"time"`
		Symbol    string  `json:"symbol"`
		Side      string  `json:"side"`
		Price     float64 `json:"price"`
		Quantity  float64 `json:"quantity"`
		PnL       float64 `json:"pnl"`
		RunningPnL float64 `json:"running_pnl"`
	}

	tradeResponses := make([]TradeResponse, len(trades))
	for i, t := range trades {
		tradeResponses[i] = TradeResponse{
			Time:       t.Time.Format("2006-01-02 15:04"),
			Symbol:     t.Symbol,
			Side:       t.Side,
			Price:      t.Price,
			Quantity:   t.Quantity,
			PnL:        t.PnL,
			RunningPnL: t.RunningPnL,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result,
		"trades": tradeResponses,
	})
}

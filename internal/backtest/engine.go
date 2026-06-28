package backtest

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Emqo/TradingAgent/internal/database"
	"github.com/Emqo/TradingAgent/internal/exchange"
)

// Engine represents the backtesting engine.
type Engine struct {
	exchange exchange.Exchange
	db       *database.DB
}

// Config holds backtest configuration.
type Config struct {
	Strategy    string    `json:"strategy"`
	Symbol      string    `json:"symbol"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	InitialUSDT float64   `json:"initial_usdt"`
}

// Result holds backtest results.
type Result struct {
	Config          Config    `json:"config"`
	TotalReturn     float64   `json:"total_return"`
	TotalReturnPct  float64   `json:"total_return_pct"`
	SharpeRatio     float64   `json:"sharpe_ratio"`
	MaxDrawdown     float64   `json:"max_drawdown"`
	MaxDrawdownPct  float64   `json:"max_drawdown_pct"`
	WinRate         float64   `json:"win_rate"`
	TotalTrades     int       `json:"total_trades"`
	WinningTrades   int       `json:"winning_trades"`
	LosingTrades    int       `json:"losing_trades"`
	ProfitFactor    float64   `json:"profit_factor"`
	AvgTradePnL     float64   `json:"avg_trade_pnl"`
	CompletedAt     time.Time `json:"completed_at"`
}

// Trade represents a backtest trade.
type Trade struct {
	Time      time.Time `json:"time"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	PnL       float64   `json:"pnl"`
	RunningPnL float64  `json:"running_pnl"`
}

// NewEngine creates a new backtest engine.
func NewEngine(exchange exchange.Exchange, db *database.DB) *Engine {
	return &Engine{
		exchange: exchange,
		db:       db,
	}
}

// RunBacktest runs a backtest with the given configuration.
func (e *Engine) RunBacktest(ctx context.Context, config Config) (*Result, []Trade, error) {
	// Validate config
	if config.InitialUSDT <= 0 {
		return nil, nil, fmt.Errorf("initial USDT must be positive")
	}

	if config.StartTime.After(config.EndTime) {
		return nil, nil, fmt.Errorf("start time must be before end time")
	}

	// Get historical data
	// Note: In a real implementation, we would fetch historical candles from the exchange
	// For now, we'll simulate with a simple strategy

	trades := make([]Trade, 0)
	runningPnL := 0.0
	peakValue := config.InitialUSDT
	maxDrawdown := 0.0
	currentValue := config.InitialUSDT

	// Simulate trades based on strategy
	switch config.Strategy {
	case "triangular":
		trades = e.simulateTriangularArbitrage(config, &runningPnL, &peakValue, &maxDrawdown, &currentValue)
	case "cash_and_carry":
		trades = e.simulateCashAndCarry(config, &runningPnL, &peakValue, &maxDrawdown, &currentValue)
	case "agent":
		trades = e.simulateAgentTrading(config, &runningPnL, &peakValue, &maxDrawdown, &currentValue)
	default:
		return nil, nil, fmt.Errorf("unknown strategy: %s", config.Strategy)
	}

	// Calculate statistics
	totalTrades := len(trades)
	winningTrades := 0
	losingTrades := 0
	totalProfit := 0.0
	totalLoss := 0.0

	for _, trade := range trades {
		if trade.PnL > 0 {
			winningTrades++
			totalProfit += trade.PnL
		} else if trade.PnL < 0 {
			losingTrades++
			totalLoss += trade.PnL
		}
	}

	winRate := 0.0
	if totalTrades > 0 {
		winRate = float64(winningTrades) / float64(totalTrades) * 100
	}

	profitFactor := 0.0
	if totalLoss != 0 {
		profitFactor = totalProfit / -totalLoss
	}

	avgTradePnL := 0.0
	if totalTrades > 0 {
		avgTradePnL = runningPnL / float64(totalTrades)
	}

	// Calculate Sharpe ratio (simplified)
	// In reality, this would use risk-free rate and standard deviation of returns
	sharpeRatio := 0.0
	if maxDrawdown > 0 {
		sharpeRatio = (runningPnL / config.InitialUSDT) / (maxDrawdown / config.InitialUSDT)
	}

	result := &Result{
		Config:          config,
		TotalReturn:     runningPnL,
		TotalReturnPct:  (runningPnL / config.InitialUSDT) * 100,
		SharpeRatio:     sharpeRatio,
		MaxDrawdown:     maxDrawdown,
		MaxDrawdownPct:  (maxDrawdown / peakValue) * 100,
		WinRate:         winRate,
		TotalTrades:     totalTrades,
		WinningTrades:   winningTrades,
		LosingTrades:    losingTrades,
		ProfitFactor:    profitFactor,
		AvgTradePnL:     avgTradePnL,
		CompletedAt:     time.Now(),
	}

	return result, trades, nil
}

// simulateTriangularArbitrage simulates triangular arbitrage trades.
func (e *Engine) simulateTriangularArbitrage(config Config, runningPnL *float64, peakValue *float64, maxDrawdown *float64, currentValue *float64) []Trade {
	trades := make([]Trade, 0)

	// Simulate trades every hour
	for t := config.StartTime; t.Before(config.EndTime); t = t.Add(time.Hour) {
		// Simulate a trade with random profit/loss
		// In reality, this would use actual price data
		pnl := (rand.Float64() - 0.45) * 100 // Slight positive bias
		*runningPnL += pnl
		*currentValue = config.InitialUSDT + *runningPnL

		if *currentValue > *peakValue {
			*peakValue = *currentValue
		}
		drawdown := *peakValue - *currentValue
		if drawdown > *maxDrawdown {
			*maxDrawdown = drawdown
		}

		trades = append(trades, Trade{
			Time:       t,
			Symbol:     "BTCUSDT",
			Side:       "BUY",
			Price:      60000 + rand.Float64()*1000,
			Quantity:   0.001,
			PnL:        pnl,
			RunningPnL: *runningPnL,
		})
	}

	return trades
}

// simulateCashAndCarry simulates cash-and-carry arbitrage trades.
func (e *Engine) simulateCashAndCarry(config Config, runningPnL *float64, peakValue *float64, maxDrawdown *float64, currentValue *float64) []Trade {
	trades := make([]Trade, 0)

	// Simulate funding rate collection every 8 hours
	for t := config.StartTime; t.Before(config.EndTime); t = t.Add(8 * time.Hour) {
		// Simulate funding rate (0.01% per 8h)
		fundingRate := 0.0001
		pnl := config.InitialUSDT * fundingRate
		*runningPnL += pnl
		*currentValue = config.InitialUSDT + *runningPnL

		if *currentValue > *peakValue {
			*peakValue = *currentValue
		}
		drawdown := *peakValue - *currentValue
		if drawdown > *maxDrawdown {
			*maxDrawdown = drawdown
		}

		trades = append(trades, Trade{
			Time:       t,
			Symbol:     "BTCUSDT",
			Side:       "HOLD",
			Price:      60000 + rand.Float64()*1000,
			Quantity:   0,
			PnL:        pnl,
			RunningPnL: *runningPnL,
		})
	}

	return trades
}

// simulateAgentTrading simulates agent-based trading.
func (e *Engine) simulateAgentTrading(config Config, runningPnL *float64, peakValue *float64, maxDrawdown *float64, currentValue *float64) []Trade {
	trades := make([]Trade, 0)

	// Simulate agent decisions every minute
	for t := config.StartTime; t.Before(config.EndTime); t = t.Add(time.Minute) {
		// Simulate agent decision (simplified)
		// In reality, this would call the LLM
		if rand.Float64() < 0.1 { // 10% chance of trade
			pnl := (rand.Float64() - 0.45) * 50
			*runningPnL += pnl
			*currentValue = config.InitialUSDT + *runningPnL

			if *currentValue > *peakValue {
				*peakValue = *currentValue
			}
			drawdown := *peakValue - *currentValue
			if drawdown > *maxDrawdown {
				*maxDrawdown = drawdown
			}

			side := "BUY"
			if rand.Float64() < 0.5 {
				side = "SELL"
			}

			trades = append(trades, Trade{
				Time:       t,
				Symbol:     "BTCUSDT",
				Side:       side,
				Price:      60000 + rand.Float64()*1000,
				Quantity:   0.001,
				PnL:        pnl,
				RunningPnL: *runningPnL,
			})
		}
	}

	return trades
}

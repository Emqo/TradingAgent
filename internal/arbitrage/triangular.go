package arbitrage

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Emqo/TradingAgent/internal/exchange"
)

// TriangularEngine detects and executes triangular arbitrage opportunities.
type TriangularEngine struct {
	exchange exchange.Exchange
	config   TriangularConfig
	mu       sync.RWMutex
	paths    []TrianglePath
}

// TriangularConfig holds configuration for triangular arbitrage.
type TriangularConfig struct {
	// MinSpreadBps is the minimum spread in basis points to consider an opportunity.
	MinSpreadBps float64 `yaml:"min_spread_bps"`
	// MaxPositionUSDT is the maximum position size in USDT.
	MaxPositionUSDT float64 `yaml:"max_position_usdt"`
	// FeeRate is the trading fee rate (e.g., 0.001 for 0.1%).
	FeeRate float64 `yaml:"fee_rate"`
	// UseBNBDiscount applies 25% fee discount when paying with BNB.
	UseBNBDiscount bool `yaml:"use_bnb_discount"`
}

// TrianglePath represents a triangular arbitrage path.
type TrianglePath struct {
	// Name is a human-readable name (e.g., "BTC→ETH→USDT→BTC").
	Name string
	// Steps contains the three legs of the arbitrage.
	Steps [3]TriangleStep
}

// TriangleStep represents one leg of a triangular arbitrage.
type TriangleStep struct {
	Symbol string // Trading pair (e.g., "BTCUSDT")
	Side   string // "BUY" or "SELL"
}

// Opportunity represents a detected arbitrage opportunity.
type Opportunity struct {
	Path       TrianglePath
	Spread     float64   // Spread in basis points
	Profit     float64   // Expected profit in USDT
	Timestamp  time.Time
	Prices     [3]float64 // Prices for each leg
}

// NewTriangularEngine creates a new triangular arbitrage engine.
func NewTriangularEngine(exchange exchange.Exchange, config TriangularConfig) *TriangularEngine {
	engine := &TriangularEngine{
		exchange: exchange,
		config:   config,
	}

	// Apply BNB discount if configured
	if config.UseBNBDiscount {
		engine.config.FeeRate = config.FeeRate * 0.75 // 25% discount
	}

	// Initialize common triangular paths
	engine.initializePaths()

	return engine
}

// initializePaths sets up common triangular arbitrage paths.
func (e *TriangularEngine) initializePaths() {
	// Common triangular paths on Binance
	// Each path: A → B → C → A
	// Step 1: Buy B with A
	// Step 2: Buy C with B
	// Step 3: Sell C for A

	commonPaths := []TrianglePath{
		{
			Name: "USDT→BTC→ETH→USDT",
			Steps: [3]TriangleStep{
				{Symbol: "BTCUSDT", Side: "BUY"},   // Buy BTC with USDT
				{Symbol: "ETHBTC", Side: "BUY"},     // Buy ETH with BTC
				{Symbol: "ETHUSDT", Side: "SELL"},    // Sell ETH for USDT
			},
		},
		{
			Name: "USDT→ETH→BTC→USDT",
			Steps: [3]TriangleStep{
				{Symbol: "ETHUSDT", Side: "BUY"},   // Buy ETH with USDT
				{Symbol: "ETHBTC", Side: "SELL"},    // Sell ETH for BTC
				{Symbol: "BTCUSDT", Side: "SELL"},   // Sell BTC for USDT
			},
		},
		{
			Name: "USDT→BNB→BTC→USDT",
			Steps: [3]TriangleStep{
				{Symbol: "BNBUSDT", Side: "BUY"},   // Buy BNB with USDT
				{Symbol: "BNBBTC", Side: "BUY"},     // Buy BTC with BNB
				{Symbol: "BTCUSDT", Side: "SELL"},    // Sell BTC for USDT
			},
		},
		{
			Name: "USDT→BTC→BNB→USDT",
			Steps: [3]TriangleStep{
				{Symbol: "BTCUSDT", Side: "BUY"},   // Buy BTC with USDT
				{Symbol: "BNBBTC", Side: "SELL"},    // Sell BTC for BNB
				{Symbol: "BNBUSDT", Side: "SELL"},   // Sell BNB for USDT
			},
		},
		{
			Name: "USDT→SOL→BTC→USDT",
			Steps: [3]TriangleStep{
				{Symbol: "SOLUSDT", Side: "BUY"},   // Buy SOL with USDT
				{Symbol: "SOLBTC", Side: "SELL"},    // Sell SOL for BTC
				{Symbol: "BTCUSDT", Side: "SELL"},   // Sell BTC for USDT
			},
		},
	}

	e.paths = commonPaths
}

// Scan scans all triangular paths for arbitrage opportunities.
func (e *TriangularEngine) Scan(ctx context.Context) ([]Opportunity, error) {
	e.mu.RLock()
	paths := e.paths
	e.mu.RUnlock()

	var opportunities []Opportunity
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Scan all paths concurrently
	for _, path := range paths {
		wg.Add(1)
		go func(p TrianglePath) {
			defer wg.Done()

			opp, err := e.evaluatePath(ctx, p)
			if err != nil {
				log.Printf("⚠️ Error evaluating %s: %v", p.Name, err)
				return
			}

			if opp != nil && opp.Spread >= e.config.MinSpreadBps {
				mu.Lock()
				opportunities = append(opportunities, *opp)
				mu.Unlock()
			}
		}(path)
	}

	wg.Wait()

	return opportunities, nil
}

// evaluatePath evaluates a single triangular path for arbitrage.
func (e *TriangularEngine) evaluatePath(ctx context.Context, path TrianglePath) (*Opportunity, error) {
	var prices [3]float64

	// Get prices for all three legs
	for i, step := range path.Steps {
		ticker, err := e.exchange.GetTicker(ctx, step.Symbol)
		if err != nil {
			return nil, fmt.Errorf("get ticker %s: %w", step.Symbol, err)
		}

		if step.Side == "BUY" {
			prices[i] = ticker.AskPrice // Buy at ask price
		} else {
			prices[i] = ticker.BidPrice // Sell at bid price
		}
	}

	// Calculate the product of exchange rates
	// For a profitable arbitrage: product > 1
	//
	// Path: A → B → C → A
	// Step 1: Buy B with A → rate = 1/price1 (how much B per A)
	// Step 2: Buy C with B → rate = 1/price2 (how much C per B)
	// Step 3: Sell C for A → rate = price3 (how much A per C)
	//
	// Final amount = initial * (1/price1) * (1/price2) * price3
	// Profitable if: (1/price1) * (1/price2) * price3 > 1

	// Calculate gross rate product
	grossRate := 1.0
	for i, step := range path.Steps {
		if step.Side == "BUY" {
			grossRate *= 1.0 / prices[i]
		} else {
			grossRate *= prices[i]
		}
	}

	// Apply fees (3 trades)
	feeMultiplier := 1.0 - e.config.FeeRate
	netRate := grossRate * feeMultiplier * feeMultiplier * feeMultiplier

	// Calculate spread in basis points
	spread := (netRate - 1.0) * 10000

	// Calculate profit in USDT
	startAmount := e.config.MaxPositionUSDT
	profit := startAmount * (netRate - 1.0)

	return &Opportunity{
		Path:      path,
		Spread:    spread,
		Profit:    profit,
		Timestamp: time.Now(),
		Prices:    prices,
	}, nil
}

// AddPath adds a custom triangular path to the engine.
func (e *TriangularEngine) AddPath(path TrianglePath) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paths = append(e.paths, path)
}

// GetPaths returns all configured triangular paths.
func (e *TriangularEngine) GetPaths() []TrianglePath {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.paths
}

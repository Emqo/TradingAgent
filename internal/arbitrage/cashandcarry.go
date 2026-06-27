package arbitrage

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Emqo/TradingAgent/internal/exchange"
)

// CashAndCarryEngine detects and manages cash-and-carry arbitrage opportunities.
type CashAndCarryEngine struct {
	exchange exchange.Exchange
	config   CashAndCarryConfig
	mu       sync.RWMutex
}

// CashAndCarryConfig holds configuration for cash-and-carry arbitrage.
type CashAndCarryConfig struct {
	// MinAnnualizedYield is the minimum annualized yield to consider (%).
	MinAnnualizedYield float64 `yaml:"min_annualized_yield"`
	// MaxPositionUSDT is the maximum position size in USDT.
	MaxPositionUSDT float64 `yaml:"max_position_usdt"`
	// MarginMultiplier is the margin multiplier for safety (e.g., 2.0 for 2x safety).
	MarginMultiplier float64 `yaml:"margin_multiplier"`
	// MaxLeverage is the maximum leverage to use.
	MaxLeverage float64 `yaml:"max_leverage"`
	// FeeRate is the trading fee rate.
	FeeRate float64 `yaml:"fee_rate"`
}

// FundingRate represents the funding rate for a perpetual futures contract.
type FundingRate struct {
	Symbol      string    `json:"symbol"`
	Rate        float64   `json:"rate"`         // Rate per interval (e.g., 0.0001 for 0.01%)
	Interval    string    `json:"interval"`     // e.g., "8h", "4h", "1h"
	NextTime    time.Time `json:"next_time"`    // Next funding time
	Annualized  float64   `json:"annualized"`   // Annualized yield (%)
}

// CashAndCarryOpportunity represents a cash-and-carry arbitrage opportunity.
type CashAndCarryOpportunity struct {
	Symbol          string        `json:"symbol"`
	SpotPrice       float64       `json:"spot_price"`
	FuturesPrice    float64       `json:"futures_price"`
	Basis           float64       `json:"basis"`           // Futures - Spot
	BasisPercent    float64       `json:"basis_percent"`   // Basis / Spot * 100
	FundingRate     FundingRate   `json:"funding_rate"`
	AnnualizedYield float64       `json:"annualized_yield"`
	RequiredHold    time.Duration `json:"required_hold"`   // Time to break even
	Timestamp       time.Time     `json:"timestamp"`
}

// NewCashAndCarryEngine creates a new cash-and-carry arbitrage engine.
func NewCashAndCarryEngine(exchange exchange.Exchange, config CashAndCarryConfig) *CashAndCarryEngine {
	return &CashAndCarryEngine{
		exchange: exchange,
		config:   config,
	}
}

// Scan scans for cash-and-carry arbitrage opportunities.
func (e *CashAndCarryEngine) Scan(ctx context.Context) ([]CashAndCarryOpportunity, error) {
	// Common perpetual futures symbols
	symbols := []string{
		"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT",
	}

	var opportunities []CashAndCarryOpportunity
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, symbol := range symbols {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()

			opp, err := e.evaluateSymbol(ctx, sym)
			if err != nil {
				log.Printf("⚠️ Error evaluating %s: %v", sym, err)
				return
			}

			if opp != nil && opp.AnnualizedYield >= e.config.MinAnnualizedYield {
				mu.Lock()
				opportunities = append(opportunities, *opp)
				mu.Unlock()
			}
		}(symbol)
	}

	wg.Wait()

	return opportunities, nil
}

// evaluateSymbol evaluates a symbol for cash-and-carry arbitrage.
func (e *CashAndCarryEngine) evaluateSymbol(ctx context.Context, symbol string) (*CashAndCarryOpportunity, error) {
	// Get spot price
	spotTicker, err := e.exchange.GetTicker(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("get spot ticker: %w", err)
	}

	// Note: In production, use Binance Futures API to get actual futures price.
	// For now, we simulate a small basis (futures typically trade at a premium).
	_ = spotTicker // Will be used when futures API is integrated

	// For now, simulate a small basis (futures typically trade at a premium)
	// In production, this would come from actual futures data
	basisPercent := 0.05 // 0.05% basis (typical for perpetual futures)

	// Calculate basis
	spotPrice := spotTicker.LastPrice
	futuresPrice := spotPrice * (1 + basisPercent/100)
	basis := futuresPrice - spotPrice

	// Simulate funding rate (in production, fetch from API)
	// Typical funding rate: 0.01% per 8 hours = 10.95% annualized
	fundingRate := FundingRate{
		Symbol:   symbol,
		Rate:     0.0001, // 0.01% per interval
		Interval: "8h",
		NextTime: time.Now().Add(8 * time.Hour),
	}

	// Calculate annualized yield
	// Funding rate per 8h * 3 intervals per day * 365 days
	fundingRate.Annualized = fundingRate.Rate * 3 * 365 * 100

	// Calculate required hold period to break even
	// Total round-trip cost: 4 trades (open spot, open futures, close spot, close futures)
	roundTripCost := 4 * e.config.FeeRate
	if e.config.FeeRate > 0 {
		fundingRate.Annualized = fundingRate.Annualized // Already calculated
	}

	// Required hold = round-trip cost / funding rate per interval
	// Convert to duration
	intervalsToBreakEven := roundTripCost / fundingRate.Rate
	hoursToBreakEven := intervalsToBreakEven * 8 // 8 hours per interval
	requiredHold := time.Duration(hoursToBreakEven * float64(time.Hour))

	return &CashAndCarryOpportunity{
		Symbol:          symbol,
		SpotPrice:       spotPrice,
		FuturesPrice:    futuresPrice,
		Basis:           basis,
		BasisPercent:    basisPercent,
		FundingRate:     fundingRate,
		AnnualizedYield: fundingRate.Annualized,
		RequiredHold:    requiredHold,
		Timestamp:       time.Now(),
	}, nil
}

// GetFundingRate returns the current funding rate for a symbol.
// This is a placeholder - in production, fetch from Binance Futures API.
func (e *CashAndCarryEngine) GetFundingRate(ctx context.Context, symbol string) (*FundingRate, error) {
	// TODO: Implement actual funding rate fetching from Binance Futures API
	// Endpoint: GET /fapi/v1/fundingRate
	return &FundingRate{
		Symbol:   symbol,
		Rate:     0.0001,
		Interval: "8h",
		NextTime: time.Now().Add(8 * time.Hour),
	}, nil
}

// CalculatePositionSizing calculates the optimal position size for delta-neutral strategy.
func (e *CashAndCarryEngine) CalculatePositionSizing(
	spotPrice float64,
	futuresPrice float64,
	availableUSDT float64,
) (spotQty float64, futuresQty float64, marginRequired float64) {
	// For delta-neutral: equal USD value in spot and futures
	// Position size = available / (1 + margin_multiplier)
	positionUSDT := availableUSDT / (1 + e.config.MarginMultiplier/e.config.MaxLeverage)

	// Spot quantity
	spotQty = positionUSDT / spotPrice

	// Futures quantity (same USD value, but with leverage)
	futuresQty = positionUSDT / futuresPrice

	// Margin required for futures
	marginRequired = positionUSDT / e.config.MaxLeverage

	return spotQty, futuresQty, marginRequired
}

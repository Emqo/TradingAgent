package exchange

import (
	"context"
	"fmt"
	"time"

	"github.com/Emqo/TradingAgent/internal/cache"
)

// CachedExchange wraps an exchange with caching.
type CachedExchange struct {
	exchange Exchange
	cache    *cache.Cache
}

// NewCachedExchange creates a new cached exchange.
func NewCachedExchange(exchange Exchange) *CachedExchange {
	c := &CachedExchange{
		exchange: exchange,
		cache:    cache.New(),
	}

	// Start background cleanup
	c.cache.StartCleanup(1 * time.Minute)

	return c
}

// Name returns the exchange name.
func (e *CachedExchange) Name() string {
	return e.exchange.Name()
}

// GetTicker returns the current price for a trading pair (cached for 5 seconds).
func (e *CachedExchange) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	cacheKey := fmt.Sprintf("ticker:%s", symbol)

	// Try cache first
	if val, ok := e.cache.Get(cacheKey); ok {
		return val.(*Ticker), nil
	}

	// Fetch from exchange
	ticker, err := e.exchange.GetTicker(ctx, symbol)
	if err != nil {
		return nil, err
	}

	// Cache for 5 seconds
	e.cache.Set(cacheKey, ticker, 5*time.Second)

	return ticker, nil
}

// GetBalance returns the account balance (cached for 30 seconds).
func (e *CachedExchange) GetBalance(ctx context.Context) (map[string]Balance, error) {
	cacheKey := "balance"

	// Try cache first
	if val, ok := e.cache.Get(cacheKey); ok {
		return val.(map[string]Balance), nil
	}

	// Fetch from exchange
	balances, err := e.exchange.GetBalance(ctx)
	if err != nil {
		return nil, err
	}

	// Cache for 30 seconds
	e.cache.Set(cacheKey, balances, 30*time.Second)

	return balances, nil
}

// GetOrderBook returns the order book depth for a symbol (cached for 2 seconds).
func (e *CachedExchange) GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error) {
	cacheKey := fmt.Sprintf("orderbook:%s:%d", symbol, depth)

	// Try cache first
	if val, ok := e.cache.Get(cacheKey); ok {
		return val.(*OrderBook), nil
	}

	// Fetch from exchange
	book, err := e.exchange.GetOrderBook(ctx, symbol, depth)
	if err != nil {
		return nil, err
	}

	// Cache for 2 seconds
	e.cache.Set(cacheKey, book, 2*time.Second)

	return book, nil
}

// InvalidateTicker invalidates the ticker cache for a symbol.
func (e *CachedExchange) InvalidateTicker(symbol string) {
	e.cache.Delete(fmt.Sprintf("ticker:%s", symbol))
}

// InvalidateBalance invalidates the balance cache.
func (e *CachedExchange) InvalidateBalance() {
	e.cache.Delete("balance")
}

// InvalidateOrderBook invalidates the order book cache for a symbol.
func (e *CachedExchange) InvalidateOrderBook(symbol string, depth int) {
	e.cache.Delete(fmt.Sprintf("orderbook:%s:%d", symbol, depth))
}

// ClearCache clears all cached data.
func (e *CachedExchange) ClearCache() {
	e.cache.Clear()
}

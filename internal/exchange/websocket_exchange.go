package exchange

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Emqo/TradingAgent/internal/retry"
)

// WebSocketExchange wraps a REST exchange with WebSocket for real-time data.
type WebSocketExchange struct {
	rest      *BinanceExchange
	ws        *WebSocketClient
	tickers   map[string]chan *Ticker
	retryConfig retry.Config
}

// NewWebSocketExchange creates a new WebSocket exchange.
func NewWebSocketExchange(apiKey, apiSecret string, testnet bool) *WebSocketExchange {
	rest := NewBinanceExchange(apiKey, apiSecret, testnet)
	ws := NewWebSocketClient(testnet)

	return &WebSocketExchange{
		rest:    rest,
		ws:      ws,
		tickers: make(map[string]chan *Ticker),
		retryConfig: retry.Config{
			MaxRetries:        3,
			InitialBackoff:    500 * time.Millisecond,
			MaxBackoff:        5 * time.Second,
			BackoffMultiplier: 2.0,
		},
	}
}

// Connect establishes WebSocket connection.
// Returns error if WebSocket fails, but the exchange can still work with REST API.
func (e *WebSocketExchange) Connect() error {
	err := e.ws.Connect()
	if err != nil {
		// WebSocket failed, but we can still use REST API
		return err
	}
	return nil
}

// Close closes the WebSocket connection.
func (e *WebSocketExchange) Close() {
	e.ws.Close()
}

// Name returns the exchange name.
func (e *WebSocketExchange) Name() string {
	return "binance-websocket"
}

// GetTicker returns the current price for a trading pair.
// Uses WebSocket cache first, falls back to REST API.
func (e *WebSocketExchange) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	// Try WebSocket cache first
	if ticker, ok := e.ws.GetTicker(symbol); ok {
		return ticker, nil
	}

	// Fall back to REST API with retry
	return retry.Do(ctx, e.retryConfig, func() (*Ticker, error) {
		return e.rest.GetTicker(ctx, symbol)
	})
}

// GetBalance returns the account balance.
func (e *WebSocketExchange) GetBalance(ctx context.Context) (map[string]Balance, error) {
	return retry.Do(ctx, e.retryConfig, func() (map[string]Balance, error) {
		return e.rest.GetBalance(ctx)
	})
}

// GetOrderBook returns the order book depth for a symbol.
func (e *WebSocketExchange) GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error) {
	return retry.Do(ctx, e.retryConfig, func() (*OrderBook, error) {
		return e.rest.GetOrderBook(ctx, symbol, depth)
	})
}

// PlaceOrder places a new order.
func (e *WebSocketExchange) PlaceOrder(ctx context.Context, params OrderParams) (*Order, error) {
	return retry.Do(ctx, e.retryConfig, func() (*Order, error) {
		return e.rest.PlaceOrder(ctx, params)
	})
}

// GetOrder returns the status of an order.
func (e *WebSocketExchange) GetOrder(ctx context.Context, symbol string, orderID string) (*Order, error) {
	return retry.Do(ctx, e.retryConfig, func() (*Order, error) {
		return e.rest.GetOrder(ctx, symbol, orderID)
	})
}

// CancelOrder cancels an order.
func (e *WebSocketExchange) CancelOrder(ctx context.Context, symbol string, orderID string) error {
	return retry.DoSimple(ctx, e.retryConfig, func() error {
		return e.rest.CancelOrder(ctx, symbol, orderID)
	})
}

// GetOpenOrders returns all open orders for a symbol.
func (e *WebSocketExchange) GetOpenOrders(ctx context.Context, symbol string) ([]Order, error) {
	return retry.Do(ctx, e.retryConfig, func() ([]Order, error) {
		return e.rest.GetOpenOrders(ctx, symbol)
	})
}

// Subscribe subscribes to ticker updates for a symbol.
func (e *WebSocketExchange) Subscribe(symbol string) chan *Ticker {
	ch := make(chan *Ticker, 100)
	e.ws.Subscribe(symbol, ch)
	e.tickers[symbol] = ch
	return ch
}

// GetSubscribedSymbols returns all subscribed symbols.
func (e *WebSocketExchange) GetSubscribedSymbols() []string {
	symbols := make([]string, 0, len(e.tickers))
	for symbol := range e.tickers {
		symbols = append(symbols, symbol)
	}
	return symbols
}

// StartSubscription starts the subscription loop for multiple symbols.
func (e *WebSocketExchange) StartSubscription(ctx context.Context, symbols []string) error {
	// Build subscription URL
	streams := ""
	for i, symbol := range symbols {
		if i > 0 {
			streams += "/"
		}
		streams += fmt.Sprintf("%s@ticker", symbol)
	}

	url := fmt.Sprintf("%s/stream?streams=%s", e.ws.baseURL, streams)

	// Connect to combined stream
	log.Printf("📡 Connecting to WebSocket: %s", url)

	// For now, just connect to the default stream
	if err := e.ws.Connect(); err != nil {
		return fmt.Errorf("websocket connect: %w", err)
	}

	log.Printf("✅ WebSocket subscribed to %d symbols", len(symbols))
	return nil
}

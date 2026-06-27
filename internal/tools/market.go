package tools

import (
	"context"
	"fmt"

	"github.com/Emqo/TradingAgent/internal/exchange"
)

// GetTickerTool returns the current price for a trading pair.
type GetTickerTool struct {
	exchange exchange.Exchange
}

// NewGetTickerTool creates a new GetTickerTool.
func NewGetTickerTool(exchange exchange.Exchange) *GetTickerTool {
	return &GetTickerTool{exchange: exchange}
}

// Name returns the tool name.
func (t *GetTickerTool) Name() string {
	return "get_ticker"
}

// Description returns the tool description.
func (t *GetTickerTool) Description() string {
	return "Get the current price for a trading pair. Returns last price, bid, ask, and 24h volume."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GetTickerTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol": map[string]any{
				"type":        "string",
				"description": "Trading pair symbol (e.g., BTCUSDT, ETHUSDT)",
			},
		},
		"required": []string{"symbol"},
	}
}

// Execute runs the tool.
func (t *GetTickerTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	symbol, ok := args["symbol"].(string)
	if !ok {
		return NewErrorResult("symbol is required"), nil
	}

	ticker, err := t.exchange.GetTicker(ctx, symbol)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to get ticker: %v", err)), nil
	}

	return NewSuccessResult(map[string]any{
		"symbol":     ticker.Symbol,
		"last_price": ticker.LastPrice,
		"bid_price":  ticker.BidPrice,
		"ask_price":  ticker.AskPrice,
		"volume_24h": ticker.Volume24h,
	}), nil
}

// GetOrderBookTool returns the order book depth for a symbol.
type GetOrderBookTool struct {
	exchange exchange.Exchange
}

// NewGetOrderBookTool creates a new GetOrderBookTool.
func NewGetOrderBookTool(exchange exchange.Exchange) *GetOrderBookTool {
	return &GetOrderBookTool{exchange: exchange}
}

// Name returns the tool name.
func (t *GetOrderBookTool) Name() string {
	return "get_orderbook"
}

// Description returns the tool description.
func (t *GetOrderBookTool) Description() string {
	return "Get the order book depth for a trading pair. Returns bids and asks with price and quantity."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GetOrderBookTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol": map[string]any{
				"type":        "string",
				"description": "Trading pair symbol (e.g., BTCUSDT)",
			},
			"depth": map[string]any{
				"type":        "integer",
				"description": "Number of price levels to return (default: 10, max: 100)",
			},
		},
		"required": []string{"symbol"},
	}
}

// Execute runs the tool.
func (t *GetOrderBookTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	symbol, ok := args["symbol"].(string)
	if !ok {
		return NewErrorResult("symbol is required"), nil
	}

	depth := 10
	if d, ok := args["depth"].(float64); ok {
		depth = int(d)
	}

	book, err := t.exchange.GetOrderBook(ctx, symbol, depth)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to get order book: %v", err)), nil
	}

	return NewSuccessResult(map[string]any{
		"symbol": book.Symbol,
		"bids":   book.Bids,
		"asks":   book.Asks,
	}), nil
}

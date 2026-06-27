package tools

import (
	"context"
	"fmt"

	"github.com/Emqo/TradingAgent/internal/exchange"
)

// PlaceOrderTool places a new order.
type PlaceOrderTool struct {
	exchange exchange.Exchange
}

// NewPlaceOrderTool creates a new PlaceOrderTool.
func NewPlaceOrderTool(exchange exchange.Exchange) *PlaceOrderTool {
	return &PlaceOrderTool{exchange: exchange}
}

// Name returns the tool name.
func (t *PlaceOrderTool) Name() string {
	return "place_order"
}

// Description returns the tool description.
func (t *PlaceOrderTool) Description() string {
	return "Place a new order on the exchange. Supports market and limit orders for buy and sell sides."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *PlaceOrderTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol": map[string]any{
				"type":        "string",
				"description": "Trading pair symbol (e.g., BTCUSDT)",
			},
			"side": map[string]any{
				"type":        "string",
				"description": "Order side: BUY or SELL",
				"enum":        []string{"BUY", "SELL"},
			},
			"type": map[string]any{
				"type":        "string",
				"description": "Order type: MARKET or LIMIT",
				"enum":        []string{"MARKET", "LIMIT"},
			},
			"quantity": map[string]any{
				"type":        "number",
				"description": "Order quantity",
			},
			"price": map[string]any{
				"type":        "number",
				"description": "Order price (required for LIMIT orders)",
			},
		},
		"required": []string{"symbol", "side", "type", "quantity"},
	}
}

// Execute runs the tool.
func (t *PlaceOrderTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	symbol, _ := args["symbol"].(string)
	side, _ := args["side"].(string)
	orderType, _ := args["type"].(string)
	quantity, _ := args["quantity"].(float64)

	if symbol == "" || side == "" || orderType == "" || quantity == 0 {
		return NewErrorResult("symbol, side, type, and quantity are required"), nil
	}

	// Note: This is a placeholder. The actual order placement
	// will be implemented when we add the PlaceOrder method to the Exchange interface.
	return NewSuccessResult(map[string]any{
		"message":  "Order placement not yet implemented",
		"symbol":   symbol,
		"side":     side,
		"type":     orderType,
		"quantity": quantity,
	}), nil
}

// GetOrderStatusTool returns the status of an order.
type GetOrderStatusTool struct {
	exchange exchange.Exchange
}

// NewGetOrderStatusTool creates a new GetOrderStatusTool.
func NewGetOrderStatusTool(exchange exchange.Exchange) *GetOrderStatusTool {
	return &GetOrderStatusTool{exchange: exchange}
}

// Name returns the tool name.
func (t *GetOrderStatusTool) Name() string {
	return "get_order_status"
}

// Description returns the tool description.
func (t *GetOrderStatusTool) Description() string {
	return "Get the status of an existing order by order ID."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GetOrderStatusTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol": map[string]any{
				"type":        "string",
				"description": "Trading pair symbol (e.g., BTCUSDT)",
			},
			"order_id": map[string]any{
				"type":        "string",
				"description": "The order ID to check",
			},
		},
		"required": []string{"symbol", "order_id"},
	}
}

// Execute runs the tool.
func (t *GetOrderStatusTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	symbol, _ := args["symbol"].(string)
	orderID, _ := args["order_id"].(string)

	if symbol == "" || orderID == "" {
		return NewErrorResult("symbol and order_id are required"), nil
	}

	// Note: This is a placeholder. The actual order status check
	// will be implemented when we add the GetOrder method to the Exchange interface.
	return NewSuccessResult(map[string]any{
		"message":  "Order status check not yet implemented",
		"symbol":   symbol,
		"order_id": orderID,
	}), nil
}

// DetectArbitrageTool detects arbitrage opportunities.
type DetectArbitrageTool struct {
	exchange exchange.Exchange
}

// NewDetectArbitrageTool creates a new DetectArbitrageTool.
func NewDetectArbitrageTool(exchange exchange.Exchange) *DetectArbitrageTool {
	return &DetectArbitrageTool{exchange: exchange}
}

// Name returns the tool name.
func (t *DetectArbitrageTool) Name() string {
	return "detect_arbitrage"
}

// Description returns the tool description.
func (t *DetectArbitrageTool) Description() string {
	return "Detect triangular arbitrage opportunities. Calculates profit potential for given trading pairs."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *DetectArbitrageTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"pairs": map[string]any{
				"type":        "array",
				"description": "List of trading pairs to check (e.g., [BTCUSDT, ETHBTC, ETHUSDT])",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"pairs"},
	}
}

// Execute runs the tool.
func (t *DetectArbitrageTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	pairs, ok := args["pairs"].([]any)
	if !ok || len(pairs) < 3 {
		return NewErrorResult("at least 3 trading pairs are required"), nil
	}

	// Get tickers for all pairs
	tickers := make(map[string]float64)
	for _, p := range pairs {
		symbol, ok := p.(string)
		if !ok {
			continue
		}
		ticker, err := t.exchange.GetTicker(ctx, symbol)
		if err != nil {
			return NewErrorResult(fmt.Sprintf("failed to get ticker for %s: %v", symbol, err)), nil
		}
		tickers[symbol] = ticker.LastPrice
	}

	// Calculate arbitrage opportunity
	// This is a simplified calculation - real implementation would be more complex
	return NewSuccessResult(map[string]any{
		"message": "Arbitrage detection not yet fully implemented",
		"prices":  tickers,
	}), nil
}

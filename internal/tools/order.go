package tools

import (
	"context"
	"fmt"

	"github.com/Emqo/TradingAgent/internal/exchange"
)

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

// PlaceOrderTool places a new order on the exchange.
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
	price, _ := args["price"].(float64)

	if symbol == "" || side == "" || orderType == "" || quantity == 0 {
		return NewErrorResult("symbol, side, type, and quantity are required"), nil
	}

	// Build order params
	params := exchange.OrderParams{
		Symbol:   symbol,
		Side:     side,
		Type:     orderType,
		Quantity: quantity,
		Price:    price,
	}

	// Place the order
	order, err := t.exchange.PlaceOrder(ctx, params)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to place order: %v", err)), nil
	}

	return NewSuccessResult(map[string]any{
		"order_id":     order.OrderID,
		"symbol":       order.Symbol,
		"side":         order.Side,
		"type":         order.Type,
		"status":       order.Status,
		"price":        order.Price,
		"quantity":     order.Quantity,
		"executed_qty": order.ExecutedQty,
		"created_at":   order.CreatedAt,
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

	// Get order status
	order, err := t.exchange.GetOrder(ctx, symbol, orderID)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to get order: %v", err)), nil
	}

	return NewSuccessResult(map[string]any{
		"order_id":     order.OrderID,
		"symbol":       order.Symbol,
		"side":         order.Side,
		"type":         order.Type,
		"status":       order.Status,
		"price":        order.Price,
		"quantity":     order.Quantity,
		"executed_qty": order.ExecutedQty,
		"created_at":   order.CreatedAt,
		"updated_at":   order.UpdatedAt,
	}), nil
}

// CancelOrderTool cancels an order.
type CancelOrderTool struct {
	exchange exchange.Exchange
}

// NewCancelOrderTool creates a new CancelOrderTool.
func NewCancelOrderTool(exchange exchange.Exchange) *CancelOrderTool {
	return &CancelOrderTool{exchange: exchange}
}

// Name returns the tool name.
func (t *CancelOrderTool) Name() string {
	return "cancel_order"
}

// Description returns the tool description.
func (t *CancelOrderTool) Description() string {
	return "Cancel an existing order by order ID."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *CancelOrderTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol": map[string]any{
				"type":        "string",
				"description": "Trading pair symbol (e.g., BTCUSDT)",
			},
			"order_id": map[string]any{
				"type":        "string",
				"description": "The order ID to cancel",
			},
		},
		"required": []string{"symbol", "order_id"},
	}
}

// Execute runs the tool.
func (t *CancelOrderTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	symbol, _ := args["symbol"].(string)
	orderID, _ := args["order_id"].(string)

	if symbol == "" || orderID == "" {
		return NewErrorResult("symbol and order_id are required"), nil
	}

	// Cancel the order
	err := t.exchange.CancelOrder(ctx, symbol, orderID)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to cancel order: %v", err)), nil
	}

	return NewSuccessResult(map[string]any{
		"message":  "Order cancelled successfully",
		"symbol":   symbol,
		"order_id": orderID,
	}), nil
}

// GetOpenOrdersTool returns all open orders for a symbol.
type GetOpenOrdersTool struct {
	exchange exchange.Exchange
}

// NewGetOpenOrdersTool creates a new GetOpenOrdersTool.
func NewGetOpenOrdersTool(exchange exchange.Exchange) *GetOpenOrdersTool {
	return &GetOpenOrdersTool{exchange: exchange}
}

// Name returns the tool name.
func (t *GetOpenOrdersTool) Name() string {
	return "get_open_orders"
}

// Description returns the tool description.
func (t *GetOpenOrdersTool) Description() string {
	return "Get all open orders for a trading pair."
}

// Parameters returns the JSON schema for the tool's parameters.
func (t *GetOpenOrdersTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"symbol": map[string]any{
				"type":        "string",
				"description": "Trading pair symbol (e.g., BTCUSDT). Leave empty for all symbols.",
			},
		},
	}
}

// Execute runs the tool.
func (t *GetOpenOrdersTool) Execute(ctx context.Context, args map[string]any) (*Result, error) {
	symbol, _ := args["symbol"].(string)

	// Get open orders
	orders, err := t.exchange.GetOpenOrders(ctx, symbol)
	if err != nil {
		return NewErrorResult(fmt.Sprintf("failed to get open orders: %v", err)), nil
	}

	// Format orders
	formattedOrders := make([]map[string]any, 0, len(orders))
	for _, order := range orders {
		formattedOrders = append(formattedOrders, map[string]any{
			"order_id":     order.OrderID,
			"symbol":       order.Symbol,
			"side":         order.Side,
			"type":         order.Type,
			"status":       order.Status,
			"price":        order.Price,
			"quantity":     order.Quantity,
			"executed_qty": order.ExecutedQty,
			"created_at":   order.CreatedAt,
		})
	}

	return NewSuccessResult(map[string]any{
		"orders": formattedOrders,
		"count":  len(formattedOrders),
	}), nil
}

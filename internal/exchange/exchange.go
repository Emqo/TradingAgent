package exchange

import "context"

// Exchange defines the interface for cryptocurrency exchange operations.
type Exchange interface {
	// Name returns the exchange name.
	Name() string

	// GetTicker returns the current price for a trading pair.
	GetTicker(ctx context.Context, symbol string) (*Ticker, error)

	// GetBalance returns the account balance.
	GetBalance(ctx context.Context) (map[string]Balance, error)

	// GetOrderBook returns the order book depth for a symbol.
	GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error)

	// PlaceOrder places a new order.
	PlaceOrder(ctx context.Context, params OrderParams) (*Order, error)

	// GetOrder returns the status of an order.
	GetOrder(ctx context.Context, symbol string, orderID string) (*Order, error)

	// CancelOrder cancels an order.
	CancelOrder(ctx context.Context, symbol string, orderID string) error

	// GetOpenOrders returns all open orders for a symbol.
	GetOpenOrders(ctx context.Context, symbol string) ([]Order, error)
}

// Ticker represents a price ticker.
type Ticker struct {
	Symbol    string  `json:"symbol"`
	LastPrice float64 `json:"last_price"`
	BidPrice  float64 `json:"bid_price"`
	AskPrice  float64 `json:"ask_price"`
	Volume24h float64 `json:"volume_24h"`
	Change24h float64 `json:"change_24h"`
}

// Balance represents an asset balance.
type Balance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Locked float64 `json:"locked"`
	Total  float64 `json:"total"`
}

// OrderBook represents the order book.
type OrderBook struct {
	Symbol string     `json:"symbol"`
	Bids   []OrderRow `json:"bids"`
	Asks   []OrderRow `json:"asks"`
}

// OrderRow represents a single order book entry.
type OrderRow struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// OrderParams represents the parameters for placing an order.
type OrderParams struct {
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`        // "BUY" or "SELL"
	Type        string  `json:"type"`        // "MARKET" or "LIMIT"
	Quantity    float64 `json:"quantity"`
	Price       float64 `json:"price"`       // Required for LIMIT orders
	TimeInForce string  `json:"time_in_force"` // "GTC", "IOC", "FOK"
}

// Order represents an order.
type Order struct {
	OrderID       string  `json:"order_id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Type          string  `json:"type"`
	Status        string  `json:"status"`      // "NEW", "FILLED", "PARTIALLY_FILLED", "CANCELED", "REJECTED"
	Price         float64 `json:"price"`
	Quantity      float64 `json:"quantity"`
	ExecutedQty   float64 `json:"executed_qty"`
	CumulativeQty float64 `json:"cumulative_qty"`
	TimeInForce   string  `json:"time_in_force"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

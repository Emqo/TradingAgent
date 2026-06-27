package exchange

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Emqo/TradingAgent/internal/retry"
)

// BinanceExchange implements the Exchange interface for Binance.
type BinanceExchange struct {
	apiKey     string
	apiSecret  string
	baseURL    string
	httpClient *http.Client
	retryConfig retry.Config
}

// NewBinanceExchange creates a new Binance exchange client.
func NewBinanceExchange(apiKey, apiSecret string, testnet bool) *BinanceExchange {
	baseURL := "https://api.binance.com"
	if testnet {
		baseURL = "https://testnet.binance.vision"
	}

	return &BinanceExchange{
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		retryConfig: retry.Config{
			MaxRetries:        3,
			InitialBackoff:    500 * time.Millisecond,
			MaxBackoff:        5 * time.Second,
			BackoffMultiplier: 2.0,
		},
	}
}

// Name returns the exchange name.
func (e *BinanceExchange) Name() string {
	return "binance"
}

// GetTicker returns the current price for a trading pair.
func (e *BinanceExchange) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	path := fmt.Sprintf("/api/v3/ticker/price?symbol=%s", symbol)

	result, err := retry.Do(ctx, e.retryConfig, func() (*Ticker, error) {
		resp, err := e.doRequest(ctx, "GET", path, nil, false)
		if err != nil {
			return nil, err
		}

		var apiResult struct {
			Symbol string `json:"symbol"`
			Price  string `json:"price"`
		}
		if err := json.Unmarshal(resp, &apiResult); err != nil {
			return nil, fmt.Errorf("parse ticker: %w", err)
		}

		price, err := strconv.ParseFloat(apiResult.Price, 64)
		if err != nil {
			return nil, fmt.Errorf("parse price: %w", err)
		}

		return &Ticker{
			Symbol:    symbol,
			LastPrice: price,
		}, nil
	})

	return result, err
}

// GetBalance returns the account balance.
func (e *BinanceExchange) GetBalance(ctx context.Context) (map[string]Balance, error) {
	resp, err := e.doRequest(ctx, "GET", "/api/v3/account", nil, true)
	if err != nil {
		return nil, err
	}

	var account struct {
		Balances []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
	}
	if err := json.Unmarshal(resp, &account); err != nil {
		return nil, fmt.Errorf("parse account: %w", err)
	}

	balances := make(map[string]Balance)
	for _, b := range account.Balances {
		free, _ := strconv.ParseFloat(b.Free, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)

		if free > 0 || locked > 0 {
			balances[b.Asset] = Balance{
				Asset:  b.Asset,
				Free:   free,
				Locked: locked,
				Total:  free + locked,
			}
		}
	}

	return balances, nil
}

// GetOrderBook returns the order book depth for a symbol.
func (e *BinanceExchange) GetOrderBook(ctx context.Context, symbol string, depth int) (*OrderBook, error) {
	path := fmt.Sprintf("/api/v3/depth?symbol=%s&limit=%d", symbol, depth)
	resp, err := e.doRequest(ctx, "GET", path, nil, false)
	if err != nil {
		return nil, err
	}

	var book struct {
		Bids [][]string `json:"bids"`
		Asks [][]string `json:"asks"`
	}
	if err := json.Unmarshal(resp, &book); err != nil {
		return nil, fmt.Errorf("parse order book: %w", err)
	}

	orderBook := &OrderBook{Symbol: symbol}

	for _, bid := range book.Bids {
		price, _ := strconv.ParseFloat(bid[0], 64)
		qty, _ := strconv.ParseFloat(bid[1], 64)
		orderBook.Bids = append(orderBook.Bids, OrderRow{Price: price, Quantity: qty})
	}

	for _, ask := range book.Asks {
		price, _ := strconv.ParseFloat(ask[0], 64)
		qty, _ := strconv.ParseFloat(ask[1], 64)
		orderBook.Asks = append(orderBook.Asks, OrderRow{Price: price, Quantity: qty})
	}

	return orderBook, nil
}

// doRequest makes an HTTP request to the Binance API.
func (e *BinanceExchange) doRequest(ctx context.Context, method, path string, params url.Values, signed bool) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}

	if signed {
		params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
		params.Set("recvWindow", "5000")
		params.Set("signature", e.sign(params.Encode()))
	}

	reqURL := e.baseURL + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", e.apiKey)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// PlaceOrder places a new order on Binance.
func (e *BinanceExchange) PlaceOrder(ctx context.Context, params OrderParams) (*Order, error) {
	// Build query parameters
	values := url.Values{}
	values.Set("symbol", params.Symbol)
	values.Set("side", params.Side)
	values.Set("type", params.Type)
	values.Set("quantity", formatFloat(params.Quantity))

	if params.Type == "LIMIT" {
		if params.Price <= 0 {
			return nil, fmt.Errorf("price is required for LIMIT orders")
		}
		values.Set("price", formatFloat(params.Price))
		if params.TimeInForce == "" {
			values.Set("timeInForce", "GTC") // Good Till Cancel
		} else {
			values.Set("timeInForce", params.TimeInForce)
		}
	}

	// Place the order
	resp, err := e.doRequest(ctx, "POST", "/api/v3/order", values, true)
	if err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}

	// Parse response
	var result struct {
		OrderID       int64  `json:"orderId"`
		Symbol        string `json:"symbol"`
		Side          string `json:"side"`
		Type          string `json:"type"`
		Status        string `json:"status"`
		Price         string `json:"price"`
		Quantity      string `json:"origQty"`
		ExecutedQty   string `json:"executedQty"`
		CumulativeQty string `json:"cummulativeQuoteQty"`
		TimeInForce   string `json:"timeInForce"`
		TransactTime  int64  `json:"transactTime"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parse order response: %w", err)
	}

	// Convert to Order
	order := &Order{
		OrderID:       strconv.FormatInt(result.OrderID, 10),
		Symbol:        result.Symbol,
		Side:          result.Side,
		Type:          result.Type,
		Status:        result.Status,
		Price:         parseFloat(result.Price),
		Quantity:      parseFloat(result.Quantity),
		ExecutedQty:   parseFloat(result.ExecutedQty),
		CumulativeQty: parseFloat(result.CumulativeQty),
		TimeInForce:   result.TimeInForce,
		CreatedAt:     formatTimestamp(result.TransactTime),
		UpdatedAt:     formatTimestamp(result.TransactTime),
	}

	return order, nil
}

// GetOrder returns the status of an order.
func (e *BinanceExchange) GetOrder(ctx context.Context, symbol string, orderID string) (*Order, error) {
	values := url.Values{}
	values.Set("symbol", symbol)
	values.Set("orderId", orderID)

	resp, err := e.doRequest(ctx, "GET", "/api/v3/order", values, true)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	var result struct {
		OrderID       int64  `json:"orderId"`
		Symbol        string `json:"symbol"`
		Side          string `json:"side"`
		Type          string `json:"type"`
		Status        string `json:"status"`
		Price         string `json:"price"`
		Quantity      string `json:"origQty"`
		ExecutedQty   string `json:"executedQty"`
		CumulativeQty string `json:"cummulativeQuoteQty"`
		TimeInForce   string `json:"timeInForce"`
		Time          int64  `json:"time"`
		UpdateTime    int64  `json:"updateTime"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parse order: %w", err)
	}

	order := &Order{
		OrderID:       strconv.FormatInt(result.OrderID, 10),
		Symbol:        result.Symbol,
		Side:          result.Side,
		Type:          result.Type,
		Status:        result.Status,
		Price:         parseFloat(result.Price),
		Quantity:      parseFloat(result.Quantity),
		ExecutedQty:   parseFloat(result.ExecutedQty),
		CumulativeQty: parseFloat(result.CumulativeQty),
		TimeInForce:   result.TimeInForce,
		CreatedAt:     formatTimestamp(result.Time),
		UpdatedAt:     formatTimestamp(result.UpdateTime),
	}

	return order, nil
}

// CancelOrder cancels an order.
func (e *BinanceExchange) CancelOrder(ctx context.Context, symbol string, orderID string) error {
	values := url.Values{}
	values.Set("symbol", symbol)
	values.Set("orderId", orderID)

	_, err := e.doRequest(ctx, "DELETE", "/api/v3/order", values, true)
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}

	return nil
}

// GetOpenOrders returns all open orders for a symbol.
func (e *BinanceExchange) GetOpenOrders(ctx context.Context, symbol string) ([]Order, error) {
	values := url.Values{}
	if symbol != "" {
		values.Set("symbol", symbol)
	}

	resp, err := e.doRequest(ctx, "GET", "/api/v3/openOrders", values, true)
	if err != nil {
		return nil, fmt.Errorf("get open orders: %w", err)
	}

	var results []struct {
		OrderID       int64  `json:"orderId"`
		Symbol        string `json:"symbol"`
		Side          string `json:"side"`
		Type          string `json:"type"`
		Status        string `json:"status"`
		Price         string `json:"price"`
		Quantity      string `json:"origQty"`
		ExecutedQty   string `json:"executedQty"`
		CumulativeQty string `json:"cummulativeQuoteQty"`
		TimeInForce   string `json:"timeInForce"`
		Time          int64  `json:"time"`
		UpdateTime    int64  `json:"updateTime"`
	}
	if err := json.Unmarshal(resp, &results); err != nil {
		return nil, fmt.Errorf("parse orders: %w", err)
	}

	orders := make([]Order, 0, len(results))
	for _, r := range results {
		orders = append(orders, Order{
			OrderID:       strconv.FormatInt(r.OrderID, 10),
			Symbol:        r.Symbol,
			Side:          r.Side,
			Type:          r.Type,
			Status:        r.Status,
			Price:         parseFloat(r.Price),
			Quantity:      parseFloat(r.Quantity),
			ExecutedQty:   parseFloat(r.ExecutedQty),
			CumulativeQty: parseFloat(r.CumulativeQty),
			TimeInForce:   r.TimeInForce,
			CreatedAt:     formatTimestamp(r.Time),
			UpdatedAt:     formatTimestamp(r.UpdateTime),
		})
	}

	return orders, nil
}

// sign creates a HMAC SHA256 signature.
func (e *BinanceExchange) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(e.apiSecret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// Helper functions

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func formatTimestamp(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).Format(time.RFC3339)
}

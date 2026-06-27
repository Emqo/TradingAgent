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
)

// BinanceExchange implements the Exchange interface for Binance.
type BinanceExchange struct {
	apiKey     string
	apiSecret  string
	baseURL    string
	httpClient *http.Client
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
	}
}

// Name returns the exchange name.
func (e *BinanceExchange) Name() string {
	return "binance"
}

// GetTicker returns the current price for a trading pair.
func (e *BinanceExchange) GetTicker(ctx context.Context, symbol string) (*Ticker, error) {
	path := fmt.Sprintf("/api/v3/ticker/price?symbol=%s", symbol)
	resp, err := e.doRequest(ctx, "GET", path, nil, false)
	if err != nil {
		return nil, err
	}

	var result struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("parse ticker: %w", err)
	}

	price, err := strconv.ParseFloat(result.Price, 64)
	if err != nil {
		return nil, fmt.Errorf("parse price: %w", err)
	}

	return &Ticker{
		Symbol:    symbol,
		LastPrice: price,
	}, nil
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

// sign creates a HMAC SHA256 signature.
func (e *BinanceExchange) sign(payload string) string {
	mac := hmac.New(sha256.New, []byte(e.apiSecret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

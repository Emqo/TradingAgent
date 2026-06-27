package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// WebSocketClient manages WebSocket connections to Binance.
type WebSocketClient struct {
	baseURL    string
	testnet    bool
	conn       *websocket.Conn
	mu         sync.RWMutex
	tickers    map[string]*Ticker
	subscribers map[string][]chan *Ticker
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWebSocketClient creates a new WebSocket client.
func NewWebSocketClient(testnet bool) *WebSocketClient {
	ctx, cancel := context.WithCancel(context.Background())

	baseURL := "wss://stream.binance.com:9443"
	if testnet {
		baseURL = "wss://testnet.binance.vision"
	}

	return &WebSocketClient{
		baseURL:     baseURL,
		testnet:     testnet,
		tickers:     make(map[string]*Ticker),
		subscribers: make(map[string][]chan *Ticker),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Connect establishes a WebSocket connection.
func (c *WebSocketClient) Connect() error {
	url := fmt.Sprintf("%s/ws/btcusdt@ticker", c.baseURL)

	conn, _, err := websocket.Dial(c.ctx, url, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	log.Println("✅ WebSocket connected")

	// Start reading messages
	go c.readMessages()

	return nil
}

// Subscribe subscribes to ticker updates for a symbol.
func (c *WebSocketClient) Subscribe(symbol string, ch chan *Ticker) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.subscribers[symbol] = append(c.subscribers[symbol], ch)
}

// GetTicker returns the current ticker for a symbol.
func (c *WebSocketClient) GetTicker(symbol string) (*Ticker, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ticker, ok := c.tickers[symbol]
	return ticker, ok
}

// Close closes the WebSocket connection.
func (c *WebSocketClient) Close() {
	c.cancel()
	if c.conn != nil {
		c.conn.Close(websocket.StatusNormalClosure, "closing")
	}
}

// readMessages reads messages from the WebSocket.
func (c *WebSocketClient) readMessages() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		_, data, err := c.conn.Read(c.ctx)
		if err != nil {
			if c.ctx.Err() != nil {
				return
			}
			log.Printf("⚠️ WebSocket read error: %v", err)
			// Try to reconnect
			time.Sleep(5 * time.Second)
			if err := c.reconnect(); err != nil {
				log.Printf("❌ WebSocket reconnect failed: %v", err)
			}
			continue
		}

		// Parse message
		var msg struct {
			Symbol    string `json:"s"`
			Price     string `json:"c"`
			BidPrice  string `json:"b"`
			AskPrice  string `json:"a"`
			Volume    string `json:"v"`
			Change    string `json:"P"`
		}

		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("⚠️ WebSocket parse error: %v", err)
			continue
		}

		// Update ticker
		ticker := &Ticker{
			Symbol:    msg.Symbol,
			LastPrice: parseFloat(msg.Price),
			BidPrice:  parseFloat(msg.BidPrice),
			AskPrice:  parseFloat(msg.AskPrice),
			Volume24h: parseFloat(msg.Volume),
			Change24h: parseFloat(msg.Change),
		}

		c.mu.Lock()
		c.tickers[msg.Symbol] = ticker

		// Notify subscribers
		if subs, ok := c.subscribers[msg.Symbol]; ok {
			for _, ch := range subs {
				select {
				case ch <- ticker:
				default:
					// Skip if channel is full
				}
			}
		}
		c.mu.Unlock()
	}
}

// reconnect attempts to reconnect to the WebSocket.
func (c *WebSocketClient) reconnect() error {
	if c.conn != nil {
		c.conn.Close(websocket.StatusNormalClosure, "reconnecting")
	}

	url := fmt.Sprintf("%s/ws/btcusdt@ticker", c.baseURL)
	conn, _, err := websocket.Dial(c.ctx, url, nil)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	log.Println("✅ WebSocket reconnected")
	return nil
}

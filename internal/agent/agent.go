package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/llm"
	"github.com/Emqo/TradingAgent/internal/tools"
)

// Agent represents the trading agent.
type Agent struct {
	llm      llm.Provider
	exchange exchange.Exchange
	registry *tools.Registry
	config   Config
}

// Config holds the agent configuration.
type Config struct {
	Interval    time.Duration
	MaxTokens   int
	Temperature float64
}

// New creates a new Agent.
func New(llmProvider llm.Provider, exchangeProvider exchange.Exchange, registry *tools.Registry, cfg Config) *Agent {
	return &Agent{
		llm:      llmProvider,
		exchange: exchangeProvider,
		registry: registry,
		config:   cfg,
	}
}

// Run starts the agent's decision loop.
func (a *Agent) Run(ctx context.Context) error {
	log.Println("🤖 Agent started")

	ticker := time.NewTicker(a.config.Interval)
	defer ticker.Stop()

	// Run immediately on start
	if err := a.decide(ctx); err != nil {
		log.Printf("❌ Decision error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Agent stopped")
			return nil
		case <-ticker.C:
			if err := a.decide(ctx); err != nil {
				log.Printf("❌ Decision error: %v", err)
			}
		}
	}
}

// decide performs one iteration of the Observe → Think → Act loop.
func (a *Agent) decide(ctx context.Context) error {
	log.Println("---")
	log.Println("🔄 Starting decision cycle...")

	// Step 1: Observe - gather initial market data
	log.Println("📊 Observing market data...")
	observation, err := a.observe(ctx)
	if err != nil {
		return fmt.Errorf("observe: %w", err)
	}
	log.Printf("   BTC: $%.2f", observation.BTCPrice)

	// Step 2: Think - send to LLM with tools
	log.Println("🤔 Thinking...")
	response, err := a.think(ctx, observation)
	if err != nil {
		return fmt.Errorf("think: %w", err)
	}

	// Step 3: Handle tool calls if any
	if len(response.ToolCalls) > 0 {
		log.Printf("🔧 Executing %d tool calls...", len(response.ToolCalls))
		if err := a.handleToolCalls(ctx, response.ToolCalls); err != nil {
			return fmt.Errorf("handle tool calls: %w", err)
		}
	}

	// Step 4: Log the analysis
	if response.Content != "" {
		log.Printf("   Analysis: %s", response.Content)
	}

	log.Println("📝 Decision logged (no real trading yet)")

	return nil
}

// Observation holds the market data observed.
type Observation struct {
	BTCPrice  float64
	Timestamp time.Time
}

// observe gathers market data.
func (a *Agent) observe(ctx context.Context) (*Observation, error) {
	ticker, err := a.exchange.GetTicker(ctx, "BTCUSDT")
	if err != nil {
		return nil, err
	}

	return &Observation{
		BTCPrice:  ticker.LastPrice,
		Timestamp: time.Now(),
	}, nil
}

// think sends the observation to the LLM for analysis.
func (a *Agent) think(ctx context.Context, obs *Observation) (*llm.Response, error) {
	messages := []llm.Message{
		{
			Role: "system",
			Content: `You are a crypto trading analyst with access to market data tools.

Your capabilities:
- Get real-time prices for any trading pair
- View order book depth
- Check account balance
- Detect arbitrage opportunities

When analyzing the market:
1. Use tools to gather additional data if needed
2. Identify potential trading opportunities
3. Assess risk factors
4. Provide actionable recommendations

Be concise and focused on actionable insights.`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(
				"Current BTC/USDT price: $%.2f\nTime: %s\n\nAnalyze the market and identify any trading opportunities. Use tools if you need more data.",
				obs.BTCPrice,
				obs.Timestamp.Format(time.RFC3339),
			),
		},
	}

	// Convert tools to LLM format
	llmTools := a.registry.ToLLMTools()

	resp, err := a.llm.ChatWithTools(ctx, messages, llmTools, llm.WithMaxTokens(a.config.MaxTokens))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// handleToolCalls executes tool calls from the LLM.
func (a *Agent) handleToolCalls(ctx context.Context, toolCalls []llm.ToolCall) error {
	for _, tc := range toolCalls {
		log.Printf("   🔧 Tool: %s", tc.Name)

		// Get the tool
		tool, err := a.registry.Get(tc.Name)
		if err != nil {
			log.Printf("   ❌ Tool not found: %v", err)
			continue
		}

		// Parse arguments
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
			log.Printf("   ❌ Invalid arguments: %v", err)
			continue
		}

		// Execute the tool
		result, err := tool.Execute(ctx, args)
		if err != nil {
			log.Printf("   ❌ Execution error: %v", err)
			continue
		}

		// Log result
		if result.Success {
			log.Printf("   ✅ Success: %v", result.Data)
		} else {
			log.Printf("   ⚠️ Error: %s", result.Error)
		}
	}

	return nil
}

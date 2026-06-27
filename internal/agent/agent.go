package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/llm"
)

// Agent represents the trading agent.
type Agent struct {
	llm      llm.Provider
	exchange exchange.Exchange
	config   Config
}

// Config holds the agent configuration.
type Config struct {
	Interval    time.Duration
	MaxTokens   int
	Temperature float64
}

// New creates a new Agent.
func New(llmProvider llm.Provider, exchangeProvider exchange.Exchange, cfg Config) *Agent {
	return &Agent{
		llm:      llmProvider,
		exchange: exchangeProvider,
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

	// Step 1: Observe
	log.Println("📊 Observing market data...")
	observation, err := a.observe(ctx)
	if err != nil {
		return fmt.Errorf("observe: %w", err)
	}
	log.Printf("   BTC: $%.2f", observation.BTCPrice)

	// Step 2: Think
	log.Println("🤔 Thinking...")
	decision, err := a.think(ctx, observation)
	if err != nil {
		return fmt.Errorf("think: %w", err)
	}
	log.Printf("   Analysis: %s", decision)

	// Step 3: Act (for now, just log)
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
func (a *Agent) think(ctx context.Context, obs *Observation) (string, error) {
	messages := []llm.Message{
		{
			Role: "system",
			Content: `You are a crypto trading analyst. Analyze the market data and provide a brief assessment.
Focus on:
1. Current price level and trend
2. Potential trading opportunities
3. Risk factors

Be concise. One paragraph max.`,
		},
		{
			Role: "user",
			Content: fmt.Sprintf(
				"Current BTC/USDT price: $%.2f\nTime: %s\n\nGive your analysis.",
				obs.BTCPrice,
				obs.Timestamp.Format(time.RFC3339),
			),
		},
	}

	resp, err := a.llm.Chat(ctx, messages, llm.WithMaxTokens(a.config.MaxTokens))
	if err != nil {
		return "", err
	}

	return resp.Content, nil
}

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Emqo/TradingAgent/internal/arbitrage"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/llm"
	"github.com/Emqo/TradingAgent/internal/tools"
)

// Agent represents the trading agent.
type Agent struct {
	llm       llm.Provider
	exchange  exchange.Exchange
	registry  *tools.Registry
	arbitrage *arbitrage.Manager
	session   *llm.Session
	config    Config
}

// Config holds the agent configuration.
type Config struct {
	Interval    time.Duration
	MaxTokens   int
	Temperature float64
}

// New creates a new Agent.
func New(
	llmProvider llm.Provider,
	exchangeProvider exchange.Exchange,
	registry *tools.Registry,
	arbitrageManager *arbitrage.Manager,
	cfg Config,
) *Agent {
	// Create session with system prompt
	systemPrompt := `You are a crypto trading analyst with access to market data tools and arbitrage detection.

Your capabilities:
- Get real-time prices for any trading pair
- View order book depth
- Check account balance
- Detect arbitrage opportunities
- Check risk status
- Generate trading strategies
- Manage memory

When analyzing the market:
1. Consider any detected arbitrage opportunities
2. Assess risk factors
3. Provide actionable recommendations

Be concise and focused on actionable insights.`

	session := llm.NewSession(systemPrompt, 20) // Keep last 20 messages

	return &Agent{
		llm:       llmProvider,
		exchange:  exchangeProvider,
		registry:  registry,
		arbitrage: arbitrageManager,
		session:   session,
		config:    cfg,
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

	// Step 2: Scan for arbitrage opportunities
	log.Println("🔍 Scanning for arbitrage opportunities...")
	arbResult, err := a.arbitrage.Scan(ctx)
	if err != nil {
		log.Printf("   ⚠️ Arbitrage scan error: %v", err)
	} else {
		log.Printf("   Found %d triangular, %d cash-and-carry opportunities",
			len(arbResult.TriangularOpportunities),
			len(arbResult.CashAndCarryOpportunities))
	}

	// Step 3: Think - send to LLM with tools
	log.Println("🤔 Thinking...")
	response, err := a.think(ctx, observation, arbResult)
	if err != nil {
		return fmt.Errorf("think: %w", err)
	}

	// Step 4: Handle tool calls if any
	if len(response.ToolCalls) > 0 {
		log.Printf("🔧 Executing %d tool calls...", len(response.ToolCalls))
		if err := a.handleToolCalls(ctx, response.ToolCalls); err != nil {
			return fmt.Errorf("handle tool calls: %w", err)
		}
	}

	// Step 5: Log the analysis
	if response.Content != "" {
		log.Printf("   Analysis: %s", response.Content)
	}

	// Step 6: Store decision in session
	a.session.AddMessage(llm.Message{
		Role:    "user",
		Content: fmt.Sprintf("BTC: $%.2f, Arbitrage: %d triangular, %d cash-and-carry",
			observation.BTCPrice,
			len(arbResult.TriangularOpportunities),
			len(arbResult.CashAndCarryOpportunities)),
	})
	a.session.AddMessage(llm.Message{
		Role:    "assistant",
		Content: response.Content,
	})

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
func (a *Agent) think(ctx context.Context, obs *Observation, arbResult *arbitrage.ScanResult) (*llm.Response, error) {
	// Build arbitrage context
	arbContext := ""
	if arbResult != nil {
		if len(arbResult.TriangularOpportunities) > 0 {
			arbContext += "\n\nTriangular Arbitrage Opportunities:\n"
			for _, opp := range arbResult.TriangularOpportunities {
				arbContext += fmt.Sprintf("- %s: spread %.2f bps, profit $%.2f\n",
					opp.Path.Name, opp.Spread, opp.Profit)
			}
		}
		if len(arbResult.CashAndCarryOpportunities) > 0 {
			arbContext += "\n\nCash-and-Carry Opportunities:\n"
			for _, opp := range arbResult.CashAndCarryOpportunities {
				arbContext += fmt.Sprintf("- %s: annualized %.2f%%, basis %.4f%%\n",
					opp.Symbol, opp.AnnualizedYield, opp.BasisPercent)
			}
		}
	}

	// Add user message to session
	userMessage := fmt.Sprintf(
		"Current BTC/USDT price: $%.2f\nTime: %s%s\n\nAnalyze the market and identify any trading opportunities. Use tools if you need more data.",
		obs.BTCPrice,
		obs.Timestamp.Format(time.RFC3339),
		arbContext,
	)

	a.session.AddMessage(llm.Message{
		Role:    "user",
		Content: userMessage,
	})

	// Get all messages from session
	messages := a.session.GetMessages()

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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Emqo/TradingAgent/internal/arbitrage"
	"github.com/Emqo/TradingAgent/internal/database"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/llm"
	"github.com/Emqo/TradingAgent/internal/logger"
	"github.com/Emqo/TradingAgent/internal/metrics"
	"github.com/Emqo/TradingAgent/internal/tools"
)

// Agent represents the trading agent.
type Agent struct {
	llm       llm.Provider
	exchange  exchange.Exchange
	registry  *tools.Registry
	arbitrage *arbitrage.Manager
	session   *llm.Session
	metrics   *metrics.Metrics
	logger    *logger.Logger
	db        *database.DB
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
	metricsInstance *metrics.Metrics,
	log *logger.Logger,
	db *database.DB,
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
		metrics:   metricsInstance,
		logger:    log,
		db:        db,
		config:    cfg,
	}
}

// Run starts the agent's decision loop.
func (a *Agent) Run(ctx context.Context) error {
	a.logger.Info("Agent started")

	ticker := time.NewTicker(a.config.Interval)
	defer ticker.Stop()

	// Run immediately on start
	if err := a.decide(ctx); err != nil {
		a.logger.Errorf("Decision error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Agent stopped")
			return nil
		case <-ticker.C:
			if err := a.decide(ctx); err != nil {
				a.logger.Errorf("Decision error: %v", err)
			}
		}
	}
}

// decide performs one iteration of the Observe → Think → Act loop.
func (a *Agent) decide(ctx context.Context) error {
	a.logger.Debug("Starting decision cycle")

	// Step 1: Observe - gather initial market data
	a.logger.Debug("Observing market data")
	observation, err := a.observe(ctx)
	if err != nil {
		return fmt.Errorf("observe: %w", err)
	}
	a.logger.WithField("btc_price", observation.BTCPrice).Info("Market data observed")

	// Step 2: Scan for arbitrage opportunities
	a.logger.Debug("Scanning for arbitrage opportunities")
	arbResult, err := a.arbitrage.Scan(ctx)
	if err != nil {
		a.logger.Warnf("Arbitrage scan error: %v", err)
	} else {
		a.logger.WithFields(map[string]any{
			"triangular":    len(arbResult.TriangularOpportunities),
			"cash_and_carry": len(arbResult.CashAndCarryOpportunities),
		}).Info("Arbitrage scan completed")

		// Record arbitrage metrics
		for _, opp := range arbResult.TriangularOpportunities {
			a.metrics.RecordArbitrageOpportunity("triangular")
			a.metrics.RecordArbitrageSpread("triangular", opp.Spread)
		}
		for range arbResult.CashAndCarryOpportunities {
			a.metrics.RecordArbitrageOpportunity("cash_and_carry")
		}
	}

	// Step 3: Think - send to LLM with tools
	a.logger.Debug("Thinking")
	startTime := time.Now()
	response, err := a.think(ctx, observation, arbResult)
	llmLatency := time.Since(startTime).Seconds()

	if err != nil {
		a.metrics.RecordLLMCall(a.llm.Name(), "error")
		return fmt.Errorf("think: %w", err)
	}

	a.metrics.RecordLLMCall(a.llm.Name(), "success")
	a.metrics.RecordLLMLatency(a.llm.Name(), llmLatency)
	a.metrics.RecordLLMTokens(a.llm.Name(), "total", response.TokenUsage.TotalTokens)

	a.logger.WithFields(map[string]any{
		"latency_seconds": llmLatency,
		"tokens":          response.TokenUsage.TotalTokens,
		"tool_calls":      len(response.ToolCalls),
	}).Info("LLM response received")

	// Step 4: Handle tool calls if any
	if len(response.ToolCalls) > 0 {
		a.logger.WithField("count", len(response.ToolCalls)).Info("Executing tool calls")
		if err := a.handleToolCalls(ctx, response.ToolCalls); err != nil {
			return fmt.Errorf("handle tool calls: %w", err)
		}
	}

	// Step 5: Log the analysis
	if response.Content != "" {
		a.logger.WithField("analysis", response.Content).Info("Analysis completed")
	}

	// Step 6: Store decision in session
	a.session.AddMessage(llm.Message{
		Role: "user",
		Content: fmt.Sprintf("BTC: $%.2f, Arbitrage: %d triangular, %d cash-and-carry",
			observation.BTCPrice,
			len(arbResult.TriangularOpportunities),
			len(arbResult.CashAndCarryOpportunities)),
	})
	a.session.AddMessage(llm.Message{
		Role:    "assistant",
		Content: response.Content,
	})

	// Step 7: Store decision in database
	if a.db != nil {
		decision := &database.Decision{
			Action:     "ANALYZE",
			Reason:     response.Content,
			Result:     "Analysis completed",
			TokensUsed: response.TokenUsage.TotalTokens,
			LatencyMs:  int(llmLatency * 1000),
		}

		if err := a.db.InsertDecision(ctx, decision); err != nil {
			a.logger.Warnf("Failed to store decision in database: %v", err)
		} else {
			a.logger.WithField("decision_id", decision.ID).Info("Decision stored in database")
		}
	}

	a.logger.Info("Decision logged")

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
		a.logger.WithField("tool", tc.Name).Debug("Executing tool")

		// Get the tool
		tool, err := a.registry.Get(tc.Name)
		if err != nil {
			a.logger.WithField("tool", tc.Name).Errorf("Tool not found: %v", err)
			continue
		}

		// Parse arguments
		var args map[string]any
		if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
			a.logger.WithField("tool", tc.Name).Errorf("Invalid arguments: %v", err)
			continue
		}

		// Execute the tool
		result, err := tool.Execute(ctx, args)
		if err != nil {
			a.logger.WithField("tool", tc.Name).Errorf("Execution error: %v", err)
			continue
		}

		// Log result
		if result.Success {
			a.logger.WithField("tool", tc.Name).Debug("Tool executed successfully")
		} else {
			a.logger.WithField("tool", tc.Name).Warnf("Tool error: %s", result.Error)
		}
	}

	return nil
}

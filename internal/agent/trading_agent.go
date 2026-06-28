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

// TradingAgent represents the trading-focused agent.
type TradingAgent struct {
	llm       llm.Provider
	exchange  exchange.Exchange
	registry  *tools.Registry
	arbitrage *arbitrage.Manager
	session   *llm.Session
	metrics   *metrics.Metrics
	logger    *logger.Logger
	db        *database.DB
	config    TradingAgentConfig
}

// TradingAgentConfig holds trading agent configuration.
type TradingAgentConfig struct {
	Interval    time.Duration
	MaxTokens   int
	Temperature float64
}

// NewTradingAgent creates a new trading agent.
func NewTradingAgent(
	llmProvider llm.Provider,
	exchangeProvider exchange.Exchange,
	registry *tools.Registry,
	arbitrageManager *arbitrage.Manager,
	metricsInstance *metrics.Metrics,
	log *logger.Logger,
	db *database.DB,
	cfg TradingAgentConfig,
) *TradingAgent {
	// Create session with system prompt
	systemPrompt := `你是一个加密货币交易分析师，可以访问市场数据工具和套利检测。

你的能力：
- 获取任何交易对的实时价格
- 查看订单簿深度
- 查询账户余额
- 检测套利机会
- 检查风险状态
- 生成交易策略
- 管理记忆
- 执行套利交易

分析市场时：
1. 考虑检测到的套利机会
2. 评估风险因素
3. 提供可操作的建议

当发现套利机会时：
1. 分析价差是否足够覆盖手续费
2. 评估执行风险（滑点、延迟）
3. 决定是否执行
4. 如果决定执行，调用 execute_arbitrage 工具

请用中文回复，简洁且聚焦于可操作的洞察。`

	session := llm.NewSession(systemPrompt, 20) // Keep last 20 messages

	return &TradingAgent{
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

// Run starts the trading agent's decision loop.
func (a *TradingAgent) Run(ctx context.Context) error {
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

// decide performs one iteration of the trading decision loop.
func (a *TradingAgent) decide(ctx context.Context) error {
	a.logger.Debug("Starting decision cycle")

	// Step 1: Observe - gather initial market data
	a.logger.Debug("Observing market data")
	observation, err := a.observe(ctx)
	if err != nil {
		return fmt.Errorf("observe: %w", err)
	}
	a.logger.WithField("btc_price", observation.BTCPrice).Info("Market data observed")

	// Step 2: Think - send to LLM with tools
	// Note: TradingAgent focuses on market analysis and trading decisions
	a.logger.Debug("Thinking")
	startTime := time.Now()
	response, err := a.think(ctx, observation)
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
	toolCallNames := make([]string, 0)
	if len(response.ToolCalls) > 0 {
		a.logger.WithField("count", len(response.ToolCalls)).Info("Executing tool calls")
		for _, tc := range response.ToolCalls {
			toolCallNames = append(toolCallNames, tc.Name)
		}
		if err := a.handleToolCalls(ctx, response.ToolCalls); err != nil {
			return fmt.Errorf("handle tool calls: %w", err)
		}
	}

	// Step 5: Parse action from LLM response
	action := a.parseAction(response.Content, toolCallNames)

	// Step 6: Log the analysis
	if response.Content != "" {
		a.logger.WithField("analysis", response.Content).Info("Analysis completed")
	}

	// Step 7: Store decision in session
	a.session.AddMessage(llm.Message{
		Role: "user",
		Content: fmt.Sprintf("BTC: $%.2f, 请分析市场并做出交易决策",
			observation.BTCPrice),
	})
	a.session.AddMessage(llm.Message{
		Role:    "assistant",
		Content: response.Content,
	})

	// Step 8: Store decision in database
	if a.db != nil {
		// Build detailed reason
		reason := response.Content
		if len(toolCallNames) > 0 {
			if reason == "" {
				reason = fmt.Sprintf("使用工具: %v", toolCallNames)
			} else {
				reason += fmt.Sprintf("\n\n使用工具: %v", toolCallNames)
			}
		}

		// If still empty, generate a default reason
		if reason == "" {
			reason = "市场分析完成"
		}

		// Add market context
		reason += fmt.Sprintf("\n\n市场数据:")
		reason += fmt.Sprintf("\n- BTC: $%.2f", observation.BTCPrice)

		decision := &database.Decision{
			Action:     action,
			Symbol:     "BTCUSDT", // Primary symbol
			Reason:     reason,
			Result:     fmt.Sprintf("分析完成，调用 %d 个工具", len(toolCallNames)),
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

// parseAction parses the action from the LLM response.
func (a *TradingAgent) parseAction(content string, toolCalls []string) string {
	// Check if any order tools were called
	for _, tool := range toolCalls {
		switch tool {
		case "place_order":
			return "交易"
		case "cancel_order":
			return "撤单"
		case "check_risk":
			return "风控检查"
		}
	}

	// Parse content for action keywords
	contentLower := content
	if contains(contentLower, "buy") || contains(contentLower, "买入") || contains(contentLower, "long") {
		return "买入信号"
	}
	if contains(contentLower, "sell") || contains(contentLower, "卖出") || contains(contentLower, "short") {
		return "卖出信号"
	}
	if contains(contentLower, "hold") || contains(contentLower, "持有") || contains(contentLower, "wait") {
		return "持有"
	}

	return "分析"
}

// contains checks if a string contains a substring (case-insensitive helper).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Observation holds the market data observed.
type Observation struct {
	BTCPrice  float64
	Timestamp time.Time
}

// observe gathers market data.
func (a *TradingAgent) observe(ctx context.Context) (*Observation, error) {
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
func (a *TradingAgent) think(ctx context.Context, obs *Observation) (*llm.Response, error) {
	// Add user message to session
	userMessage := fmt.Sprintf(
		"当前 BTC/USDT 价格: $%.2f\n时间: %s\n\n请分析市场并做出交易决策。使用工具获取更多数据。",
		obs.BTCPrice,
		obs.Timestamp.Format(time.RFC3339),
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
func (a *TradingAgent) handleToolCalls(ctx context.Context, toolCalls []llm.ToolCall) error {
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

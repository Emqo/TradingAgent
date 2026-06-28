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

// ArbitrageAgent represents the arbitrage-focused agent.
type ArbitrageAgent struct {
	llm       llm.Provider
	exchange  exchange.Exchange
	registry  *tools.Registry
	arbitrage *arbitrage.Manager
	session   *llm.Session
	metrics   *metrics.Metrics
	logger    *logger.Logger
	db        *database.DB
	config    ArbitrageAgentConfig
}

// ArbitrageAgentConfig holds arbitrage agent configuration.
type ArbitrageAgentConfig struct {
	Interval    time.Duration
	MaxTokens   int
	Temperature float64
}

// NewArbitrageAgent creates a new arbitrage agent.
func NewArbitrageAgent(
	llmProvider llm.Provider,
	exchangeProvider exchange.Exchange,
	registry *tools.Registry,
	arbitrageManager *arbitrage.Manager,
	metricsInstance *metrics.Metrics,
	log *logger.Logger,
	db *database.DB,
	cfg ArbitrageAgentConfig,
) *ArbitrageAgent {
	// Create session with system prompt
	systemPrompt := `你是一个加密货币套利专家，专注于检测和执行套利交易。

你的能力：
- 检测三角套利机会
- 检测期现套利机会
- 分析价差是否足够覆盖手续费
- 评估执行风险（滑点、延迟）
- 执行套利交易

分析套利机会时：
1. 计算扣除手续费后的净利润
2. 评估执行风险
3. 考虑当前市场状况
4. 决定是否执行

当发现套利机会时：
1. 分析价差是否足够覆盖手续费
2. 评估执行风险（滑点、延迟）
3. 决定是否执行
4. 如果决定执行，调用 execute_arbitrage 工具

请用中文回复，简洁且聚焦于可操作的洞察。`

	session := llm.NewSession(systemPrompt, 20)

	return &ArbitrageAgent{
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

// Run starts the arbitrage agent's decision loop.
func (a *ArbitrageAgent) Run(ctx context.Context) error {
	a.logger.Info("Arbitrage Agent started")

	ticker := time.NewTicker(a.config.Interval)
	defer ticker.Stop()

	// Run immediately on start
	if err := a.decide(ctx); err != nil {
		a.logger.Errorf("Arbitrage decision error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Arbitrage Agent stopped")
			return nil
		case <-ticker.C:
			if err := a.decide(ctx); err != nil {
				a.logger.Errorf("Arbitrage decision error: %v", err)
			}
		}
	}
}

// decide performs one iteration of the arbitrage decision loop.
func (a *ArbitrageAgent) decide(ctx context.Context) error {
	a.logger.Debug("Starting arbitrage decision cycle")

	// Step 1: Scan for arbitrage opportunities
	a.logger.Debug("Scanning for arbitrage opportunities")
	arbResult, err := a.arbitrage.Scan(ctx)
	if err != nil {
		return fmt.Errorf("arbitrage scan: %w", err)
	}

	totalOpportunities := len(arbResult.TriangularOpportunities) + len(arbResult.CashAndCarryOpportunities)
	a.logger.WithFields(map[string]any{
		"triangular":     len(arbResult.TriangularOpportunities),
		"cash_and_carry": len(arbResult.CashAndCarryOpportunities),
	}).Info("Arbitrage scan completed")

	// If no opportunities, skip
	if totalOpportunities == 0 {
		a.logger.Debug("No arbitrage opportunities found")
		return nil
	}

	// Step 2: Think - send to LLM for analysis
	a.logger.Debug("Analyzing arbitrage opportunities")
	startTime := time.Now()
	response, err := a.think(ctx, arbResult)
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

	// Step 3: Handle tool calls if any
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

	// Step 4: Parse action from LLM response
	action := a.parseAction(response.Content, toolCallNames)

	// Step 5: Log the analysis
	if response.Content != "" {
		a.logger.WithField("analysis", response.Content).Info("Arbitrage analysis completed")
	}

	// Step 6: Store decision in database
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

		if reason == "" {
			reason = "套利分析完成"
		}

		// Add arbitrage opportunity details
		if totalOpportunities > 0 {
			reason += fmt.Sprintf("\n\n发现 %d 个套利机会:", totalOpportunities)
			for _, opp := range arbResult.TriangularOpportunities {
				reason += fmt.Sprintf("\n- %s: 价差 %.2f bps", opp.Path.Name, opp.Spread)
			}
			for _, opp := range arbResult.CashAndCarryOpportunities {
				reason += fmt.Sprintf("\n- %s: 年化 %.2f%%", opp.Symbol, opp.AnnualizedYield)
			}
		}

		decision := &database.Decision{
			Action:     action,
			Symbol:     "ARBITRAGE",
			Reason:     reason,
			Result:     fmt.Sprintf("套利分析完成，调用 %d 个工具", len(toolCallNames)),
			TokensUsed: response.TokenUsage.TotalTokens,
			LatencyMs:  int(llmLatency * 1000),
		}

		if err := a.db.InsertDecision(ctx, decision); err != nil {
			a.logger.Warnf("Failed to store arbitrage decision: %v", err)
		} else {
			a.logger.WithField("decision_id", decision.ID).Info("Arbitrage decision stored")
		}
	}

	a.logger.Info("Arbitrage decision logged")

	return nil
}

// think sends the arbitrage opportunities to the LLM for analysis.
func (a *ArbitrageAgent) think(ctx context.Context, arbResult *arbitrage.ScanResult) (*llm.Response, error) {
	// Build arbitrage context
	arbContext := ""
	if len(arbResult.TriangularOpportunities) > 0 {
		arbContext += "三角套利机会:\n"
		for _, opp := range arbResult.TriangularOpportunities {
			arbContext += fmt.Sprintf("- %s: 价差 %.2f bps, 预计收益 $%.2f\n",
				opp.Path.Name, opp.Spread, opp.Profit)
		}
	}
	if len(arbResult.CashAndCarryOpportunities) > 0 {
		arbContext += "\n期现套利机会:\n"
		for _, opp := range arbResult.CashAndCarryOpportunities {
			arbContext += fmt.Sprintf("- %s: 年化 %.2f%%, 基差 %.4f%%\n",
				opp.Symbol, opp.AnnualizedYield, opp.BasisPercent)
		}
	}

	// Add user message to session
	userMessage := fmt.Sprintf("发现 %d 个套利机会:\n%s\n\n请分析这些机会，决定是否执行。如果决定执行，调用 execute_arbitrage 工具。",
		len(arbResult.TriangularOpportunities)+len(arbResult.CashAndCarryOpportunities),
		arbContext)

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
func (a *ArbitrageAgent) handleToolCalls(ctx context.Context, toolCalls []llm.ToolCall) error {
	for _, tc := range toolCalls {
		a.logger.WithField("tool", tc.Name).Debug("Executing arbitrage tool")

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

// parseAction parses the action from the LLM response.
func (a *ArbitrageAgent) parseAction(content string, toolCalls []string) string {
	// Check if any arbitrage tools were called
	for _, tool := range toolCalls {
		switch tool {
		case "execute_arbitrage":
			return "执行套利"
		case "get_arbitrage_opportunities":
			return "检测套利"
		}
	}

	// Parse content for action keywords
	if contains(content, "执行") || contains(content, "买入") || contains(content, "卖出") {
		return "执行套利"
	}
	if contains(content, "跳过") || contains(content, "等待") || contains(content, "观望") {
		return "观望"
	}

	return "分析套利"
}

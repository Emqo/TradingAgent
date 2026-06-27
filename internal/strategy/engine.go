package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Emqo/TradingAgent/internal/llm"
	"github.com/Emqo/TradingAgent/internal/memory"
)

// Engine manages trading strategies.
type Engine struct {
	llm      llm.Provider
	memory   *memory.Manager
	mu       sync.RWMutex
	strategies map[string]*Strategy
	active     *Strategy
	config     Config
}

// Config holds strategy engine configuration.
type Config struct {
	// StrategyTTL is how long a strategy is valid before regeneration.
	StrategyTTL time.Duration `yaml:"strategy_ttl"`
	// MaxTokens is the max tokens for LLM calls.
	MaxTokens int `yaml:"max_tokens"`
}

// Strategy represents a trading strategy.
type Strategy struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Config      StrategyConfig `json:"config"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	Version     int            `json:"version"`
	Performance *Performance   `json:"performance,omitempty"`
}

// StrategyConfig holds the strategy parameters.
type StrategyConfig struct {
	// TradingPairs is the list of trading pairs to monitor.
	TradingPairs []string `json:"trading_pairs"`
	// RiskLevel is the risk level (conservative, moderate, aggressive).
	RiskLevel string `json:"risk_level"`
	// MaxPositionPercent is the max position as % of portfolio.
	MaxPositionPercent float64 `json:"max_position_percent"`
	// EnableTriangular enables triangular arbitrage.
	EnableTriangular bool `json:"enable_triangular"`
	// EnableCashAndCarry enables cash-and-carry arbitrage.
	EnableCashAndCarry bool `json:"enable_cash_and_carry"`
	// MinSpreadBps is the minimum spread in basis points.
	MinSpreadBps float64 `json:"min_spread_bps"`
	// StopLossPercent is the stop loss percentage.
	StopLossPercent float64 `json:"stop_loss_percent"`
	// TakeProfitPercent is the take profit percentage.
	TakeProfitPercent float64 `json:"take_profit_percent"`
}

// Performance tracks strategy performance.
type Performance struct {
	TotalTrades int     `json:"total_trades"`
	WinningTrades int   `json:"winning_trades"`
	LosingTrades int    `json:"losing_trades"`
	TotalPnL    float64 `json:"total_pnl"`
	WinRate     float64 `json:"win_rate"`
	SharpeRatio float64 `json:"sharpe_ratio"`
}

// NewEngine creates a new strategy engine.
func NewEngine(llmProvider llm.Provider, mem *memory.Manager, config Config) *Engine {
	return &Engine{
		llm:        llmProvider,
		memory:     mem,
		strategies: make(map[string]*Strategy),
		config:     config,
	}
}

// GetActiveStrategy returns the current active strategy.
func (e *Engine) GetActiveStrategy() *Strategy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.active
}

// GenerateStrategy generates a new strategy based on market conditions.
func (e *Engine) GenerateStrategy(ctx context.Context, marketContext string) (*Strategy, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	log.Println("🧠 Generating new strategy...")

	// Get memory context
	memContext := e.memory.GetContext()

	// Build prompt
	prompt := fmt.Sprintf(`You are a crypto trading strategy generator. Based on the current market conditions and historical context, generate a trading strategy.

Market Context:
%s

Memory Context:
%s

Generate a JSON strategy configuration with the following structure:
{
  "name": "strategy_name",
  "description": "Brief description of the strategy",
  "config": {
    "trading_pairs": ["BTCUSDT", "ETHUSDT"],
    "risk_level": "conservative|moderate|aggressive",
    "max_position_percent": 10,
    "enable_triangular": true,
    "enable_cash_and_carry": true,
    "min_spread_bps": 15,
    "stop_loss_percent": 2,
    "take_profit_percent": 5
  }
}

Respond with ONLY the JSON, no other text.`, marketContext, memContext)

	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are a trading strategy generator. Output only valid JSON.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := e.llm.Chat(ctx, messages, llm.WithMaxTokens(e.config.MaxTokens))
	if err != nil {
		return nil, fmt.Errorf("generate strategy: %w", err)
	}

	// Parse strategy
	var strategy Strategy
	if err := json.Unmarshal([]byte(resp.Content), &strategy); err != nil {
		return nil, fmt.Errorf("parse strategy: %w", err)
	}

	// Set metadata
	strategy.ID = fmt.Sprintf("strategy_%d", time.Now().Unix())
	strategy.CreatedAt = time.Now()
	strategy.ExpiresAt = time.Now().Add(e.config.StrategyTTL)
	strategy.Version = len(e.strategies) + 1

	// Store strategy
	e.strategies[strategy.ID] = &strategy
	e.active = &strategy

	// Add to memory
	e.memory.AddToLongTerm(memory.MemoryEntry{
		ID:      strategy.ID,
		Type:    memory.MemoryTypeStrategy,
		Content: fmt.Sprintf("Generated strategy: %s", strategy.Name),
		Data: map[string]any{
			"strategy_id": strategy.ID,
			"name":        strategy.Name,
			"risk_level":  strategy.Config.RiskLevel,
		},
		Tags: []string{"strategy", "generated"},
	})

	log.Printf("✅ Strategy generated: %s (ID: %s)", strategy.Name, strategy.ID)

	return &strategy, nil
}

// IsStrategyExpired checks if the current strategy has expired.
func (e *Engine) IsStrategyExpired() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.active == nil {
		return true
	}

	return time.Now().After(e.active.ExpiresAt)
}

// ReflectOnPerformance analyzes past performance and adjusts strategy.
func (e *Engine) ReflectOnPerformance(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	log.Println("🤔 Reflecting on performance...")

	// Get recent trades from memory
	recentTrades := e.memory.GetLongTerm(10)

	// Build reflection prompt
	tradeContext := ""
	for _, entry := range recentTrades {
		if entry.Type == memory.MemoryTypeTrade {
			tradeContext += "- " + entry.Content + "\n"
		}
	}

	if tradeContext == "" {
		tradeContext = "No recent trades recorded."
	}

	prompt := fmt.Sprintf(`Analyze the recent trading performance and suggest improvements.

Recent Trades:
%s

Current Strategy:
%+v

Provide a brief reflection on:
1. What's working well
2. What needs improvement
3. Specific adjustments to make

Be concise.`, tradeContext, e.active)

	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are a trading performance analyst. Be concise and actionable.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	resp, err := e.llm.Chat(ctx, messages, llm.WithMaxTokens(500))
	if err != nil {
		return fmt.Errorf("reflection: %w", err)
	}

	// Store reflection
	e.memory.AddToLongTerm(memory.MemoryEntry{
		ID:      fmt.Sprintf("reflection_%d", time.Now().Unix()),
		Type:    memory.MemoryTypeReflection,
		Content: resp.Content,
		Tags:    []string{"reflection", "performance"},
	})

	log.Println("✅ Performance reflection completed")

	return nil
}

// GetStrategyHistory returns all generated strategies.
func (e *Engine) GetStrategyHistory() []*Strategy {
	e.mu.RLock()
	defer e.mu.RUnlock()

	strategies := make([]*Strategy, 0, len(e.strategies))
	for _, s := range e.strategies {
		strategies = append(strategies, s)
	}

	return strategies
}

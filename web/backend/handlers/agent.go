package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/Emqo/TradingAgent/internal/llm"
	"github.com/gin-gonic/gin"
)

// AgentHandler handles agent-related requests.
type AgentHandler struct {
	mu        sync.RWMutex
	decisions []Decision
	stats     AgentStats
}

// Decision represents an agent decision.
type Decision struct {
	Time    string  `json:"time"`
	Action  string  `json:"action"`
	Reason  string  `json:"reason"`
	Result  string  `json:"result"`
	PnL     float64 `json:"pnl"`
}

// AgentStats represents agent statistics.
type AgentStats struct {
	TodayDecisions int     `json:"today_decisions"`
	TodayTrades    int     `json:"today_trades"`
	TodayPnL       float64 `json:"today_pnl"`
	WinRate        float64 `json:"win_rate"`
	LLMCalls       int     `json:"llm_calls"`
	TokensUsed     int     `json:"tokens_used"`
}

// NewAgentHandler creates a new agent handler.
func NewAgentHandler() *AgentHandler {
	return &AgentHandler{
		decisions: make([]Decision, 0),
		stats: AgentStats{
			TodayDecisions: 0,
			TodayTrades:    0,
			TodayPnL:       0,
			WinRate:        0,
			LLMCalls:       0,
			TokensUsed:     0,
		},
	}
}

// AddDecision adds a new decision to the history.
func (h *AgentHandler) AddDecision(decision Decision) {
	h.mu.Lock()
	defer h.mu.Unlock()

	decision.Time = time.Now().Format("2006-01-02T15:04:05-07:00")
	h.decisions = append(h.decisions, decision)

	// Keep only last 100 decisions
	if len(h.decisions) > 100 {
		h.decisions = h.decisions[len(h.decisions)-100:]
	}

	// Update stats
	h.stats.TodayDecisions++
	if decision.PnL != 0 {
		h.stats.TodayTrades++
		h.stats.TodayPnL += decision.PnL
	}
}

// UpdateLLMStats updates LLM call statistics.
func (h *AgentHandler) UpdateLLMStats(calls int, tokens int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stats.LLMCalls += calls
	h.stats.TokensUsed += tokens
}

// GetDecisions returns the decision history.
func (h *AgentHandler) GetDecisions(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"decisions": h.decisions,
	})
}

// GetStats returns agent statistics.
func (h *AgentHandler) GetStats(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Calculate win rate
	winRate := 0.0
	if h.stats.TodayTrades > 0 {
		// TODO: Calculate actual win rate from trade history
		winRate = 62.0 // Placeholder
	}

	c.JSON(http.StatusOK, gin.H{
		"today_decisions": h.stats.TodayDecisions,
		"today_trades":    h.stats.TodayTrades,
		"today_pnl":       h.stats.TodayPnL,
		"win_rate":        winRate,
		"llm_calls":       h.stats.LLMCalls,
		"tokens_used":     h.stats.TokensUsed,
	})
}

// ResetDailyStats resets daily statistics.
func (h *AgentHandler) ResetDailyStats() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stats.TodayDecisions = 0
	h.stats.TodayTrades = 0
	h.stats.TodayPnL = 0
}

// GetDecisionFromLLMResponse creates a decision from an LLM response.
func (h *AgentHandler) GetDecisionFromLLMResponse(resp *llm.Response, action string, pnl float64) Decision {
	return Decision{
		Time:   time.Now().Format("2006-01-02T15:04:05-07:00"),
		Action: action,
		Reason: resp.Content,
		Result: "已记录",
		PnL:    pnl,
	}
}

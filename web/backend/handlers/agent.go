package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Emqo/TradingAgent/internal/database"
	"github.com/Emqo/TradingAgent/internal/llm"
	"github.com/gin-gonic/gin"
)

// AgentHandler handles agent-related requests.
type AgentHandler struct {
	mu        sync.RWMutex
	db        *database.DB
	stats     AgentStats
}

// Decision represents an agent decision.
type Decision struct {
	ID          int64     `json:"id"`
	Time        string    `json:"time"`
	Action      string    `json:"action"`
	Symbol      string    `json:"symbol"`
	Reason      string    `json:"reason"`
	Result      string    `json:"result"`
	PnL         float64   `json:"pnl"`
	TokensUsed  int       `json:"tokens_used"`
	LatencyMs   int       `json:"latency_ms"`
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
func NewAgentHandler(db *database.DB) *AgentHandler {
	return &AgentHandler{
		db: db,
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

// AddDecision adds a new decision to the database.
func (h *AgentHandler) AddDecision(decision Decision) error {
	if h.db == nil {
		return nil
	}

	ctx := context.Background()
	dbDecision := &database.Decision{
		Action:     decision.Action,
		Symbol:     decision.Symbol,
		Reason:     decision.Reason,
		Result:     decision.Result,
		PnL:        decision.PnL,
		TokensUsed: decision.TokensUsed,
		LatencyMs:  decision.LatencyMs,
	}

	if err := h.db.InsertDecision(ctx, dbDecision); err != nil {
		return err
	}

	decision.ID = dbDecision.ID
	decision.Time = dbDecision.CreatedAt.Format("2006-01-02T15:04:05-07:00")

	// Update stats
	h.mu.Lock()
	h.stats.TodayDecisions++
	if decision.PnL != 0 {
		h.stats.TodayTrades++
		h.stats.TodayPnL += decision.PnL
	}
	h.mu.Unlock()

	return nil
}

// UpdateLLMStats updates LLM call statistics.
func (h *AgentHandler) UpdateLLMStats(calls int, tokens int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stats.LLMCalls += calls
	h.stats.TokensUsed += tokens
}

// GetDecisions returns the decision history from database.
func (h *AgentHandler) GetDecisions(c *gin.Context) {
	ctx := c.Request.Context()

	if h.db == nil {
		c.JSON(http.StatusOK, gin.H{"decisions": []Decision{}})
		return
	}

	// Get decisions from database
	dbDecisions, err := h.db.GetDecisions(ctx, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get decisions"})
		return
	}

	// Debug log
	fmt.Printf("Found %d decisions in database\n", len(dbDecisions))

	// Convert to response format
	decisions := make([]Decision, len(dbDecisions))
	for i, d := range dbDecisions {
		decisions[i] = Decision{
			ID:         d.ID,
			Time:       d.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
			Action:     d.Action,
			Symbol:     d.Symbol,
			Reason:     d.Reason,
			Result:     d.Result,
			PnL:        d.PnL,
			TokensUsed: d.TokensUsed,
			LatencyMs:  d.LatencyMs,
		}
	}

	c.JSON(http.StatusOK, gin.H{"decisions": decisions})
}

// GetStats returns agent statistics.
func (h *AgentHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get decisions from database to calculate stats
	totalDecisions := 0
	totalTokens := 0

	if h.db != nil {
		decisions, err := h.db.GetDecisions(ctx, 1000)
		if err == nil {
			totalDecisions = len(decisions)
			for _, d := range decisions {
				totalTokens += d.TokensUsed
			}
		} else {
			fmt.Printf("Error getting decisions: %v\n", err)
		}
	} else {
		fmt.Println("Database is nil in GetStats")
	}

	c.JSON(http.StatusOK, gin.H{
		"today_decisions": totalDecisions,
		"today_trades":    0, // TODO: Calculate from trades table
		"today_pnl":       0, // TODO: Calculate from trades table
		"win_rate":        0, // TODO: Calculate from trades table
		"llm_calls":       totalDecisions,
		"tokens_used":     totalTokens,
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
		Action: action,
		Reason: resp.Content,
		Result: "已记录",
		PnL:    pnl,
	}
}

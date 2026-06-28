package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents the database connection.
type DB struct {
	pool *pgxpool.Pool
}

// Config holds database configuration.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// New creates a new database connection.
func New(cfg Config) (*DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{pool: pool}, nil
}

// Close closes the database connection.
func (db *DB) Close() {
	db.pool.Close()
}

// Trade represents a trade record.
type Trade struct {
	ID        int64     `json:"id"`
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Type      string    `json:"type"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Total     float64   `json:"total"`
	Fee       float64   `json:"fee"`
	PnL       float64   `json:"pnl"`
	Strategy  string    `json:"strategy"`
	Status    string    `json:"status"`
	OrderID   string    `json:"order_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Decision represents a decision record.
type Decision struct {
	ID          int64     `json:"id"`
	Action      string    `json:"action"`
	Symbol      string    `json:"symbol"`
	Reason      string    `json:"reason"`
	Result      string    `json:"result"`
	PnL         float64   `json:"pnl"`
	TokensUsed  int       `json:"tokens_used"`
	LatencyMs   int       `json:"latency_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

// ArbitrageOpportunity represents an arbitrage opportunity.
type ArbitrageOpportunity struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Path        string    `json:"path"`
	SpreadBps   float64   `json:"spread_bps"`
	ProfitUSDT  float64   `json:"profit_usdt"`
	Executed    bool      `json:"executed"`
	CreatedAt   time.Time `json:"created_at"`
}

// DailyStats represents daily statistics.
type DailyStats struct {
	ID                     int64     `json:"id"`
	Date                   time.Time `json:"date"`
	TotalBalance           float64   `json:"total_balance"`
	DailyPnL               float64   `json:"daily_pnl"`
	DailyPnLPct            float64   `json:"daily_pnl_pct"`
	TotalTrades            int       `json:"total_trades"`
	WinningTrades          int       `json:"winning_trades"`
	LosingTrades           int       `json:"losing_trades"`
	WinRate                float64   `json:"win_rate"`
	MaxDrawdown            float64   `json:"max_drawdown"`
	ArbitrageOpportunities int       `json:"arbitrage_opportunities"`
	ArbitrageProfit        float64   `json:"arbitrage_profit"`
	AgentDecisions         int       `json:"agent_decisions"`
	AgentPnL               float64   `json:"agent_pnl"`
	LLMCalls               int       `json:"llm_calls"`
	LLMTokens              int       `json:"llm_tokens"`
}

// EquitySnapshot represents an equity snapshot.
type EquitySnapshot struct {
	ID         int64     `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	TotalValue float64   `json:"total_value"`
}

// InsertTrade inserts a new trade record.
func (db *DB) InsertTrade(ctx context.Context, trade *Trade) error {
	query := `
		INSERT INTO trades (symbol, side, type, price, quantity, total, fee, pnl, strategy, status, order_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at`

	return db.pool.QueryRow(ctx, query,
		trade.Symbol, trade.Side, trade.Type, trade.Price, trade.Quantity,
		trade.Total, trade.Fee, trade.PnL, trade.Strategy, trade.Status, trade.OrderID,
	).Scan(&trade.ID, &trade.CreatedAt)
}

// GetTrades returns trades with optional filters.
func (db *DB) GetTrades(ctx context.Context, limit int, strategy string) ([]Trade, error) {
	query := `SELECT id, symbol, side, type, price, quantity, total, fee, pnl, strategy, status, order_id, created_at FROM trades`
	args := []any{}

	if strategy != "" {
		query += " WHERE strategy = $1"
		args = append(args, strategy)
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query trades: %w", err)
	}
	defer rows.Close()

	var trades []Trade
	for rows.Next() {
		var t Trade
		if err := rows.Scan(&t.ID, &t.Symbol, &t.Side, &t.Type, &t.Price, &t.Quantity,
			&t.Total, &t.Fee, &t.PnL, &t.Strategy, &t.Status, &t.OrderID, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan trade: %w", err)
		}
		trades = append(trades, t)
	}

	return trades, nil
}

// InsertDecision inserts a new decision record.
func (db *DB) InsertDecision(ctx context.Context, decision *Decision) error {
	query := `
		INSERT INTO decisions (action, symbol, reason, result, pnl, tokens_used, latency_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	return db.pool.QueryRow(ctx, query,
		decision.Action, decision.Symbol, decision.Reason, decision.Result,
		decision.PnL, decision.TokensUsed, decision.LatencyMs,
	).Scan(&decision.ID, &decision.CreatedAt)
}

// GetDecisions returns decisions with optional limit.
func (db *DB) GetDecisions(ctx context.Context, limit int) ([]Decision, error) {
	query := `SELECT id, action, symbol, reason, result, pnl, tokens_used, latency_ms, created_at
		FROM decisions ORDER BY created_at DESC LIMIT $1`

	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query decisions: %w", err)
	}
	defer rows.Close()

	var decisions []Decision
	for rows.Next() {
		var d Decision
		if err := rows.Scan(&d.ID, &d.Action, &d.Symbol, &d.Reason, &d.Result,
			&d.PnL, &d.TokensUsed, &d.LatencyMs, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan decision: %w", err)
		}
		decisions = append(decisions, d)
	}

	return decisions, nil
}

// InsertArbitrageOpportunity inserts a new arbitrage opportunity.
func (db *DB) InsertArbitrageOpportunity(ctx context.Context, opp *ArbitrageOpportunity) error {
	query := `
		INSERT INTO arbitrage_opportunities (type, path, spread_bps, profit_usdt, executed)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	return db.pool.QueryRow(ctx, query,
		opp.Type, opp.Path, opp.SpreadBps, opp.ProfitUSDT, opp.Executed,
	).Scan(&opp.ID, &opp.CreatedAt)
}

// GetArbitrageOpportunities returns arbitrage opportunities.
func (db *DB) GetArbitrageOpportunities(ctx context.Context, limit int) ([]ArbitrageOpportunity, error) {
	query := `SELECT id, type, path, spread_bps, profit_usdt, executed, created_at
		FROM arbitrage_opportunities ORDER BY created_at DESC LIMIT $1`

	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query arbitrage opportunities: %w", err)
	}
	defer rows.Close()

	var opps []ArbitrageOpportunity
	for rows.Next() {
		var o ArbitrageOpportunity
		if err := rows.Scan(&o.ID, &o.Type, &o.Path, &o.SpreadBps, &o.ProfitUSDT,
			&o.Executed, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan opportunity: %w", err)
		}
		opps = append(opps, o)
	}

	return opps, nil
}

// InsertEquitySnapshot inserts an equity snapshot.
func (db *DB) InsertEquitySnapshot(ctx context.Context, snapshot *EquitySnapshot) error {
	query := `
		INSERT INTO equity_snapshots (timestamp, total_value)
		VALUES ($1, $2)
		RETURNING id`

	return db.pool.QueryRow(ctx, query,
		snapshot.Timestamp, snapshot.TotalValue,
	).Scan(&snapshot.ID)
}

// GetEquitySnapshots returns equity snapshots for a time range.
func (db *DB) GetEquitySnapshots(ctx context.Context, from, to time.Time) ([]EquitySnapshot, error) {
	query := `SELECT id, timestamp, total_value
		FROM equity_snapshots
		WHERE timestamp BETWEEN $1 AND $2
		ORDER BY timestamp ASC`

	rows, err := db.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("query equity snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []EquitySnapshot
	for rows.Next() {
		var s EquitySnapshot
		if err := rows.Scan(&s.ID, &s.Timestamp, &s.TotalValue); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}

// InsertDailyStats inserts or updates daily statistics.
func (db *DB) InsertDailyStats(ctx context.Context, stats *DailyStats) error {
	query := `
		INSERT INTO daily_stats (date, total_balance, daily_pnl, daily_pnl_pct, total_trades,
			winning_trades, losing_trades, win_rate, max_drawdown, arbitrage_opportunities,
			arbitrage_profit, agent_decisions, agent_pnl, llm_calls, llm_tokens)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (date) DO UPDATE SET
			total_balance = EXCLUDED.total_balance,
			daily_pnl = EXCLUDED.daily_pnl,
			daily_pnl_pct = EXCLUDED.daily_pnl_pct,
			total_trades = EXCLUDED.total_trades,
			winning_trades = EXCLUDED.winning_trades,
			losing_trades = EXCLUDED.losing_trades,
			win_rate = EXCLUDED.win_rate,
			max_drawdown = EXCLUDED.max_drawdown,
			arbitrage_opportunities = EXCLUDED.arbitrage_opportunities,
			arbitrage_profit = EXCLUDED.arbitrage_profit,
			agent_decisions = EXCLUDED.agent_decisions,
			agent_pnl = EXCLUDED.agent_pnl,
			llm_calls = EXCLUDED.llm_calls,
			llm_tokens = EXCLUDED.llm_tokens
		RETURNING id`

	return db.pool.QueryRow(ctx, query,
		stats.Date, stats.TotalBalance, stats.DailyPnL, stats.DailyPnLPct,
		stats.TotalTrades, stats.WinningTrades, stats.LosingTrades, stats.WinRate,
		stats.MaxDrawdown, stats.ArbitrageOpportunities, stats.ArbitrageProfit,
		stats.AgentDecisions, stats.AgentPnL, stats.LLMCalls, stats.LLMTokens,
	).Scan(&stats.ID)
}

// GetDailyStats returns daily statistics for a date range.
func (db *DB) GetDailyStats(ctx context.Context, from, to time.Time) ([]DailyStats, error) {
	query := `SELECT id, date, total_balance, daily_pnl, daily_pnl_pct, total_trades,
		winning_trades, losing_trades, win_rate, max_drawdown, arbitrage_opportunities,
		arbitrage_profit, agent_decisions, agent_pnl, llm_calls, llm_tokens
		FROM daily_stats
		WHERE date BETWEEN $1 AND $2
		ORDER BY date ASC`

	rows, err := db.pool.Query(ctx, query, from, to)
	if err != nil {
		return nil, fmt.Errorf("query daily stats: %w", err)
	}
	defer rows.Close()

	var stats []DailyStats
	for rows.Next() {
		var s DailyStats
		if err := rows.Scan(&s.ID, &s.Date, &s.TotalBalance, &s.DailyPnL, &s.DailyPnLPct,
			&s.TotalTrades, &s.WinningTrades, &s.LosingTrades, &s.WinRate, &s.MaxDrawdown,
			&s.ArbitrageOpportunities, &s.ArbitrageProfit, &s.AgentDecisions, &s.AgentPnL,
			&s.LLMCalls, &s.LLMTokens); err != nil {
			return nil, fmt.Errorf("scan daily stats: %w", err)
		}
		stats = append(stats, s)
	}

	return stats, nil
}

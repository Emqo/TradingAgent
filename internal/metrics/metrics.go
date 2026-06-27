package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics.
type Metrics struct {
	// Trade metrics
	TradesTotal    *prometheus.CounterVec
	TradesSuccess  *prometheus.CounterVec
	TradesFailed   *prometheus.CounterVec
	TradePnL       *prometheus.CounterVec

	// LLM metrics
	LLMCallsTotal  *prometheus.CounterVec
	LLMTokensTotal *prometheus.CounterVec
	LLMLatency     *prometheus.HistogramVec

	// Arbitrage metrics
	ArbitrageOpportunities *prometheus.CounterVec
	ArbitrageSpread        *prometheus.HistogramVec

	// Risk metrics
	RiskChecksTotal *prometheus.CounterVec
	RiskAlertsTotal *prometheus.CounterVec

	// System metrics
	ActiveSessions prometheus.Gauge
	CacheHits      *prometheus.CounterVec
	CacheMisses    *prometheus.CounterVec
}

// NewMetrics creates new Prometheus metrics.
func NewMetrics() *Metrics {
	m := &Metrics{
		// Trade metrics
		TradesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_trades_total",
				Help: "Total number of trades executed",
			},
			[]string{"symbol", "side", "type"},
		),
		TradesSuccess: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_trades_success_total",
				Help: "Total number of successful trades",
			},
			[]string{"symbol", "side"},
		),
		TradesFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_trades_failed_total",
				Help: "Total number of failed trades",
			},
			[]string{"symbol", "side", "reason"},
		),
		TradePnL: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_pnl_total",
				Help: "Total profit/loss in USDT",
			},
			[]string{"symbol"},
		),

		// LLM metrics
		LLMCallsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_llm_calls_total",
				Help: "Total number of LLM calls",
			},
			[]string{"provider", "status"},
		),
		LLMTokensTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_llm_tokens_total",
				Help: "Total tokens used",
			},
			[]string{"provider", "type"},
		),
		LLMLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "trading_agent_llm_latency_seconds",
				Help:    "LLM call latency in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"provider"},
		),

		// Arbitrage metrics
		ArbitrageOpportunities: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_arbitrage_opportunities_total",
				Help: "Total arbitrage opportunities detected",
			},
			[]string{"type"},
		),
		ArbitrageSpread: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "trading_agent_arbitrage_spread_bps",
				Help:    "Arbitrage spread in basis points",
				Buckets: []float64{5, 10, 15, 20, 30, 50, 100},
			},
			[]string{"type"},
		),

		// Risk metrics
		RiskChecksTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_risk_checks_total",
				Help: "Total risk checks performed",
			},
			[]string{"result"},
		),
		RiskAlertsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_risk_alerts_total",
				Help: "Total risk alerts triggered",
			},
			[]string{"level", "type"},
		),

		// System metrics
		ActiveSessions: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "trading_agent_active_sessions",
				Help: "Number of active LLM sessions",
			},
		),
		CacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_cache_hits_total",
				Help: "Total cache hits",
			},
			[]string{"cache_type"},
		),
		CacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "trading_agent_cache_misses_total",
				Help: "Total cache misses",
			},
			[]string{"cache_type"},
		),
	}

	// Register all metrics
	prometheus.MustRegister(
		m.TradesTotal,
		m.TradesSuccess,
		m.TradesFailed,
		m.TradePnL,
		m.LLMCallsTotal,
		m.LLMTokensTotal,
		m.LLMLatency,
		m.ArbitrageOpportunities,
		m.ArbitrageSpread,
		m.RiskChecksTotal,
		m.RiskAlertsTotal,
		m.ActiveSessions,
		m.CacheHits,
		m.CacheMisses,
	)

	return m
}

// StartMetricsServer starts the Prometheus metrics HTTP server.
func StartMetricsServer(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}

// RecordTrade records a trade.
func (m *Metrics) RecordTrade(symbol, side, orderType string) {
	m.TradesTotal.WithLabelValues(symbol, side, orderType).Inc()
}

// RecordTradeSuccess records a successful trade.
func (m *Metrics) RecordTradeSuccess(symbol, side string) {
	m.TradesSuccess.WithLabelValues(symbol, side).Inc()
}

// RecordTradeFailed records a failed trade.
func (m *Metrics) RecordTradeFailed(symbol, side, reason string) {
	m.TradesFailed.WithLabelValues(symbol, side, reason).Inc()
}

// RecordPnL records profit/loss.
func (m *Metrics) RecordPnL(symbol string, pnl float64) {
	m.TradePnL.WithLabelValues(symbol).Add(pnl)
}

// RecordLLMCall records an LLM call.
func (m *Metrics) RecordLLMCall(provider, status string) {
	m.LLMCallsTotal.WithLabelValues(provider, status).Inc()
}

// RecordLLMTokens records LLM token usage.
func (m *Metrics) RecordLLMTokens(provider, tokenType string, count int) {
	m.LLMTokensTotal.WithLabelValues(provider, tokenType).Add(float64(count))
}

// RecordLLMLatency records LLM call latency.
func (m *Metrics) RecordLLMLatency(provider string, seconds float64) {
	m.LLMLatency.WithLabelValues(provider).Observe(seconds)
}

// RecordArbitrageOpportunity records an arbitrage opportunity.
func (m *Metrics) RecordArbitrageOpportunity(arbType string) {
	m.ArbitrageOpportunities.WithLabelValues(arbType).Inc()
}

// RecordArbitrageSpread records an arbitrage spread.
func (m *Metrics) RecordArbitrageSpread(arbType string, spreadBps float64) {
	m.ArbitrageSpread.WithLabelValues(arbType).Observe(spreadBps)
}

// RecordRiskCheck records a risk check.
func (m *Metrics) RecordRiskCheck(result string) {
	m.RiskChecksTotal.WithLabelValues(result).Inc()
}

// RecordRiskAlert records a risk alert.
func (m *Metrics) RecordRiskAlert(level, alertType string) {
	m.RiskAlertsTotal.WithLabelValues(level, alertType).Inc()
}

// SetActiveSessions sets the number of active sessions.
func (m *Metrics) SetActiveSessions(count int) {
	m.ActiveSessions.Set(float64(count))
}

// RecordCacheHit records a cache hit.
func (m *Metrics) RecordCacheHit(cacheType string) {
	m.CacheHits.WithLabelValues(cacheType).Inc()
}

// RecordCacheMiss records a cache miss.
func (m *Metrics) RecordCacheMiss(cacheType string) {
	m.CacheMisses.WithLabelValues(cacheType).Inc()
}

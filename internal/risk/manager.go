package risk

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Manager manages risk checks and position tracking.
type Manager struct {
	mu       sync.RWMutex
	config   Config
	state    State
	alerts   []Alert
	listeners []AlertListener
}

// Config holds risk management configuration.
type Config struct {
	// MaxPositionUSDT is the maximum position size in USDT.
	MaxPositionUSDT float64 `yaml:"max_position_usdt"`
	// MaxTotalPositionUSDT is the maximum total position size in USDT.
	MaxTotalPositionUSDT float64 `yaml:"max_total_position_usdt"`
	// MaxDailyLossUSDT is the maximum daily loss in USDT.
	MaxDailyLossUSDT float64 `yaml:"max_daily_loss_usdt"`
	// MaxDailyLossPercent is the maximum daily loss as a percentage.
	MaxDailyLossPercent float64 `yaml:"max_daily_loss_percent"`
	// MaxDrawdownPercent is the maximum drawdown percentage.
	MaxDrawdownPercent float64 `yaml:"max_drawdown_percent"`
	// MaxLeverage is the maximum leverage allowed.
	MaxLeverage float64 `yaml:"max_leverage"`
	// MaxExposurePerPairUSDT is the maximum exposure per trading pair.
	MaxExposurePerPairUSDT float64 `yaml:"max_exposure_per_pair_usdt"`
	// CooldownAfterLoss is the cooldown period after a loss.
	CooldownAfterLoss time.Duration `yaml:"cooldown_after_loss"`
}

// State tracks the current risk state.
type State struct {
	// TotalPositionUSDT is the current total position size.
	TotalPositionUSDT float64 `json:"total_position_usdt"`
	// DailyPnL is the daily profit/loss.
	DailyPnL float64 `json:"daily_pnl"`
	// DailyPnLPercent is the daily PnL as a percentage.
	DailyPnLPercent float64 `json:"daily_pnl_percent"`
	// PeakValue is the peak portfolio value.
	PeakValue float64 `json:"peak_value"`
	// CurrentValue is the current portfolio value.
	CurrentValue float64 `json:"current_value"`
	// DrawdownPercent is the current drawdown percentage.
	DrawdownPercent float64 `json:"drawdown_percent"`
	// LastLossTime is the time of the last loss.
	LastLossTime time.Time `json:"last_loss_time"`
	// Positions tracks positions by symbol.
	Positions map[string]Position `json:"positions"`
	// DailyResetTime is when the daily counters were last reset.
	DailyResetTime time.Time `json:"daily_reset_time"`
	// IsPaused indicates if trading is paused.
	IsPaused bool `json:"is_paused"`
	// PauseReason is the reason for pausing.
	PauseReason string `json:"pause_reason"`
}

// Position represents a trading position.
type Position struct {
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"` // "LONG" or "SHORT"
	Size      float64   `json:"size"`
	EntryPrice float64  `json:"entry_price"`
	CurrentPrice float64 `json:"current_price"`
	USDTValue float64   `json:"usdt_value"`
	PnL       float64   `json:"pnl"`
	PnLPercent float64  `json:"pnl_percent"`
	Timestamp time.Time `json:"timestamp"`
}

// Alert represents a risk alert.
type Alert struct {
	Level     AlertLevel `json:"level"`
	Type      AlertType  `json:"type"`
	Message   string     `json:"message"`
	Timestamp time.Time  `json:"timestamp"`
	Data      any        `json:"data,omitempty"`
}

// AlertLevel represents the severity of an alert.
type AlertLevel string

const (
	AlertLevelInfo    AlertLevel = "INFO"
	AlertLevelWarning AlertLevel = "WARNING"
	AlertLevelError   AlertLevel = "ERROR"
	AlertLevelCritical AlertLevel = "CRITICAL"
)

// AlertType represents the type of alert.
type AlertType string

const (
	AlertTypePositionLimit    AlertType = "POSITION_LIMIT"
	AlertTypeDailyLoss        AlertType = "DAILY_LOSS"
	AlertTypeDrawdown         AlertType = "DRAWDOWN"
	AlertTypeLeverage         AlertType = "LEVERAGE"
	AlertTypeExposure         AlertType = "EXPOSURE"
	AlertTypeCooldown         AlertType = "COOLDOWN"
	AlertTypeTradingPaused    AlertType = "TRADING_PAUSED"
	AlertTypeTradingResumed   AlertType = "TRADING_RESUMED"
)

// AlertListener is a function that handles alerts.
type AlertListener func(alert Alert)

// NewManager creates a new risk manager.
func NewManager(config Config) *Manager {
	return &Manager{
		config: config,
		state: State{
			Positions:     make(map[string]Position),
			DailyResetTime: time.Now(),
		},
	}
}

// AddListener adds an alert listener.
func (m *Manager) AddListener(listener AlertListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// CheckTrade checks if a trade is allowed.
func (m *Manager) CheckTrade(symbol string, side string, sizeUSDT float64, leverage float64) (*TradeCheckResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset daily counters if needed
	m.resetDailyIfNeeded()

	result := &TradeCheckResult{
		Allowed: true,
		Checks:  []CheckResult{},
	}

	// Check if trading is paused
	if m.state.IsPaused {
		result.Allowed = false
		result.Reason = m.state.PauseReason
		result.Checks = append(result.Checks, CheckResult{
			Name:    "trading_paused",
			Passed:  false,
			Message: m.state.PauseReason,
		})
		return result, nil
	}

	// Check cooldown after loss
	if !m.state.LastLossTime.IsZero() && m.config.CooldownAfterLoss > 0 {
		elapsed := time.Since(m.state.LastLossTime)
		if elapsed < m.config.CooldownAfterLoss {
			result.Allowed = false
			result.Reason = fmt.Sprintf("cooldown period active (%v remaining)", m.config.CooldownAfterLoss-elapsed)
			result.Checks = append(result.Checks, CheckResult{
				Name:    "cooldown",
				Passed:  false,
				Message: result.Reason,
			})
			return result, nil
		}
	}

	// Check single position limit
	if sizeUSDT > m.config.MaxPositionUSDT {
		result.Allowed = false
		result.Reason = fmt.Sprintf("position size $%.2f exceeds limit $%.2f", sizeUSDT, m.config.MaxPositionUSDT)
		result.Checks = append(result.Checks, CheckResult{
			Name:    "position_limit",
			Passed:  false,
			Message: result.Reason,
		})
		m.emitAlert(Alert{
			Level:     AlertLevelWarning,
			Type:      AlertTypePositionLimit,
			Message:   result.Reason,
			Timestamp: time.Now(),
		})
	} else {
		result.Checks = append(result.Checks, CheckResult{
			Name:   "position_limit",
			Passed: true,
		})
	}

	// Check total position limit
	newTotal := m.state.TotalPositionUSDT + sizeUSDT
	if newTotal > m.config.MaxTotalPositionUSDT {
		result.Allowed = false
		result.Reason = fmt.Sprintf("total position $%.2f would exceed limit $%.2f", newTotal, m.config.MaxTotalPositionUSDT)
		result.Checks = append(result.Checks, CheckResult{
			Name:    "total_position",
			Passed:  false,
			Message: result.Reason,
		})
	} else {
		result.Checks = append(result.Checks, CheckResult{
			Name:   "total_position",
			Passed: true,
		})
	}

	// Check leverage
	if leverage > m.config.MaxLeverage {
		result.Allowed = false
		result.Reason = fmt.Sprintf("leverage %.1fx exceeds limit %.1fx", leverage, m.config.MaxLeverage)
		result.Checks = append(result.Checks, CheckResult{
			Name:    "leverage",
			Passed:  false,
			Message: result.Reason,
		})
		m.emitAlert(Alert{
			Level:     AlertLevelWarning,
			Type:      AlertTypeLeverage,
			Message:   result.Reason,
			Timestamp: time.Now(),
		})
	} else {
		result.Checks = append(result.Checks, CheckResult{
			Name:   "leverage",
			Passed: true,
		})
	}

	// Check exposure per pair
	currentExposure := m.getExposure(symbol)
	newExposure := currentExposure + sizeUSDT
	if newExposure > m.config.MaxExposurePerPairUSDT {
		result.Allowed = false
		result.Reason = fmt.Sprintf("exposure for %s $%.2f would exceed limit $%.2f", symbol, newExposure, m.config.MaxExposurePerPairUSDT)
		result.Checks = append(result.Checks, CheckResult{
			Name:    "exposure",
			Passed:  false,
			Message: result.Reason,
		})
		m.emitAlert(Alert{
			Level:     AlertLevelWarning,
			Type:      AlertTypeExposure,
			Message:   result.Reason,
			Timestamp: time.Now(),
		})
	} else {
		result.Checks = append(result.Checks, CheckResult{
			Name:   "exposure",
			Passed: true,
		})
	}

	// Check daily loss limit
	if m.state.DailyPnL < 0 {
		absLoss := -m.state.DailyPnL
		if absLoss >= m.config.MaxDailyLossUSDT {
			result.Allowed = false
			result.Reason = fmt.Sprintf("daily loss $%.2f exceeds limit $%.2f", absLoss, m.config.MaxDailyLossUSDT)
			result.Checks = append(result.Checks, CheckResult{
				Name:    "daily_loss",
				Passed:  false,
				Message: result.Reason,
			})
			m.emitAlert(Alert{
				Level:     AlertLevelError,
				Type:      AlertTypeDailyLoss,
				Message:   result.Reason,
				Timestamp: time.Now(),
			})
		} else {
			result.Checks = append(result.Checks, CheckResult{
				Name:   "daily_loss",
				Passed: true,
			})
		}
	}

	// Check drawdown
	if m.state.DrawdownPercent >= m.config.MaxDrawdownPercent {
		result.Allowed = false
		result.Reason = fmt.Sprintf("drawdown %.2f%% exceeds limit %.2f%%", m.state.DrawdownPercent, m.config.MaxDrawdownPercent)
		result.Checks = append(result.Checks, CheckResult{
			Name:    "drawdown",
			Passed:  false,
			Message: result.Reason,
		})
		m.emitAlert(Alert{
			Level:     AlertLevelCritical,
			Type:      AlertTypeDrawdown,
			Message:   result.Reason,
			Timestamp: time.Now(),
		})
	} else {
		result.Checks = append(result.Checks, CheckResult{
			Name:   "drawdown",
			Passed: true,
		})
	}

	return result, nil
}

// UpdatePosition updates a position in the risk manager.
func (m *Manager) UpdatePosition(symbol string, side string, size float64, entryPrice float64, currentPrice float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	usdtValue := size * currentPrice
	pnl := 0.0
	pnlPercent := 0.0

	if side == "LONG" {
		pnl = (currentPrice - entryPrice) * size
		pnlPercent = (currentPrice - entryPrice) / entryPrice * 100
	} else {
		pnl = (entryPrice - currentPrice) * size
		pnlPercent = (entryPrice - currentPrice) / entryPrice * 100
	}

	m.state.Positions[symbol] = Position{
		Symbol:        symbol,
		Side:          side,
		Size:          size,
		EntryPrice:    entryPrice,
		CurrentPrice:  currentPrice,
		USDTValue:     usdtValue,
		PnL:           pnl,
		PnLPercent:    pnlPercent,
		Timestamp:     time.Now(),
	}

	// Update totals
	m.recalculateState()
}

// ClosePosition closes a position.
func (m *Manager) ClosePosition(symbol string, pnl float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.state.Positions, symbol)

	// Update daily PnL
	m.state.DailyPnL += pnl

	// Update last loss time if this was a loss
	if pnl < 0 {
		m.state.LastLossTime = time.Now()
	}

	// Recalculate
	m.recalculateState()
}

// GetState returns the current risk state.
func (m *Manager) GetState() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// PauseTrading pauses trading.
func (m *Manager) PauseTrading(reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.IsPaused = true
	m.state.PauseReason = reason

	m.emitAlert(Alert{
		Level:     AlertLevelWarning,
		Type:      AlertTypeTradingPaused,
		Message:   fmt.Sprintf("Trading paused: %s", reason),
		Timestamp: time.Now(),
	})
}

// ResumeTrading resumes trading.
func (m *Manager) ResumeTrading() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.IsPaused = false
	m.state.PauseReason = ""

	m.emitAlert(Alert{
		Level:     AlertLevelInfo,
		Type:      AlertTypeTradingResumed,
		Message:   "Trading resumed",
		Timestamp: time.Now(),
	})
}

// GetAlerts returns recent alerts.
func (m *Manager) GetAlerts(limit int) []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.alerts) {
		limit = len(m.alerts)
	}

	start := len(m.alerts) - limit
	if start < 0 {
		start = 0
	}

	return m.alerts[start:]
}

// Private methods

func (m *Manager) recalculateState() {
	totalUSDT := 0.0
	for _, pos := range m.state.Positions {
		totalUSDT += pos.USDTValue
	}
	m.state.TotalPositionUSDT = totalUSDT

	// Update current value and drawdown
	m.state.CurrentValue = totalUSDT
	if m.state.CurrentValue > m.state.PeakValue {
		m.state.PeakValue = m.state.CurrentValue
	}
	if m.state.PeakValue > 0 {
		m.state.DrawdownPercent = (m.state.PeakValue - m.state.CurrentValue) / m.state.PeakValue * 100
	}
}

func (m *Manager) getExposure(symbol string) float64 {
	if pos, ok := m.state.Positions[symbol]; ok {
		return pos.USDTValue
	}
	return 0
}

func (m *Manager) resetDailyIfNeeded() {
	now := time.Now()
	if now.Sub(m.state.DailyResetTime) >= 24*time.Hour {
		m.state.DailyPnL = 0
		m.state.DailyPnLPercent = 0
		m.state.DailyResetTime = now
		log.Println("📊 Daily risk counters reset")
	}
}

func (m *Manager) emitAlert(alert Alert) {
	m.alerts = append(m.alerts, alert)

	// Keep only last 100 alerts
	if len(m.alerts) > 100 {
		m.alerts = m.alerts[len(m.alerts)-100:]
	}

	// Notify listeners
	for _, listener := range m.listeners {
		listener(alert)
	}
}

// TradeCheckResult represents the result of a trade check.
type TradeCheckResult struct {
	Allowed bool          `json:"allowed"`
	Reason  string        `json:"reason,omitempty"`
	Checks  []CheckResult `json:"checks"`
}

// CheckResult represents a single check result.
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}

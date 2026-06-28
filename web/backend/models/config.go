package models

import "time"

// Config represents the trading agent configuration.
type Config struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RiskConfig represents risk management configuration.
type RiskConfig struct {
	MaxPositionUSDT    float64 `json:"max_position_usdt"`
	MaxDailyLossUSDT   float64 `json:"max_daily_loss_usdt"`
	MaxDrawdownPct     float64 `json:"max_drawdown_pct"`
	MaxLeverage        float64 `json:"max_leverage"`
	CooldownAfterLoss  int     `json:"cooldown_after_loss_minutes"`
}

// ArbitrageConfig represents arbitrage configuration.
type ArbitrageConfig struct {
	EnableTriangular    bool    `json:"enable_triangular"`
	EnableCashAndCarry  bool    `json:"enable_cash_and_carry"`
	MinSpreadBps        float64 `json:"min_spread_bps"`
	MaxPositionUSDT     float64 `json:"max_position_usdt"`
}

// NotificationConfig represents notification configuration.
type NotificationConfig struct {
	EnableTelegram bool   `json:"enable_telegram"`
	TelegramToken  string `json:"telegram_token,omitempty"`
	TelegramChatID string `json:"telegram_chat_id,omitempty"`
	EnableEmail    bool   `json:"enable_email"`
	EmailSMTP      string `json:"email_smtp,omitempty"`
	EmailFrom      string `json:"email_from,omitempty"`
	EmailTo        string `json:"email_to,omitempty"`
}

// LLMConfig represents LLM configuration.
type LLMConfig struct {
	Provider    string  `json:"provider"`
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key,omitempty"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

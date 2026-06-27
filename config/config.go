package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LLM    LLMConfig    `yaml:"llm"`
	Binance BinanceConfig `yaml:"binance"`
	Agent  AgentConfig  `yaml:"agent"`
	Risk   RiskConfig   `yaml:"risk"`
}

type LLMConfig struct {
	Default   string                  `yaml:"default"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
}

type BinanceConfig struct {
	Testnet   bool   `yaml:"testnet"`
	APIKey    string `yaml:"api_key"`
	APISecret string `yaml:"api_secret"`
}

type AgentConfig struct {
	Interval    string  `yaml:"interval"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
}

type RiskConfig struct {
	MaxPositionUSDT  float64 `yaml:"max_position_usdt"`
	MaxDailyLossUSDT float64 `yaml:"max_daily_loss_usdt"`
	MaxDrawdownPct   float64 `yaml:"max_drawdown_pct"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}

func (c *LLMConfig) GetProvider() (ProviderConfig, error) {
	provider, ok := c.Providers[c.Default]
	if !ok {
		return ProviderConfig{}, fmt.Errorf("provider %q not found", c.Default)
	}
	return provider, nil
}

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Emqo/TradingAgent/config"
	"github.com/Emqo/TradingAgent/internal/exchange"
	"github.com/Emqo/TradingAgent/internal/llm"
)

func main() {
	// Load config
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get LLM provider config
	providerCfg, err := cfg.LLM.GetProvider()
	if err != nil {
		log.Fatalf("Failed to get provider config: %v", err)
	}

	// Create LLM provider
	llmProvider := llm.NewClaudeProvider(
		providerCfg.BaseURL,
		providerCfg.APIKey,
		providerCfg.Model,
	)

	fmt.Printf("✅ LLM Provider: %s (model: %s)\n", llmProvider.Name(), providerCfg.Model)

	// Create Binance exchange
	exchangeProvider := exchange.NewBinanceExchange(
		cfg.Binance.APIKey,
		cfg.Binance.APISecret,
		cfg.Binance.Testnet,
	)

	fmt.Printf("✅ Exchange: %s (testnet: %v)\n", exchangeProvider.Name(), cfg.Binance.Testnet)
	fmt.Println("---")

	// Test Binance connection - get BTC price
	ctx := context.Background()
	ticker, err := exchangeProvider.GetTicker(ctx, "BTCUSDT")
	if err != nil {
		log.Fatalf("Failed to get ticker: %v", err)
	}

	fmt.Printf("📈 BTC/USDT Price: $%.2f\n", ticker.LastPrice)

	// Test LLM with market data
	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are a crypto trading analyst. Analyze the market data and provide a brief assessment.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Current BTC price: $%.2f. Give a one-sentence market assessment.", ticker.LastPrice),
		},
	}

	resp, err := llmProvider.Chat(ctx, messages, llm.WithMaxTokens(200))
	if err != nil {
		log.Fatalf("LLM call failed: %v", err)
	}

	fmt.Println("---")
	fmt.Printf("🤖 LLM Analysis: %s\n", resp.Content)
	fmt.Printf("📊 Token Usage: %d tokens\n", resp.TokenUsage.TotalTokens)
}
